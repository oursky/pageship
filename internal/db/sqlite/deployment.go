package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/oursky/pageship/internal/models"
)

func (c Conn) CreateDeployment(ctx context.Context, deployment *models.Deployment) error {
	_, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO deployment (id, created_at, updated_at, deleted_at, app_id, site_id, storage_key_prefix, metadata, uploaded_at)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :app_id, :site_id, :storage_key_prefix, :metadata, :uploaded_at)
	`, deployment)
	if err != nil {
		return err
	}

	return nil
}

func (c Conn) GetDeployment(ctx context.Context, appID string, siteName string, id string) (*models.Deployment, error) {
	var aggregate struct {
		*models.Deployment
		SiteDeploymentID *string `db:"site_deployment_id"`
	}

	err := c.tx.GetContext(ctx, &aggregate, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.app_id, d.site_id, d.storage_key_prefix, d.metadata, d.uploaded_at, sd.deployment_id AS site_deployment_id FROM deployment d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			JOIN site s ON (s.id = d.site_id AND s.deleted_at IS NULL)
			LEFT JOIN site_deployment sd USING (site_id)
			WHERE d.app_id = ? AND s.name = ? AND d.id = ? AND d.deleted_at IS NULL
	`, appID, siteName, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDeploymentNotFound
	} else if err != nil {
		return nil, err
	}

	aggregate.Deployment.SetStatus(aggregate.SiteDeploymentID)
	return aggregate.Deployment, nil
}

func (c Conn) MarkDeploymentUploaded(ctx context.Context, now time.Time, deployment *models.Deployment) error {
	err := c.tx.GetContext(ctx, deployment, `
		UPDATE deployment SET uploaded_at = ?
			WHERE id = ? AND deleted_at IS NULL AND uploaded_at IS NULL
			RETURNING id, created_at, updated_at, deleted_at, app_id, site_id, storage_key_prefix, metadata, uploaded_at
	`, now, deployment.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrDeploymentNotFound
	} else if err != nil {
		return err
	}

	deployment.Status = models.DeploymentStatusInactive

	return nil
}

func (c Conn) ActivateSiteDeployment(ctx context.Context, deployment *models.Deployment) error {
	_, err := c.tx.ExecContext(ctx, `
		INSERT INTO site_deployment (site_id, deployment_id)
			VALUES (?, ?)
			ON CONFLICT SET deployment_id = excluded.deployment_id
	`, deployment.SiteID, deployment.ID)
	if err != nil {
		return err
	}

	deployment.SetStatus(&deployment.ID)
	return nil
}

func (c Conn) DeactivateSiteDeployment(ctx context.Context, deployment *models.Deployment) error {
	result, err := c.tx.ExecContext(ctx, `
		DELETE FROM site_deployment (site_id, deployment_id)
			WHERE site_id = ? AND deployment_id = ?
	`, deployment.SiteID, deployment.ID)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n != 0 {
		deployment.SetStatus(nil)
	}
	return nil
}

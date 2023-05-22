package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/oursky/pageship/internal/models"
)

func (c Conn) CreateDeployment(ctx context.Context, deployment *models.Deployment) error {
	result, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO deployment (id, created_at, updated_at, deleted_at, name, app_id, storage_key_prefix, metadata, uploaded_at)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :name, :app_id, :storage_key_prefix, :metadata, :uploaded_at)
			ON CONFLICT (app_id, name) WHERE deleted_at IS NULL DO NOTHING
	`, deployment)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return models.ErrDeploymentUsedName
	}

	return nil
}

func (c Conn) GetDeployment(ctx context.Context, appID string, id string) (*models.Deployment, error) {
	var deployment models.Deployment

	err := c.tx.GetContext(ctx, &deployment, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.name, d.app_id, d.storage_key_prefix, d.metadata, d.uploaded_at FROM deployment d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			WHERE d.app_id = ? AND d.id = ? AND d.deleted_at IS NULL
	`, appID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDeploymentNotFound
	} else if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (c Conn) GetDeploymentByName(ctx context.Context, appID string, name string) (*models.Deployment, error) {
	var deployment models.Deployment

	err := c.tx.GetContext(ctx, &deployment, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.name, d.app_id, d.storage_key_prefix, d.metadata, d.uploaded_at FROM deployment d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			WHERE d.app_id = ? AND d.name = ? AND d.deleted_at IS NULL
	`, appID, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDeploymentNotFound
	} else if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (c Conn) ListDeployments(ctx context.Context, appID string) ([]*models.Deployment, error) {
	var deployments []*models.Deployment
	err := c.tx.SelectContext(ctx, &deployments, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.name, d.app_id, d.storage_key_prefix, d.metadata, d.uploaded_at FROM deployment d
			WHERE d.app_id = ? AND d.deleted_at IS NULL
			ORDER BY d.app_id, d.created_at
	`, appID)
	if err != nil {
		return nil, err
	}

	return deployments, nil
}

func (c Conn) MarkDeploymentUploaded(ctx context.Context, now time.Time, deployment *models.Deployment) error {
	err := c.tx.GetContext(ctx, deployment, `
		UPDATE deployment SET uploaded_at = ?
			WHERE id = ? AND deleted_at IS NULL AND uploaded_at IS NULL
			RETURNING id, created_at, updated_at, deleted_at, name, app_id, storage_key_prefix, metadata, uploaded_at
	`, now, deployment.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrDeploymentNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (c Conn) AssignDeploymentSite(ctx context.Context, deployment *models.Deployment, siteID string) error {
	_, err := c.tx.ExecContext(ctx, `
		INSERT INTO site_deployment (deployment_id, site_id)
			VALUES (?, ?)
			ON CONFLICT (site_id) DO UPDATE SET deployment_id = excluded.deployment_id
	`, deployment.ID, siteID)
	if err != nil {
		return err
	}

	return nil
}

func (c Conn) UnassignDeploymentSite(ctx context.Context, deployment *models.Deployment, siteID string) error {
	_, err := c.tx.ExecContext(ctx, `
		DELETE FROM site_deployment WHERE deployment_id = ? AND site_id = ?
	`, deployment.ID, siteID)
	if err != nil {
		return err
	}

	return nil
}

func (c Conn) GetSiteDeployment(ctx context.Context, appID string, siteName string) (*models.Deployment, error) {
	var deployment models.Deployment

	err := c.tx.GetContext(ctx, &deployment, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.name, d.app_id, d.storage_key_prefix, d.metadata, d.uploaded_at FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			JOIN site_deployment sd ON (sd.site_id = s.id)
			JOIN deployment d ON (d.id = sd.deployment_id AND d.deleted_at IS NULL)
			WHERE d.app_id = ? AND s.name = ? AND d.deleted_at IS NULL
	`, appID, siteName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDeploymentNotFound
	} else if err != nil {
		return nil, err
	}

	return &deployment, nil
}

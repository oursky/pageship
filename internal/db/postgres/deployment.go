package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

func (c Conn) CreateDeployment(ctx context.Context, deployment *models.Deployment) error {
	result, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO deployment (id, created_at, updated_at, deleted_at, name, app_id, storage_key_prefix, metadata, uploaded_at, expire_at)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :name, :app_id, :storage_key_prefix, :metadata, :uploaded_at, :expire_at)
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
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.name, d.app_id, d.storage_key_prefix, d.metadata, d.uploaded_at, d.expire_at FROM deployment d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			WHERE d.app_id = $1 AND d.id = $2 AND d.deleted_at IS NULL
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
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.name, d.app_id, d.storage_key_prefix, d.metadata, d.uploaded_at, d.expire_at FROM deployment d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			WHERE d.app_id = $1 AND d.name = $2 AND d.deleted_at IS NULL
	`, appID, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDeploymentNotFound
	} else if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (c Conn) ListDeployments(ctx context.Context, appID string) ([]db.DeploymentInfo, error) {
	var deployments []db.DeploymentInfo
	err := c.tx.SelectContext(ctx, &deployments, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.name, d.app_id, d.storage_key_prefix, d.metadata, d.uploaded_at, d.expire_at, min(s.name) as site_name FROM deployment d
			LEFT JOIN site s ON (s.deployment_id = d.id AND s.deleted_at IS NULL)
			WHERE d.app_id = $1 AND d.deleted_at IS NULL
			GROUP BY d.id
			ORDER BY d.app_id, d.created_at
	`, appID)
	if err != nil {
		return nil, err
	}

	return deployments, nil
}

func (c Conn) MarkDeploymentUploaded(ctx context.Context, now time.Time, deployment *models.Deployment) error {
	err := c.tx.GetContext(ctx, deployment, `
		UPDATE deployment SET uploaded_at = $1
			WHERE id = $2 AND deleted_at IS NULL AND uploaded_at IS NULL
			RETURNING id, created_at, updated_at, deleted_at, name, app_id, storage_key_prefix, metadata, uploaded_at
	`, now, deployment.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrDeploymentNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (c Conn) GetSiteDeployment(ctx context.Context, appID string, siteName string) (*models.Deployment, error) {
	var deployment models.Deployment

	err := c.tx.GetContext(ctx, &deployment, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.name, d.app_id, d.storage_key_prefix, d.metadata, d.uploaded_at, d.expire_at FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			JOIN deployment d ON (d.id = s.deployment_id AND d.deleted_at IS NULL)
			WHERE d.app_id = $1 AND s.name = $2 AND s.deleted_at IS NULL
	`, appID, siteName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDeploymentNotFound
	} else if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (c Conn) GetDeploymentSiteNames(ctx context.Context, deployment *models.Deployment) ([]string, error) {
	var names []string
	err := c.tx.SelectContext(ctx, &names, `
		SELECT s.name FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			JOIN deployment d ON (d.id = s.deployment_id AND d.deleted_at IS NULL)
			WHERE d.app_id = $1 AND d.id = $2 AND s.deleted_at IS NULL
			ORDER BY s.name
	`, deployment.AppID, deployment.ID)
	if err != nil {
		return nil, err
	}

	return names, nil
}

func (c Conn) SetDeploymentExpiry(ctx context.Context, deployment *models.Deployment) error {
	_, err := c.tx.ExecContext(ctx, `
		UPDATE deployment SET expire_at = $1 WHERE id = $2
	`, deployment.ExpireAt, deployment.ID)
	if err != nil {
		return err
	}

	return nil
}

func (c Conn) DeleteExpiredDeployments(ctx context.Context, now time.Time, expireBefore time.Time) (int64, error) {
	result, err := c.tx.ExecContext(ctx, `
		UPDATE deployment SET deleted_at = $1 WHERE deleted_at IS NULL AND expire_at < $2
	`, now, expireBefore)
	if err != nil {
		return 0, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return n, nil
}

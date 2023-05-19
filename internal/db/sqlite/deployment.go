package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/oursky/pageship/internal/models"
)

func (c Conn) CreateDeployment(ctx context.Context, deployment *models.Deployment) error {
	_, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO deployment (id, created_at, updated_at, deleted_at, app_id, site_id, status, storage_key_prefix, metadata)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :app_id, :site_id, :status, :storage_key_prefix, :metadata)
	`, deployment)
	if err != nil {
		return err
	}

	return nil
}

func (c Conn) GetDeployment(ctx context.Context, appID string, siteName string, id string) (*models.Deployment, error) {
	var deployment models.Deployment
	err := c.tx.GetContext(ctx, &deployment, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.app_id, d.site_id, d.status, d.storage_key_prefix, d.metadata FROM deployment d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			JOIN site s ON (s.id = d.site_id AND s.deleted_at IS NULL)
			WHERE d.app_id = ? AND s.name = ? AND d.id = ? AND d.deleted_at IS NULL
	`, appID, siteName, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDeploymentNotFound
	} else if err != nil {
		return nil, err
	}

	return &deployment, nil
}

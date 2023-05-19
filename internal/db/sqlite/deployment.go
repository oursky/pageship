package sqlite

import (
	"context"

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

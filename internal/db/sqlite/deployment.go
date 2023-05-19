package sqlite

import (
	"context"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
)

func (c Conn) CreateDeployment(
	ctx context.Context,
	appID string,
	siteID string,
	storageKeyPrefix string,
	files []models.FileEntry,
	config *config.SiteConfig,
) (*models.Deployment, error) {
	metadata := &models.DeploymentMetadata{
		Files:  files,
		Config: *config,
	}
	deployment := models.NewDeployment(c.clock.Now().UTC(), appID, siteID, storageKeyPrefix, metadata)

	_, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO deployment (id, created_at, updated_at, deleted_at, app_id, site_id, status, storage_key_prefix, metadata)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :app_id, :site_id, :status, :storage_key_prefix, :metadata)
	`, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

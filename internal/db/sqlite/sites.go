package sqlite

import (
	"context"

	"github.com/oursky/pageship/internal/models"
)

func (c Conn) CreateSiteIfNotExist(ctx context.Context, site *models.Site) error {
	_, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO site (id, app_id, name, created_at, updated_at, deleted_at)
			VALUES (:id, :app_id, :name, :created_at, :updated_at, :deleted_at)
			ON CONFLICT DO NOTHING
	`, site)
	if err != nil {
		return err
	}

	err = c.tx.GetContext(ctx, site, `
		SELECT id, app_id, name, created_at, updated_at, deleted_at FROM site
			WHERE app_id = ? AND name = ? AND deleted_at IS NULL
	`, site.AppID, site.Name)
	if err != nil {
		return err
	}

	return nil
}

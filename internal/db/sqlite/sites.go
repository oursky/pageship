package sqlite

import (
	"context"

	"github.com/oursky/pageship/internal/models"
)

func (c Conn) EnsureSite(ctx context.Context, appID string, siteName string) (*models.Site, error) {
	site := models.NewSite(appID, siteName, c.clock.Now().UTC())

	_, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO site (id, app_id, name, created_at, updated_at, deleted_at)
			VALUES (:id, :app_id, :name, :created_at, :updated_at, :deleted_at)
			ON CONFLICT DO NOTHING
	`, site)
	if err != nil {
		return nil, err
	}

	err = c.tx.GetContext(ctx, site, `
		SELECT id, app_id, name, created_at, updated_at, deleted_at FROM site
			WHERE app_id = ? AND name = ? AND deleted_at IS NULL
	`, appID, siteName)
	if err != nil {
		return nil, err
	}

	return site, nil
}

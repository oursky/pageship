package sqlite

import (
	"context"

	"github.com/oursky/pageship/internal/db"
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

func (c Conn) ListSites(ctx context.Context, appID string) ([]db.SiteInfo, error) {
	var info []db.SiteInfo
	err := c.tx.SelectContext(ctx, &info, `
		SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at, sd.deployment_id, d.created_at AS last_deployed_at FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			LEFT JOIN site_deployment sd ON (sd.site_id = s.id)
			LEFT JOIN deployment d ON (d.id = sd.deployment_id AND d.deleted_at IS NULL)
			WHERE s.app_id = ? AND s.deleted_at IS NULL
	`, appID)
	if err != nil {
		return nil, err
	}

	return info, nil
}

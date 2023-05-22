package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

func (c Conn) CreateSiteIfNotExist(ctx context.Context, site *models.Site) (*db.SiteInfo, error) {
	_, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO site (id, app_id, name, created_at, updated_at, deleted_at)
			VALUES (:id, :app_id, :name, :created_at, :updated_at, :deleted_at)
			ON CONFLICT (app_id, name) WHERE deleted_at IS NULL DO NOTHING
	`, site)
	if err != nil {
		return nil, err
	}

	var info db.SiteInfo
	err = c.tx.GetContext(ctx, &info, `
		SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at, sd.deployment_id, d.name AS deployment_name FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			LEFT JOIN site_deployment sd ON (sd.site_id = s.id)
			LEFT JOIN deployment d ON (d.id = sd.deployment_id AND d.deleted_at IS NULL)
			WHERE s.app_id = ? AND s.name = ? AND s.deleted_at IS NULL
	`, site.AppID, site.Name)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (c Conn) GetSiteByName(ctx context.Context, appID string, name string) (*models.Site, error) {
	var site models.Site
	err := c.tx.GetContext(ctx, &site, `
	SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at FROM site s
		JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
		WHERE s.app_id = ? AND s.name = ? AND s.deleted_at IS NULL
	`, appID, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrSiteNotFound
	} else if err != nil {
		return nil, err
	}

	return &site, nil
}

func (c Conn) GetSiteInfo(ctx context.Context, appID string, siteID string) (*db.SiteInfo, error) {
	var info db.SiteInfo
	err := c.tx.GetContext(ctx, &info, `
		SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at, sd.deployment_id, d.name AS deployment_name FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			LEFT JOIN site_deployment sd ON (sd.site_id = s.id)
			LEFT JOIN deployment d ON (d.id = sd.deployment_id AND d.deleted_at IS NULL)
			WHERE s.app_id = ? AND s.id = ? AND s.deleted_at IS NULL
	`, appID, siteID)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (c Conn) ListSitesInfo(ctx context.Context, appID string) ([]db.SiteInfo, error) {
	var info []db.SiteInfo
	err := c.tx.SelectContext(ctx, &info, `
		SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at, sd.deployment_id, d.name AS deployment_name FROM site s
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

func (c Conn) ListDeploymentSites(ctx context.Context, appID string, deploymentID string) ([]*models.Site, error) {
	var sites []*models.Site

	err := c.tx.SelectContext(ctx, &sites, `
		SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			JOIN site_deployment sd ON (sd.site_id = s.id)
			JOIN deployment d ON (d.id = sd.deployment_id AND d.deleted_at IS NULL)
			WHERE s.app_id = ? AND d.id = ? AND s.deleted_at IS NULL
	`, appID, deploymentID)
	if err != nil {
		return nil, err
	}

	return sites, nil
}

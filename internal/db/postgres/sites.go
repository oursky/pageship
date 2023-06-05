package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

func (q query[T]) CreateSiteIfNotExist(ctx context.Context, site *models.Site) (*db.SiteInfo, error) {
	_, err := sqlx.NamedExecContext(ctx, q.ext, `
		INSERT INTO site (id, app_id, name, created_at, updated_at, deleted_at, deployment_id)
			VALUES (:id, :app_id, :name, :created_at, :updated_at, :deleted_at, :deployment_id)
			ON CONFLICT (app_id, name) WHERE deleted_at IS NULL DO NOTHING
	`, site)
	if err != nil {
		return nil, err
	}

	var info db.SiteInfo
	err = sqlx.GetContext(ctx, q.ext, &info, `
		SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at, s.deployment_id, d.name AS deployment_name FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			LEFT JOIN deployment d ON (d.id = s.deployment_id AND d.deleted_at IS NULL)
			WHERE s.app_id = $1 AND s.name = $2 AND s.deleted_at IS NULL
	`, site.AppID, site.Name)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (q query[T]) GetSiteByName(ctx context.Context, appID string, name string) (*models.Site, error) {
	var site models.Site
	err := sqlx.GetContext(ctx, q.ext, &site, `
		SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			WHERE s.app_id = $1 AND s.name = $2 AND s.deleted_at IS NULL
	`, appID, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrSiteNotFound
	} else if err != nil {
		return nil, err
	}

	return &site, nil
}

func (q query[T]) GetSiteInfo(ctx context.Context, appID string, siteID string) (*db.SiteInfo, error) {
	var info db.SiteInfo
	err := sqlx.GetContext(ctx, q.ext, &info, `
		SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at, s.deployment_id, d.name AS deployment_name FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			LEFT JOIN deployment d ON (d.id = s.deployment_id AND d.deleted_at IS NULL)
			WHERE s.app_id = $1 AND s.id = $2 AND s.deleted_at IS NULL
	`, appID, siteID)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (q query[T]) ListSitesInfo(ctx context.Context, appID string) ([]db.SiteInfo, error) {
	var info []db.SiteInfo
	err := sqlx.SelectContext(ctx, q.ext, &info, `
		SELECT s.id, s.app_id, s.name, s.created_at, s.updated_at, s.deleted_at, s.deployment_id, s.deployment_id, d.name AS deployment_name FROM site s
			JOIN app a ON (a.id = s.app_id AND a.deleted_at IS NULL)
			LEFT JOIN deployment d ON (d.id = s.deployment_id AND d.deleted_at IS NULL)
			WHERE s.app_id = $1 AND s.deleted_at IS NULL
			ORDER BY s.app_id, s.name
	`, appID)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (q query[T]) SetSiteDeployment(ctx context.Context, site *models.Site) error {
	_, err := q.ext.ExecContext(ctx, `
		UPDATE site SET deployment_id = $1, updated_at = $2 WHERE id = $3
	`, site.DeploymentID, site.UpdatedAt, site.ID)
	if err != nil {
		return err
	}

	return nil
}

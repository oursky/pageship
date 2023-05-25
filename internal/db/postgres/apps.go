package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
)

func (c Conn) CreateApp(ctx context.Context, app *models.App) error {
	result, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO app (id, created_at, updated_at, deleted_at, config)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :config)
			ON CONFLICT (id) DO NOTHING
	`, app)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return models.ErrAppUsedID
	}

	return nil
}

func (c Conn) ListApps(ctx context.Context, userID string) ([]*models.App, error) {
	apps := []*models.App{}
	err := c.tx.SelectContext(ctx, &apps, `
		SELECT a.id, a.created_at, a.updated_at, a.deleted_at, a.config FROM app a
			JOIN user_app ua ON (ua.app_id = a.id)
			WHERE a.deleted_at IS NULL AND ua.user_id = $1
			ORDER BY a.id
	`, userID)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (c Conn) GetApp(ctx context.Context, id string) (*models.App, error) {
	var app models.App
	err := c.tx.GetContext(ctx, &app, `
		SELECT id, created_at, updated_at, deleted_at, config FROM app
			WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrAppNotFound
	} else if err != nil {
		return nil, err
	}

	return &app, nil
}

func (c Conn) UpdateAppConfig(ctx context.Context, id string, config *config.AppConfig) (*models.App, error) {
	var app models.App
	err := c.tx.GetContext(ctx, &app, `
		UPDATE app SET config = $1 WHERE id = $2 AND deleted_at IS NULL
			RETURNING id, created_at, updated_at, deleted_at, config
	`, config, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrAppNotFound
	} else if err != nil {
		return nil, err
	}

	return &app, nil
}

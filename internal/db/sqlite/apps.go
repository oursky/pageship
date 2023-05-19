package sqlite

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
		return models.ErrUsedAppID
	}

	return nil
}

func (c Conn) ListApps(ctx context.Context) ([]*models.App, error) {
	apps := []*models.App{}
	err := c.tx.SelectContext(ctx, &apps, `
		SELECT id, created_at, updated_at, deleted_at, config FROM app
			WHERE deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (c Conn) GetApp(ctx context.Context, id string) (*models.App, error) {
	var app models.App
	err := c.tx.GetContext(ctx, &app, `
		SELECT id, created_at, updated_at, deleted_at, config FROM app
			WHERE id = ? AND deleted_at IS NULL
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
		UPDATE app SET config = ? WHERE id = ? AND deleted_at IS NULL
			RETURNING id, created_at, updated_at, deleted_at, config
	`, config, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrAppNotFound
	} else if err != nil {
		return nil, err
	}

	return &app, nil
}

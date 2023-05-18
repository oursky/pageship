package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

func (c Conn) CreateApp(ctx context.Context, id string) (*models.App, error) {
	rows, err := c.tx.QueryContext(ctx, `
		SELECT 1 FROM app WHERE id = ? LIMIT 1
	`, id)
	if err != nil {
		return nil, err
	} else if err := db.EnsureNoRow(rows, models.ErrUsedAppID); err != nil {
		return nil, err
	}

	app := models.NewApp(id, c.clock.Now().UTC())

	_, err = c.tx.NamedExecContext(ctx, `
		INSERT INTO app (id, created_at, updated_at, deleted_at, config)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :config)
	`, app)
	if err != nil {
		return nil, err
	}

	return app, nil
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

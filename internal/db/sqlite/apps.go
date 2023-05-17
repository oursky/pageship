package sqlite

import (
	"context"

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

	now := c.clock.Now().UTC()
	app := &models.App{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	_, err = c.tx.NamedExecContext(ctx, `
		INSERT INTO app (id, created_at, updated_at, deleted_at)
			VALUES (:id, :created_at, :updated_at, :deleted_at)
	`, app)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (c Conn) ListApps(ctx context.Context) ([]*models.App, error) {
	apps := []*models.App{}
	err := c.tx.SelectContext(ctx, &apps, `
		SELECT id, created_at, updated_at, deleted_at FROM app
			WHERE deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

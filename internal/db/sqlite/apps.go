package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/models"
)

func (q query[T]) CreateApp(ctx context.Context, app *models.App) error {
	credentialIDs := app.CredentialIDs()
	cids, err := json.Marshal(credentialIDs)
	if err != nil {
		return err
	}

	a := struct {
		*models.App
		CredentialIDs string `db:"credential_ids"`
	}{app, string(cids)}

	result, err := sqlx.NamedExecContext(ctx, q.ext, `
		INSERT INTO app (id, created_at, updated_at, deleted_at, config, owner_user_id, credential_ids)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :config, :owner_user_id, :credential_ids)
			ON CONFLICT (id) DO NOTHING
	`, a)
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

func (q query[T]) ListApps(ctx context.Context, credentialIDs []models.UserCredentialID) ([]*models.App, error) {
	if len(credentialIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
		SELECT DISTINCT a.id, a.created_at, a.updated_at, a.deleted_at, a.config, a.owner_user_id FROM app a, json_each(a.credential_ids) AS cids
			WHERE a.deleted_at IS NULL AND cids.value IN (?)
			ORDER BY a.id
	`, credentialIDs)
	if err != nil {
		return nil, err
	}

	query = q.ext.Rebind(query)

	apps := []*models.App{}
	err = sqlx.SelectContext(ctx, q.ext, &apps, query, args...)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (q query[T]) GetApp(ctx context.Context, id string) (*models.App, error) {
	var app models.App
	err := sqlx.GetContext(ctx, q.ext, &app, `
		SELECT id, created_at, updated_at, deleted_at, config, owner_user_id FROM app
			WHERE id = ? AND deleted_at IS NULL
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrAppNotFound
	} else if err != nil {
		return nil, err
	}

	return &app, nil
}

func (q query[T]) UpdateAppConfig(ctx context.Context, app *models.App) error {
	credentialIDs := app.CredentialIDs()
	cids, err := json.Marshal(credentialIDs)
	if err != nil {
		return err
	}

	_, err = q.ext.ExecContext(ctx, `
		UPDATE app SET config = ?, credential_ids = ?, updated_at = ? WHERE id = ?
	`, app.Config, string(cids), app.UpdatedAt, app.ID)
	if err != nil {
		return err
	}

	return nil
}

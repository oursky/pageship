package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/models"
)

func (q query[T]) CreateApp(ctx context.Context, app *models.App) error {
	indexKeys := app.CredentialIndexKeys()
	index, err := json.Marshal(indexKeys)
	if err != nil {
		return err
	}

	a := struct {
		*models.App
		CredentialIndex string `db:"credential_index"`
	}{app, string(index)}

	result, err := sqlx.NamedExecContext(ctx, q.ext, `
		INSERT INTO app (id, created_at, updated_at, deleted_at, config, owner_user_id, credential_index)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :config, :owner_user_id, :credential_index)
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

func (q query[T]) ListApps(ctx context.Context, credentialIDs []models.CredentialID) ([]*models.App, error) {
	keys := models.CollectCredentialIDIndexKeys(credentialIDs)
	if len(keys) == 0 {
		return nil, nil
	}

	// ?| operator confuses sqlx.In; construct the query manually.
	vars := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, k := range keys {
		vars[i] = fmt.Sprintf("$%d", i+1)
		args[i] = k
	}
	query := fmt.Sprintf(`
		SELECT DISTINCT a.id, a.created_at, a.updated_at, a.deleted_at, a.config, a.owner_user_id FROM app a
			WHERE a.deleted_at IS NULL AND a.credential_index ?| ARRAY[%s]
			ORDER BY a.id
	`, strings.Join(vars, ", "))

	apps := []*models.App{}
	err := sqlx.SelectContext(ctx, q.ext, &apps, query, args...)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (q query[T]) GetApp(ctx context.Context, id string) (*models.App, error) {
	var app models.App
	err := sqlx.GetContext(ctx, q.ext, &app, `
		SELECT id, created_at, updated_at, deleted_at, config, owner_user_id FROM app
			WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrAppNotFound
	} else if err != nil {
		return nil, err
	}

	return &app, nil
}

func (q query[T]) UpdateAppConfig(ctx context.Context, app *models.App) error {
	indexKeys := app.CredentialIndexKeys()
	index, err := json.Marshal(indexKeys)
	if err != nil {
		return err
	}

	_, err = q.ext.ExecContext(ctx, `
		UPDATE app SET config = $1, credential_index = $2, updated_at = $3 WHERE id = $4
	`, app.Config, string(index), app.UpdatedAt, app.ID)
	if err != nil {
		return err
	}

	return nil
}

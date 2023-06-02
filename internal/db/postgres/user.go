package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/models"
)

func (q query[T]) GetUser(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := sqlx.GetContext(ctx, q.ext, &user, `
		SELECT u.id, u.created_at, u.updated_at, u.deleted_at, u.name FROM "user" u
			WHERE u.id = $1 AND u.deleted_at IS NULL
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (q query[T]) GetCredential(ctx context.Context, id models.UserCredentialID) (*models.UserCredential, error) {
	var cred models.UserCredential
	err := sqlx.GetContext(ctx, q.ext, &cred, `
		SELECT uc.id, uc.created_at, uc.updated_at, uc.deleted_at, uc.user_id, uc.data FROM user_credential uc
			WHERE uc.id = $1 AND uc.deleted_at IS NULL
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return &cred, nil
}

func (q query[T]) CreateUserWithCredential(ctx context.Context, user *models.User, credential *models.UserCredential) error {
	_, err := sqlx.NamedExecContext(ctx, q.ext, `
		INSERT INTO "user" (id, created_at, updated_at, deleted_at, name)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :name)
	`, user)
	if err != nil {
		return err
	}

	_, err = sqlx.NamedExecContext(ctx, q.ext, `
		INSERT INTO user_credential (id, created_at, updated_at, deleted_at, user_id, data)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :user_id, :data)
	`, credential)
	if err != nil {
		return err
	}

	return nil
}

func (q query[T]) UpdateCredentialData(ctx context.Context, cred *models.UserCredential) error {
	_, err := q.ext.ExecContext(ctx, `
		UPDATE user_credential SET data = $1, updated_at = $2 WHERE id = $3
	`, cred.Data, cred.UpdatedAt, cred.ID)
	if err != nil {
		return err
	}

	return nil
}

func (q query[T]) AssignAppUser(ctx context.Context, appID string, userID string) error {
	_, err := q.ext.ExecContext(ctx, `
		INSERT INTO user_app (user_id, app_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id, app_id) DO NOTHING
	`, userID, appID)
	if err != nil {
		return err
	}

	return nil
}

func (q query[T]) UnassignAppUser(ctx context.Context, appID string, userID string) error {
	result, err := q.ext.ExecContext(ctx, `
		DELETE FROM user_app WHERE user_id = $1 AND app_id = $2
	`, userID, appID)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

func (q query[T]) ListAppUsers(ctx context.Context, appID string) ([]*models.User, error) {
	var users []*models.User
	err := sqlx.SelectContext(ctx, q.ext, &users, `
		SELECT u.id, u.created_at, u.updated_at, u.deleted_at, u.name FROM "user" u
			JOIN user_app ua ON (ua.user_id = u.id)
			WHERE ua.app_id = $1 AND u.deleted_at IS NULL
			ORDER BY u.name
	`, appID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (q query[T]) IsAppAccessible(ctx context.Context, appID string, userID string) error {
	err := sqlx.GetContext(ctx, q.ext, new(string), `
		SELECT app_id FROM user_app WHERE user_id = $1 AND app_id = $2
	`, userID, appID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrAppNotFound
	} else if err != nil {
		return err
	}

	return nil
}

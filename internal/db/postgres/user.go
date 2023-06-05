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

func (q query[T]) GetCredential(ctx context.Context, id models.CredentialID) (*models.UserCredential, error) {
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

func (q query[T]) CreateUser(ctx context.Context, user *models.User) error {
	_, err := sqlx.NamedExecContext(ctx, q.ext, `
		INSERT INTO "user" (id, created_at, updated_at, deleted_at, name)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :name)
	`, user)
	if err != nil {
		return err
	}

	return nil
}

func (q query[T]) AddCredential(ctx context.Context, credential *models.UserCredential) error {
	_, err := sqlx.NamedExecContext(ctx, q.ext, `
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

func (q query[T]) ListCredentialIDs(ctx context.Context, userID string) ([]models.CredentialID, error) {
	var ids []models.CredentialID
	err := sqlx.SelectContext(ctx, q.ext, &ids, `
		SELECT id FROM user_credential WHERE user_id = $1 AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

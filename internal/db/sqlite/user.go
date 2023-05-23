package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/oursky/pageship/internal/models"
)

func (c Conn) GetUserByCredential(ctx context.Context, id models.UserCredentialID) (*models.User, error) {
	var user models.User
	err := c.tx.GetContext(ctx, &user, `
		SELECT u.id, u.created_at, u.updated_at, u.deleted_at, u.name FROM user u
			JOIN user_credential uc ON (uc.user_id = u.id AND uc.deleted_at IS NULL)
			WHERE uc.id = ? AND u.deleted_at IS NULL
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c Conn) CreateUserWithCredential(ctx context.Context, user *models.User, credential *models.UserCredential) error {
	_, err := c.tx.NamedExecContext(ctx, `
		INSERT INTO user (id, created_at, updated_at, deleted_at, name)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :name)
	`, user)
	if err != nil {
		return err
	}

	_, err = c.tx.NamedExecContext(ctx, `
		INSERT INTO user_credential (id, created_at, updated_at, deleted_at, user_id, data)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :user_id, :data)
	`, credential)
	if err != nil {
		return err
	}

	return nil
}

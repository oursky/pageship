package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/oursky/pageship/internal/models"
)

func (c Conn) GetUser(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := c.tx.GetContext(ctx, &user, `
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

func (c Conn) GetUserByCredential(ctx context.Context, id models.UserCredentialID) (*models.User, error) {
	var user models.User
	err := c.tx.GetContext(ctx, &user, `
		SELECT u.id, u.created_at, u.updated_at, u.deleted_at, u.name FROM "user" u
			JOIN user_credential uc ON (uc.user_id = u.id AND uc.deleted_at IS NULL)
			WHERE uc.id = $1 AND u.deleted_at IS NULL
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
		INSERT INTO "user" (id, created_at, updated_at, deleted_at, name)
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

func (c Conn) AssignAppUser(ctx context.Context, appID string, userID string) error {
	_, err := c.tx.ExecContext(ctx, `
		INSERT INTO user_app (user_id, app_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id, app_id) DO NOTHING
	`, userID, appID)
	if err != nil {
		return err
	}

	return nil
}

func (c Conn) UnassignAppUser(ctx context.Context, appID string, userID string) error {
	result, err := c.tx.ExecContext(ctx, `
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

func (c Conn) ListAppUsers(ctx context.Context, appID string) ([]*models.User, error) {
	var users []*models.User
	err := c.tx.SelectContext(ctx, &users, `
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

func (c Conn) IsAppAccessible(ctx context.Context, appID string, userID string) error {
	err := c.tx.GetContext(ctx, new(string), `
		SELECT app_id FROM user_app WHERE user_id = $1 AND app_id = $2
	`, userID, appID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrAppNotFound
	} else if err != nil {
		return err
	}

	return nil
}

package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

const tokenValidDuration time.Duration = 10 * time.Minute

func (c *Controller) generateUserToken(
	ctx context.Context,
	username string,
	credentialID models.UserCredentialID,
	data *models.UserCredentialData,
) (string, error) {
	now := c.Clock.Now().UTC()

	user, err := tx(ctx, c.DB, func(conn db.Conn) (*models.User, error) {
		user, err := conn.GetUserByCredential(ctx, credentialID)
		if errors.Is(err, models.ErrUserNotFound) {
			user = nil
			err = nil
		}
		if err != nil {
			return nil, err
		}

		if user == nil {
			user = models.NewUser(now, username)
			cred := models.NewUserCredential(now, user.ID, credentialID, data)

			err = conn.CreateUserWithCredential(ctx, user, cred)
			if err != nil {
				return nil, err
			}
		}

		return user, nil
	})
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &models.TokenClaims{
		Username: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.Config.TokenAuthority,
			Audience:  jwt.ClaimStrings{c.Config.TokenAuthority},
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenValidDuration)),
		},
	}).SignedString(c.Config.TokenSignSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return token, nil
}

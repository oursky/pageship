package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

const tokenValidDuration time.Duration = 30 * time.Minute

const (
	contextUserID    string = "user-id"
	contextAuthnInfo string = "authn"
)

type authnInfo struct {
	UserID string
}

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
	}).SignedString(c.Config.TokenSigningKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return token, nil
}

func (c *Controller) requireAuthn(ctx *gin.Context) bool {
	token, ok := parseAuthorizationHeader(ctx)
	if !ok {
		writeResponse(ctx, nil, models.ErrInvalidCredentials)
		return false
	}

	authn, err := c.verifyToken(ctx, token)
	if err != nil {
		writeResponse(ctx, nil, err)
		return false
	}

	ctx.Set(contextUserID, authn.UserID)
	ctx.Set(contextAuthnInfo, authn)

	return true
}

func (c *Controller) verifyToken(ctx context.Context, token string) (*authnInfo, error) {
	claims := &models.TokenClaims{}
	_, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) { return c.Config.TokenSigningKey, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithAudience(c.Config.TokenAuthority),
		jwt.WithTimeFunc(c.Clock.Now),
	)
	if err != nil {
		return nil, models.ErrInvalidCredentials
	}

	user, err := tx(ctx, c.DB, func(conn db.Conn) (*models.User, error) {
		user, err := conn.GetUser(ctx, claims.Subject)
		if errors.Is(err, models.ErrUserNotFound) {
			return nil, models.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}

		return user, nil
	})
	if err != nil {
		return nil, err
	}

	return &authnInfo{UserID: user.ID}, nil
}

func parseAuthorizationHeader(ctx *gin.Context) (string, bool) {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		return "", false
	}

	bearer, token, ok := strings.Cut(auth, " ")
	if !ok || strings.ToLower(bearer) != "bearer" {
		return "", false
	}

	return token, true
}

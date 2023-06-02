package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/models"
)

const tokenValidDuration time.Duration = 30 * time.Minute

type authnContextKey struct{}

type authnInfo struct {
	UserID string
}

func ensureUser(
	ctx context.Context,
	tx db.Tx,
	now time.Time,
	credentialID models.UserCredentialID,
) (*models.User, *models.UserCredential, error) {
	cred, err := tx.GetCredential(ctx, credentialID)
	if errors.Is(err, models.ErrUserNotFound) {
		user := models.NewUser(now, credentialID.Name())
		cred := models.NewUserCredential(now, user.ID, credentialID, &models.UserCredentialData{})

		err = tx.CreateUserWithCredential(ctx, user, cred)
		if err != nil {
			return nil, nil, err
		}
	} else if err != nil {
		return nil, nil, err
	}

	user, err := tx.GetUser(ctx, cred.UserID)
	if err != nil {
		return nil, nil, err
	}

	return user, cred, nil
}

func (c *Controller) generateUserToken(
	ctx context.Context,
	credentialID models.UserCredentialID,
	data *models.UserCredentialData,
) (string, error) {
	now := c.Clock.Now().UTC()

	user, err := withTx(ctx, c.DB, func(tx db.Tx) (*models.User, error) {
		user, cred, err := ensureUser(ctx, tx, now, credentialID)
		if err != nil {
			return nil, err
		}

		cred.Data = data
		cred.UpdatedAt = now

		err = tx.UpdateCredentialData(ctx, cred)
		if err != nil {
			return nil, err
		}

		return user, nil
	})()
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

func (c *Controller) middlewareAuthn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := parseAuthorizationHeader(r)
		if !ok {
			writeResponse(w, nil, models.ErrInvalidCredentials)
			return
		}

		var authn *authnInfo
		if token != "" {
			info, err := c.verifyToken(r, token)
			if err != nil {
				writeResponse(w, nil, err)
				return
			}
			authn = info

			middleware.GetLogEntry(r).(*httputil.LogEntry).User = authn.UserID
		}

		r = r.WithContext(context.WithValue(r.Context(), authnContextKey{}, authn))
		next.ServeHTTP(w, r)
	})
}

func (c *Controller) verifyToken(r *http.Request, token string) (*authnInfo, error) {
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

	user, err := c.DB.GetUser(r.Context(), claims.Subject)
	if errors.Is(err, models.ErrUserNotFound) {
		return nil, models.ErrInvalidCredentials
	} else if err != nil {
		return nil, err
	}

	return &authnInfo{UserID: user.ID}, nil
}

func authn(r *http.Request) *authnInfo {
	return r.Context().Value(authnContextKey{}).(*authnInfo)
}

func parseAuthorizationHeader(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", true
	}

	bearer, token, ok := strings.Cut(auth, " ")
	if !ok || strings.ToLower(bearer) != "bearer" {
		return "", false
	}

	return token, true
}

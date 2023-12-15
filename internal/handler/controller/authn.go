package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/models"
	"go.uber.org/zap"
)

type authnInfo struct {
	Subject       string
	Name          string
	IsBot         bool
	CredentialIDs []models.CredentialID
}

func (i *authnInfo) UserID() string {
	if i.IsBot {
		return ""
	}
	return i.Subject
}

func createUser(
	ctx context.Context,
	tx db.Tx,
	now time.Time,
	name string,
) (*models.User, error) {
	user := models.NewUser(now, name)
	cred := models.NewUserCredential(now, user.ID, models.CredentialUserID(user.ID), &models.UserCredentialData{})

	err := tx.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	err = tx.AddCredential(ctx, cred)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func ensureUser(
	ctx context.Context,
	tx db.Tx,
	now time.Time,
	name string,
	credentialID models.CredentialID,
) (*models.User, *models.UserCredential, error) {
	cred, err := tx.GetCredential(ctx, credentialID)
	if errors.Is(err, models.ErrUserNotFound) {
		user, err := createUser(ctx, tx, now, name)
		if err != nil {
			return nil, nil, err
		}

		cred = models.NewUserCredential(now, user.ID, credentialID, &models.UserCredentialData{})
		err = tx.AddCredential(ctx, cred)
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
	name string,
	credentialID models.CredentialID,
	data *models.UserCredentialData,
) (string, error) {
	now := c.Clock.Now().UTC()

	user, err := withTx(ctx, c.DB, func(tx db.Tx) (*models.User, error) {
		user, cred, err := ensureUser(ctx, tx, now, name, credentialID)
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

	claims := models.NewTokenClaims(models.TokenSubjectUser(user.ID), user.Name)
	return c.issueToken(claims)
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
		}

		r = set(r, authn)

		if authn != nil {
			loggers := get[*loggers](r)
			loggers.Logger = loggers.Logger.With(zap.String("user", authn.Subject))
			loggers.authn = loggers.Logger

			entry := middleware.GetLogEntry(r)
			if entry != nil {
				e := entry.(*httputil.LogEntry)
				e.Logger = e.Logger.With(zap.String("user", authn.Subject))
			}
		}

		next.ServeHTTP(w, r)
	})
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

func getSubject(r *http.Request) string {
	return get[*authnInfo](r).Subject
}

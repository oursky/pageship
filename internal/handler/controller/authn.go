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
)

type authnInfo struct {
	Subject         string
	Name            string
	IsBot           bool
	CredentialIDs   []models.CredentialID
	CredentialIDMap map[models.CredentialID]struct{}
}

func (i *authnInfo) checkCredentialID(id models.CredentialID) bool {
	if id == "" {
		return false
	}
	_, ok := i.CredentialIDMap[id]
	return ok
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
	credentialID models.CredentialID,
) (*models.User, *models.UserCredential, error) {
	cred, err := tx.GetCredential(ctx, credentialID)
	if errors.Is(err, models.ErrUserNotFound) {
		user, err := createUser(ctx, tx, now, credentialID.Name())
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
	credentialID models.CredentialID,
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
			info, err := c.verifyToken(r.Context(), token)
			if err != nil {
				writeResponse(w, nil, err)
				return
			}
			authn = info

			middleware.GetLogEntry(r).(*httputil.LogEntry).User = authn.Subject
		}

		r = set(r, authn)
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

package app

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oursky/pageship/internal/api"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
)

const reauthThreshold time.Duration = time.Minute * 10

var initialCheck atomic.Bool

func ensureAuth(ctx context.Context) (string, error) {
	conf, err := config.LoadClientConfig()
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}

	token := conf.AuthToken
	if isTokenValid(token) == nil {
		return token, nil
	}

	if canAuthGitHubOIDC() {
		Info("Authenticating using GItHub Actions OIDC token...")
		token, err = authGitHubOIDC(ctx)
	} else {
		token, err = authGitHubSSH(ctx)
	}

	if err == nil {
		initialCheck.Store(true)
		_ = saveToken(token)
	}

	return token, err
}

func saveToken(token string) error {
	conf, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	conf.AuthToken = token
	err = conf.Save()
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

func isTokenValid(token string) error {
	claims := &models.TokenClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)
	if err != nil {
		return err
	}
	if time.Until(claims.ExpiresAt.Time) < reauthThreshold {
		// Expires soon, need reauth
		return models.ErrInvalidCredentials
	}

	if !initialCheck.Swap(true) {
		_, err = apiClient.GetMe(context.Background())
		if code, ok := api.ErrorStatusCode(err); ok && (code == http.StatusUnauthorized || code == http.StatusForbidden) {
			err = models.ErrInvalidCredentials
		} else {
			err = nil
		}
	}
	return err
}

package controller

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oursky/pageship/internal/models"
	"go.uber.org/zap"
)

func (c *Controller) handleAuthGithubOIDC(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"token" binding:"required"`
	}
	if !bindJSON(w, r, &request) {
		return
	}

	key, err := c.oidcKeys.Get("https://token.actions.githubusercontent.com")
	if err != nil {
		writeResponse(w, nil, err)
		return
	}

	type githubOIDCClaims struct {
		jwt.RegisteredClaims
		Repository string `json:"repository"`
	}

	oidcClaims := &githubOIDCClaims{}
	_, err = jwt.ParseWithClaims(
		request.Token,
		oidcClaims,
		key.JWKS.Keyfunc,
		jwt.WithAudience(c.Config.TokenAuthority),
		jwt.WithIssuer(key.Issuer),
		jwt.WithTimeFunc(c.Clock.Now),
	)
	if err != nil {
		log(r).Debug("invalid OIDC token",
			zap.Error(err),
			zap.Any("claims", oidcClaims))
		writeResponse(w, nil, models.ErrInvalidCredentials)
		return
	}

	if oidcClaims.ID == "" {
		log(r).Warn("missing OIDC token ID",
			zap.String("subject", oidcClaims.Subject),
		)
		writeResponse(w, nil, models.ErrInvalidCredentials)
		return
	}

	var credentials []models.CredentialID
	if oidcClaims.Repository != "" {
		credentials = append(credentials, models.CredentialGitHubRepositoryActions(oidcClaims.Repository))
	}

	log(r).Info("github actions authenticated",
		zap.String("subject", oidcClaims.Subject),
		zap.String("token_id", oidcClaims.ID),
		zap.Any("credentials", credentials),
	)

	claims := models.NewTokenClaims(models.TokenSubjectGitHubActions(oidcClaims.ID), oidcClaims.Subject)
	claims.Credentials = credentials

	token, err := c.issueToken(claims)
	writeResponse(w, token, err)
}

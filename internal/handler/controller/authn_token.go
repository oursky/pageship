package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oursky/pageship/internal/models"
)

const tokenValidDuration time.Duration = 5 * time.Minute

func (c *Controller) issueToken(claims *models.TokenClaims) (string, error) {
	now := c.Clock.Now()
	claims.Issuer = c.Config.TokenAuthority
	claims.Audience = jwt.ClaimStrings{c.Config.TokenAuthority}
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(tokenValidDuration))

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(c.Config.TokenSigningKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return token, nil
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

	kind, data, ok := models.TokenSubject(claims.Subject).Parse()
	if !ok {
		return nil, models.ErrInvalidCredentials
	}

	switch kind {
	case models.TokenSubjectKindUser:
		return c.handleTokenUser(ctx, data)
	case models.TokenSubjectKindGitHubActions:
		return c.handleTokenGitHubActions(ctx, claims.Subject, claims.Name, claims.Credentials)
	default:
		panic("unexpected kind: " + kind)
	}
}

func (c *Controller) handleTokenUser(ctx context.Context, userID string) (*authnInfo, error) {
	user, err := c.DB.GetUser(ctx, userID)
	if errors.Is(err, models.ErrUserNotFound) {
		return nil, models.ErrInvalidCredentials
	} else if err != nil {
		return nil, err
	}

	credIDs, err := c.DB.ListCredentialIDs(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	idMap := make(map[models.CredentialID]struct{}, len(credIDs))
	for _, id := range credIDs {
		idMap[id] = struct{}{}
	}

	return &authnInfo{
		Subject:         user.ID,
		Name:            user.Name,
		CredentialIDs:   credIDs,
		CredentialIDMap: idMap,
	}, nil
}

func (c *Controller) handleTokenGitHubActions(
	ctx context.Context,
	subject string,
	name string,
	credentials []models.CredentialID,
) (*authnInfo, error) {
	idMap := make(map[models.CredentialID]struct{}, len(credentials))
	for _, id := range credentials {
		idMap[id] = struct{}{}
	}

	return &authnInfo{
		Subject:         subject,
		Name:            name,
		IsBot:           true,
		CredentialIDs:   credentials,
		CredentialIDMap: idMap,
	}, nil
}

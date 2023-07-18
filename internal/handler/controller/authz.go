package controller

import (
	"net/http"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
	"go.uber.org/zap"
)

func (c *Controller) requireAccess(level config.AccessLevel) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := get[*authnInfo](r)
			if info == nil {
				writeResponse(w, nil, models.ErrInvalidCredentials)
				return
			}

			app := get[*models.App](r)
			authz, err := app.CheckAuthz(level, info.UserID(), info.CredentialIDs)
			if err != nil {
				writeResponse(w, nil, err)
				return
			}

			fields := []zap.Field{
				zap.String("app", string(app.ID)),
				zap.String("credential", string(authz.CredentialID)),
				zap.String("credential_rule", authz.MatchedRule()),
			}

			loggers := get[*loggers](r)
			loggers.Logger = loggers.authn.With(fields...) // Replace authz logger fields

			next.ServeHTTP(w, r)
		})
	}
}

func (c *Controller) requireAccessAdmin() func(next http.Handler) http.Handler {
	return c.requireAccess(config.AccessLevelAdmin)
}

func (c *Controller) requireAccessDeployer() func(next http.Handler) http.Handler {
	return c.requireAccess(config.AccessLevelDeployer)
}

func (c *Controller) requireAccessReader() func(next http.Handler) http.Handler {
	return c.requireAccess(config.AccessLevelReader)
}

func denyBot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := get[*authnInfo](r)
		if info.IsBot {
			writeResponse(w, nil, models.ErrAccessDenied)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := get[*authnInfo](r)
		if info == nil {
			writeResponse(w, nil, models.ErrInvalidCredentials)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (c *Controller) checkACL(r *http.Request, incoming []models.CredentialID) error {
	if c.Config.ACL != nil {
		acl, err := c.Config.ACL.Get(r.Context())
		if err != nil {
			return err
		}

		creds := make([]models.CredentialID, len(incoming))
		copy(creds, incoming)
		creds = appendRequestCredentials(r, creds)

		if _, err := models.CheckACLAuthz(acl, creds); err != nil {
			log(r).Info("user rejected", zap.Any("credentials", creds))
			return err
		}
	}
	return nil
}

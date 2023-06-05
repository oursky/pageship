package controller

import (
	"net/http"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

func checkAuthz(r *http.Request, q db.DBQuery, level config.AccessLevel, authn *authnInfo) error {
	app := get[*models.App](r)
	if app.OwnerUserID != authn.UserID {
		return models.ErrAccessDenied
	}
	return nil
}

func (c *Controller) requireAccess(level config.AccessLevel) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := get[*authnInfo](r)
			if info == nil {
				writeResponse(w, nil, models.ErrInvalidCredentials)
				return
			}

			err := checkAuthz(r, c.DB, level, info)
			if err != nil {
				writeResponse(w, nil, err)
				return
			}

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

package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/storage"
	apptime "github.com/oursky/pageship/internal/time"
	"go.uber.org/zap"
)

type Controller struct {
	Logger  *zap.Logger
	Clock   apptime.Clock
	Config  Config
	Storage *storage.Storage
	DB      db.DB
}

func (c *Controller) Handler() http.Handler {
	if c.Clock == nil {
		c.Clock = apptime.SystemClock
	}

	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/healthz"))
	r.Use(c.middlewareAuthn)

	r.Route("/api/v1", func(r chi.Router) {
		r.With(c.requireAuth()).Get("/apps", c.handleAppList)
		r.With(c.requireAuth()).Post("/apps", c.handleAppCreate)
		r.With(c.requireAuth(authzReadApp)).Get("/apps/{app-id}", c.handleAppGet)
		r.With(c.requireAuth(authzReadApp)).Get("/apps/{app-id}/config", c.handleAppConfigGet)
		r.With(c.requireAuth(authzWriteApp)).Put("/apps/{app-id}/config", c.handleAppConfigSet)
		r.With(c.requireAuth(authzReadApp)).Get("/apps/{app-id}/users", c.handleAppUserList)
		r.With(c.requireAuth(authzWriteApp)).Post("/apps/{app-id}/users", c.handleAppUserAdd)
		r.With(c.requireAuth(authzWriteApp)).Delete("/apps/{app-id}/users/{user-id}", c.handleAppUserDelete)

		r.With(c.requireAuth(authzReadApp)).Get("/apps/{app-id}/sites", c.handleSiteList)
		r.With(c.requireAuth(authzWriteApp)).Post("/apps/{app-id}/sites", c.handleSiteCreate)
		r.With(c.requireAuth(authzWriteApp)).Patch("/apps/{app-id}/sites/{site-name}", c.handleSiteUpdate)

		r.With(c.requireAuth(authzReadApp)).Get("/apps/{app-id}/deployments", c.handleDeploymentList)
		r.With(c.requireAuth(authzWriteApp)).Post("/apps/{app-id}/deployments", c.handleDeploymentCreate)
		r.With(c.requireAuth(authzReadApp)).Get("/apps/{app-id}/deployments/{deployment-name}", c.handleDeploymentGet)
		r.With(c.requireAuth(authzWriteApp)).Put("/apps/{app-id}/deployments/{deployment-name}/tarball", c.handleDeploymentUpload)

		r.With(c.requireAuth()).Get("/auth/me", c.handleMe)
		r.Get("/auth/github-ssh", c.handleAuthGithubSSH)
	})
	return r
}

package controller

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/oidc"
	"github.com/oursky/pageship/internal/sshkey"
	"github.com/oursky/pageship/internal/storage"
	apptime "github.com/oursky/pageship/internal/time"
	"go.uber.org/zap"
)

type Controller struct {
	Context context.Context
	Logger  *zap.Logger
	Clock   apptime.Clock
	Config  Config
	Storage *storage.Storage
	DB      db.DB

	githubKeys *sshkey.GitHubKeys
	oidcKeys   *oidc.Keys
}

func (c *Controller) Handler() http.Handler {
	if c.Clock == nil {
		c.Clock = apptime.SystemClock
	}
	if c.githubKeys == nil {
		keys, err := sshkey.NewGitHubKeys(c.Context)
		if err != nil {
			panic(err)
		}
		c.githubKeys = keys
	}
	if c.oidcKeys == nil {
		keys, err := oidc.NewKeys(c.Context)
		if err != nil {
			panic(err)
		}
		c.oidcKeys = keys
	}

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = set(r, &loggers{
				Logger: c.Logger.With(zap.String("request_id", requestID(r))),
			})
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Heartbeat("/healthz"))
	r.Use(c.middlewareAuthn)

	r.Route("/api/v1", func(r chi.Router) {
		r.With(requireAuth).Route("/apps", func(r chi.Router) {
			r.Get("/", c.handleAppList)
			r.With(denyBot).Post("/", c.handleAppCreate)

			r.With(
				c.middlewareLoadApp(),
				c.requireAccessReader(),
			).Route("/{app-id}", func(r chi.Router) {
				r.Get("/", c.handleAppGet)
				r.Get("/config", c.handleAppConfigGet)
				r.With(c.requireAccessAdmin()).Put("/config", c.handleAppConfigSet)

				r.Route("/sites", func(r chi.Router) {
					r.Get("/", c.handleSiteList)
					r.With(c.requireAccessDeployer()).Post("/", c.handleSiteCreate)

					r.With(c.middlewareLoadSite()).Route("/{site-name}", func(r chi.Router) {
						r.With(c.requireAccessDeployer()).Patch("/", c.handleSiteUpdate)
					})
				})

				r.Route("/deployments", func(r chi.Router) {
					r.Get("/", c.handleDeploymentList)
					r.With(c.requireAccessDeployer()).Post("/", c.handleDeploymentCreate)

					r.With(c.middlewareLoadDeployment()).Route("/{deployment-name}", func(r chi.Router) {
						r.With(c.requireAccessDeployer()).Get("/", c.handleDeploymentGet)
						r.With(c.requireAccessDeployer()).Put("/tarball", c.handleDeploymentUpload)
					})
				})
			})
		})

		r.With(requireAuth).Get("/auth/me", c.handleMe)
		r.Get("/auth/github-ssh", c.handleAuthGithubSSH)
		r.Post("/auth/github-oidc", c.handleAuthGithubOIDC)
	})
	return r
}

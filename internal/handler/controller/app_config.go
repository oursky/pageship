package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"go.uber.org/zap"
)

func (c *Controller) handleAppConfigGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "app-id")

	respond(w, func() (any, error) {
		app, err := c.DB.GetApp(r.Context(), id)
		if err != nil {
			return nil, err
		}

		return app.Config, nil
	})
}

func (c *Controller) handleAppConfigSet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "app-id")

	var request struct {
		Config *config.AppConfig `json:"config" binding:"required"`
	}
	if !bindJSON(w, r, &request) {
		return
	}

	if err := config.ValidateAppConfig(request.Config); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: err})
		return
	}

	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		app, err := c.DB.UpdateAppConfig(r.Context(), id, request.Config)
		if err != nil {
			return nil, err
		}

		c.Logger.Info("updating config",
			zap.String("request_id", requestID(r)),
			zap.String("user", authn(r).UserID),
			zap.String("app", id),
		)

		return app.Config, nil
	}))
}

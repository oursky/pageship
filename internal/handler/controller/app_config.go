package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
)

func (c *Controller) handleAppConfigGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "app-id")

	config, err := tx(r.Context(), c.DB, func(conn db.Conn) (*config.AppConfig, error) {
		app, err := conn.GetApp(r.Context(), id)
		if err != nil {
			return nil, err
		}
		return app.Config, nil
	})

	writeResponse(w, config, err)
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

	config, err := tx(r.Context(), c.DB, func(conn db.Conn) (*config.AppConfig, error) {
		app, err := conn.UpdateAppConfig(r.Context(), id, request.Config)
		if err != nil {
			return nil, err
		}

		return app.Config, nil
	})

	writeResponse(w, config, err)
}

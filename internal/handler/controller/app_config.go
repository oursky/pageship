package controller

import (
	"net/http"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

func (c *Controller) handleAppConfigGet(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, get[*models.App](r).Config, nil)
}

func (c *Controller) handleAppConfigSet(w http.ResponseWriter, r *http.Request) {
	app := get[*models.App](r)

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
		app.Config = request.Config
		app.UpdatedAt = c.Clock.Now().UTC()

		err := c.DB.UpdateAppConfig(r.Context(), app)
		if err != nil {
			return nil, err
		}

		log(r).Info("updating config")

		return app.Config, nil
	}))
}

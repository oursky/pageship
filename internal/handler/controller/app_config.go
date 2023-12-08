package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
	"go.uber.org/zap"
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
		now := c.Clock.Now().UTC()
		app.UpdatedAt = now

		err := tx.UpdateAppConfig(r.Context(), app)
		if err != nil {
			return nil, err
		}

		log(r).Info("updating config")

		// Deactivated removed domains; added domains need manual activation.
		domains, err := tx.ListDomains(r.Context(), app.ID)
		if err != nil && !errors.Is(err, models.ErrDomainNotFound) {
			return nil, err
		}
		for _, d := range domains {
			if _, exists := app.Config.ResolveDomain(d.Domain); exists {
				continue
			}

			err = tx.DeleteDomain(r.Context(), d.ID, now)
			if err != nil {
				return nil, fmt.Errorf("failed to deactivate domain: %w", err)
			}

			log(r).Info("deleting domain", zap.String("domain", d.Domain))
		}
		domainVerifications, err := tx.ListDomainVerifications(r.Context(), &app.ID, nil, nil)
		if err != nil && !errors.Is(err, models.ErrDomainNotFound) {
			return nil, err
		}
		for _, d := range domainVerifications {
			if _, exists := app.Config.ResolveDomain(d.Domain); exists && c.Config.DomainVerification {
				continue
			}

			err = tx.DeleteDomainVerification(r.Context(), d.ID, now)
			if err != nil {
				return nil, fmt.Errorf("failed to deactivate domain: %w", err)
			}

			log(r).Info("deleting domain verification", zap.String("domain", d.Domain))
		}

		return app.Config, nil
	}))
}

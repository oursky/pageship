package controller

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
	"go.uber.org/zap"
)

type apiDomain struct {
	*models.Domain
}

func (c *Controller) makeAPIDomain(domain *models.Domain) *apiDomain {
	return &apiDomain{Domain: domain}
}

func (c *Controller) handleDomainList(w http.ResponseWriter, r *http.Request) {
	app := get[*models.App](r)

	respond(w, func() (any, error) {
		var domains []*models.Domain
		for _, dconf := range app.Config.Domains {
			domain, err := c.DB.GetDomainByName(r.Context(), dconf.Domain)
			if errors.Is(err, models.ErrDomainNotFound) {
				continue
			} else if err != nil {
				return nil, err
			}
			domains = append(domains, domain)
		}

		return mapModels(domains, func(d *models.Domain) *apiDomain {
			return c.makeAPIDomain(d)
		}), nil
	})
}

func (c *Controller) handleDomainCreate(w http.ResponseWriter, r *http.Request) {
	app := get[*models.App](r)

	domainName := chi.URLParam(r, "domain-name")
	replaceApp := r.URL.Query().Get("replaceApp")

	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		config, ok := app.Config.ResolveDomain(domainName)
		if !ok {
			return nil, models.ErrUndefinedDomain
		}

		domain, err := tx.GetDomainByName(r.Context(), domainName)
		if errors.Is(err, models.ErrDomainNotFound) {
			// Continue create new domain.
		} else if err != nil {
			return nil, err
		} else {
			if domain.AppID == app.ID {
				return c.makeAPIDomain(domain), nil
			} else if replaceApp != domain.AppID {
				return nil, models.ErrDomainUsedName
			}

			err = tx.DeleteDomain(r.Context(), domain.ID, c.Clock.Now().UTC())
			if err != nil {
				return nil, err
			}
		}

		domain = models.NewDomain(c.Clock.Now().UTC(), domainName, app.ID, config.Site)
		err = tx.CreateDomain(r.Context(), domain)
		if err != nil {
			return nil, err
		}

		log(r).Info("creating domain",
			zap.String("domain", domain.Domain),
			zap.String("site", domain.SiteName))

		return c.makeAPIDomain(domain), nil
	}))
}

func (c *Controller) handleDomainDelete(w http.ResponseWriter, r *http.Request) {
	app := get[*models.App](r)

	domainName := chi.URLParam(r, "domain-name")

	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		domain, err := tx.GetDomainByName(r.Context(), domainName)
		if err != nil {
			return nil, err
		}
		if domain.AppID != app.ID {
			return nil, models.ErrAccessDenied
		}

		err = tx.DeleteDomain(r.Context(), domain.ID, c.Clock.Now().UTC())
		if err != nil {
			return nil, err
		}

		log(r).Info("deleting domain",
			zap.String("domain", domain.Domain),
			zap.String("site", domain.SiteName))

		return struct{}{}, nil
	}))
}

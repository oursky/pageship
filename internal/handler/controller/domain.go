package controller

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oursky/pageship/internal/api"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
	"go.uber.org/zap"
)

func (c *Controller) makeAPIDomain(domain *models.Domain, dominVerification *models.DomainVerification) *api.APIDomain {
	return &api.APIDomain{Domain: domain, DomainVerification: dominVerification}
}

type apiDomainVerification struct {
	*models.DomainVerification
}

type DomainInfo struct {
	domain       *models.Domain
	verification *models.DomainVerification
}

func (c *Controller) handleDomainList(w http.ResponseWriter, r *http.Request) {
	app := get[*models.App](r)

	respond(w, func() (any, error) {
		var domains []DomainInfo
		for _, dconf := range app.Config.Domains {
			domain, err := c.DB.GetDomainByName(r.Context(), dconf.Domain)
			if err != nil && !errors.Is(err, models.ErrDomainNotFound) {
				return nil, err
			}
			domainVerification, err := c.DB.GetDomainVerificationByName(r.Context(), dconf.Domain, app.ID)
			if err != nil && !errors.Is(err, models.ErrDomainNotFound) {
				return nil, err
			}
			if domain != nil || domainVerification != nil {
				domains = append(domains, DomainInfo{domain: domain, verification: domainVerification})
			}
		}

		return mapModels(domains, func(d DomainInfo) *api.APIDomain {
			return c.makeAPIDomain(d.domain, d.verification)
		}), nil
	})
}

func (c *Controller) handleDomainVerification(w http.ResponseWriter, r *http.Request) {
	app := get[*models.App](r)
	domainName := chi.URLParam(r, "domain-name")

	config, ok := app.Config.ResolveDomain(domainName)
	if !ok {
		respond(w, func() (any, error) { return nil, models.ErrUndefinedDomain })
	}

	domain, _ := c.DB.GetDomainByName(r.Context(), domainName)
	if domain != nil {
		c.handleDomainCreate(w, r)
		return
	}
	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		var domainVerification *models.DomainVerification
		domainVerification, _ = tx.GetDomainVerificationByName(r.Context(), domainName, app.ID)
		if domainVerification == nil {
			domainVerification = models.NewDomainVerification(c.Clock.Now().UTC(), domainName, app.ID)
			err := tx.CreateDomainVerification(r.Context(), domainVerification)
			if err != nil {
				return nil, err
			}
			log(r).Info("creating domain verification",
				zap.String("domain", domainName),
				zap.String("site", config.Site))
		} else if domainVerification.WillCheckAt == nil {
			err := tx.ScheduleDomainVerificationAt(r.Context(), domainVerification.ID, c.Clock.Now().UTC())
			if err != nil {
				return nil, err
			}
			log(r).Info("triggering domain verification",
				zap.String("domain", domainName),
				zap.String("site", config.Site))
		}
		domainVerification, _ = tx.GetDomainVerificationByName(r.Context(), domainName, app.ID)
		return c.makeAPIDomain(domain, domainVerification), nil
	}))
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
		domainVerification, _ := tx.GetDomainVerificationByName(r.Context(), domainName, app.ID)
		if errors.Is(err, models.ErrDomainNotFound) {
			// Continue create new domain.
		} else if err != nil {
			return nil, err
		} else {
			if domain.AppID == app.ID {
				return c.makeAPIDomain(domain, domainVerification), nil
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

		return c.makeAPIDomain(domain, domainVerification), nil
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

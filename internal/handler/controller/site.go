package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
	"go.uber.org/zap"
)

type apiSite struct {
	*models.Site
	URL            string  `json:"url"`
	DeploymentName *string `json:"deploymentName"`
}

func (c *Controller) makeAPISite(app *models.App, site db.SiteInfo) *apiSite {
	sub := site.Name
	if site.Name == app.Config.DefaultSite {
		sub = ""
	}

	return &apiSite{
		Site: site.Site,
		URL: c.Config.HostPattern.MakeURL(
			c.Config.HostIDScheme.Make(site.AppID, sub),
		),
		DeploymentName: site.DeploymentName,
	}
}

func (c *Controller) handleSiteList(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")

	respond(w, func() (any, error) {
		app, err := c.DB.GetApp(r.Context(), appID)
		if err != nil {
			return nil, err
		}

		sites, err := c.DB.ListSitesInfo(r.Context(), appID)
		if err != nil {
			return nil, err
		}

		return mapModels(sites, func(site db.SiteInfo) *apiSite {
			return c.makeAPISite(app, site)
		}), nil
	})
}

func (c *Controller) handleSiteCreate(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")

	var request struct {
		Name string `json:"name" binding:"required,dnsLabel"`
	}
	if !bindJSON(w, r, &request) {
		return
	}

	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		app, err := tx.GetApp(r.Context(), appID)
		if err != nil {
			return nil, err
		}

		if _, ok := app.Config.ResolveSite(request.Name); !ok {
			return nil, models.ErrUndefinedSite
		}

		site := models.NewSite(c.Clock.Now().UTC(), appID, request.Name)
		info, err := tx.CreateSiteIfNotExist(r.Context(), site)
		if err != nil {
			return nil, err
		}

		c.Logger.Info("creating site",
			zap.String("request_id", requestID(r)),
			zap.String("user", authn(r).UserID),
			zap.String("app", appID),
			zap.String("deployment", site.ID),
		)

		return c.makeAPISite(app, *info), nil
	}))
}

func (c *Controller) updateDeploymentExpiry(
	ctx context.Context,
	tx db.Tx,
	now time.Time,
	conf *config.AppConfig,
	deployment *models.Deployment,
) error {
	sites, err := tx.GetDeploymentSiteNames(ctx, deployment)
	if err != nil {
		return err
	}

	if len(sites) == 0 && deployment.ExpireAt == nil {
		deploymentTTL, err := time.ParseDuration(conf.Deployments.TTL)
		if err != nil {
			return err
		}

		expireAt := now.Add(deploymentTTL)
		deployment.ExpireAt = &expireAt
		err = tx.SetDeploymentExpiry(ctx, deployment)
		if err != nil {
			return err
		}
	} else if len(sites) > 0 && deployment.ExpireAt != nil {
		deployment.ExpireAt = nil
		err = tx.SetDeploymentExpiry(ctx, deployment)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) siteUpdateDeploymentName(
	ctx context.Context,
	tx db.Tx,
	now time.Time,
	conf *config.AppConfig,
	site *models.Site,
	deploymentName string,
) error {
	var currentDeployment *models.Deployment
	if site.DeploymentID != nil {
		d, err := tx.GetDeployment(ctx, site.AppID, *site.DeploymentID)
		if err != nil {
			return err
		}

		if d.Name == deploymentName {
			// Same deployment
			return nil
		}
		currentDeployment = d
	} else if deploymentName == "" {
		// Same deployment
		return nil
	}

	var newDeployment *models.Deployment
	if deploymentName != "" {
		d, err := tx.GetDeploymentByName(ctx, site.AppID, deploymentName)
		if err != nil {
			return err
		}

		if err := d.CheckAlive(now); err != nil {
			return err
		}

		err = tx.AssignDeploymentSite(ctx, d, site.ID)
		if err != nil {
			return err
		}
		site.DeploymentID = &d.ID
		newDeployment = d
	} else {
		err := tx.UnassignDeploymentSite(ctx, currentDeployment, site.ID)
		if err != nil {
			return err
		}
		site.DeploymentID = nil
		newDeployment = nil
	}

	if currentDeployment != nil {
		if err := c.updateDeploymentExpiry(ctx, tx, now, conf, currentDeployment); err != nil {
			return err
		}
	}
	if newDeployment != nil {
		if err := c.updateDeploymentExpiry(ctx, tx, now, conf, newDeployment); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handleSiteUpdate(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")
	siteName := chi.URLParam(r, "site-name")

	var request struct {
		DeploymentName *string `json:"deploymentName,omitempty" binding:"omitempty"`
	}
	if !bindJSON(w, r, &request) {
		return
	}

	now := c.Clock.Now().UTC()

	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		app, err := tx.GetApp(r.Context(), appID)
		if err != nil {
			return nil, err
		}

		site, err := tx.GetSiteByName(r.Context(), appID, siteName)
		if err != nil {
			return nil, err
		}

		if request.DeploymentName != nil {
			oldDeployment := ""
			if site.DeploymentID != nil {
				oldDeployment = *site.DeploymentID
			}
			c.Logger.Info("updating deployment",
				zap.String("request_id", requestID(r)),
				zap.String("user", authn(r).UserID),
				zap.String("app", appID),
				zap.String("old_deployment", oldDeployment),
				zap.String("new_deployment", *request.DeploymentName),
			)

			if err := c.siteUpdateDeploymentName(r.Context(), tx, now, app.Config, site, *request.DeploymentName); err != nil {
				return nil, err
			}
		}

		info, err := tx.GetSiteInfo(r.Context(), appID, site.ID)
		if err != nil {
			return nil, err
		}

		return c.makeAPISite(app, *info), nil
	}))
}

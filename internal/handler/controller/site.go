package controller

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
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

func (c *Controller) handleSiteList(ctx *gin.Context) {
	appID := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzReadApp(appID)) {
		return
	}

	sites, err := tx(ctx, c.DB, func(conn db.Conn) ([]*apiSite, error) {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return nil, err
		}

		sites, err := conn.ListSitesInfo(ctx, appID)
		if err != nil {
			return nil, err
		}

		return mapModels(sites, func(site db.SiteInfo) *apiSite {
			return c.makeAPISite(app, site)
		}), nil
	})

	writeResponse(ctx, sites, err)
}

func (c *Controller) handleSiteCreate(ctx *gin.Context) {
	appID := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzWriteApp(appID)) {
		return
	}

	var request struct {
		Name string `json:"name" binding:"required,dnsLabel"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	site, err := tx(ctx, c.DB, func(conn db.Conn) (*apiSite, error) {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return nil, err
		}

		if _, ok := app.Config.ResolveSite(request.Name); !ok {
			return nil, models.ErrUndefinedSite
		}

		site := models.NewSite(c.Clock.Now().UTC(), appID, request.Name)
		info, err := conn.CreateSiteIfNotExist(ctx, site)
		if err != nil {
			return nil, err
		}

		return c.makeAPISite(app, *info), nil
	})

	writeResponse(ctx, site, err)
}

func (c *Controller) updateDeploymentExpiry(
	ctx context.Context,
	conn db.Conn,
	now time.Time,
	conf *config.AppConfig,
	deployment *models.Deployment,
) error {
	sites, err := conn.GetDeploymentSiteNames(ctx, deployment)
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
		err = conn.SetDeploymentExpiry(ctx, deployment)
		if err != nil {
			return err
		}
	} else if len(sites) > 0 && deployment.ExpireAt != nil {
		deployment.ExpireAt = nil
		err = conn.SetDeploymentExpiry(ctx, deployment)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) siteUpdateDeploymentName(
	ctx context.Context,
	conn db.Conn,
	now time.Time,
	conf *config.AppConfig,
	site *models.Site,
	deploymentName string,
) error {
	var currentDeployment *models.Deployment
	if site.DeploymentID != nil {
		d, err := conn.GetDeployment(ctx, site.AppID, *site.DeploymentID)
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
		d, err := conn.GetDeploymentByName(ctx, site.AppID, deploymentName)
		if err != nil {
			return err
		}

		if err := d.CheckAlive(now); err != nil {
			return err
		}

		err = conn.AssignDeploymentSite(ctx, d, site.ID)
		if err != nil {
			return err
		}
		site.DeploymentID = &d.ID
		newDeployment = d
	} else {
		err := conn.UnassignDeploymentSite(ctx, currentDeployment, site.ID)
		if err != nil {
			return err
		}
		site.DeploymentID = nil
		newDeployment = nil
	}

	if currentDeployment != nil {
		if err := c.updateDeploymentExpiry(ctx, conn, now, conf, currentDeployment); err != nil {
			return err
		}
	}
	if newDeployment != nil {
		if err := c.updateDeploymentExpiry(ctx, conn, now, conf, newDeployment); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handleSiteUpdate(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	siteName := ctx.Param("site-name")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzWriteApp(appID)) {
		return
	}

	var request struct {
		DeploymentName *string `json:"deploymentName,omitempty" binding:"omitempty"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	now := c.Clock.Now().UTC()

	site, err := tx(ctx, c.DB, func(conn db.Conn) (*apiSite, error) {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return nil, err
		}

		site, err := conn.GetSiteByName(ctx, appID, siteName)
		if err != nil {
			return nil, err
		}

		if request.DeploymentName != nil {
			if err := c.siteUpdateDeploymentName(ctx, conn, now, app.Config, site, *request.DeploymentName); err != nil {
				return nil, err
			}
		}

		info, err := conn.GetSiteInfo(ctx, appID, site.ID)
		if err != nil {
			return nil, err
		}

		return c.makeAPISite(app, *info), nil
	})

	writeResponse(ctx, site, err)
}

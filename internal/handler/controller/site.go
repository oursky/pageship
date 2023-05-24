package controller

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type apiSite struct {
	*models.Site
	URL            string  `json:"url"`
	DeploymentName *string `json:"deploymentName"`
}

func (c *Controller) makeAPISite(app *models.App, site db.SiteInfo) *apiSite {
	hostID := site.AppID
	if app.Config.DefaultSite != site.Name {
		hostID = site.Name + "." + site.AppID
	}

	return &apiSite{
		Site:           site.Site,
		URL:            c.Config.HostPattern.MakeURL(hostID),
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
			if *request.DeploymentName != "" {
				deployment, err := conn.GetDeploymentByName(ctx, appID, *request.DeploymentName)
				if err != nil {
					return nil, err
				}

				if err := deployment.CheckAlive(now); err != nil {
					return nil, err
				}

				err = conn.AssignDeploymentSite(ctx, deployment, site.ID)
				if err != nil {
					return nil, err
				}

				if deployment.ExpireAt != nil {
					deployment.ExpireAt = nil
					err = conn.SetDeploymentExpiry(ctx, deployment)
					if err != nil {
						return nil, err
					}
				}
			} else {
				deployment, err := conn.GetSiteDeployment(ctx, appID, siteName)
				if errors.Is(err, models.ErrDeploymentNotFound) {
					err = nil
				}
				if err != nil {
					return nil, err
				}

				if deployment != nil {
					err = conn.UnassignDeploymentSite(ctx, deployment, siteName)
					if err != nil {
						return nil, err
					}

					sites, err := conn.GetDeploymentSiteNames(ctx, deployment)
					if err != nil {
						return nil, err
					}

					if len(sites) == 0 {
						deploymentTTL, err := time.ParseDuration(app.Config.Deployments.TTL)
						if err != nil {
							return nil, err
						}

						expireAt := now.Add(deploymentTTL)
						deployment.ExpireAt = &expireAt
						err = conn.SetDeploymentExpiry(ctx, deployment)
						if err != nil {
							return nil, err
						}
					}
				}
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

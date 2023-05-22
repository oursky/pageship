package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type apiSite struct {
	*models.Site
	URL            string  `json:"url"`
	DeploymentName *string `json:"deploymentName"`
}

func (c *Controller) makeAPISite(app *models.App, site db.SiteInfo) apiSite {
	hostID := site.AppID
	if app.Config.DefaultSite != site.Name {
		hostID = site.Name + "." + site.AppID
	}

	return apiSite{
		Site:           site.Site,
		URL:            c.Config.HostPattern.MakeURL(hostID),
		DeploymentName: site.DeploymentName,
	}
}

func (c *Controller) handleSiteList(ctx *gin.Context) {
	appID := ctx.Param("app-id")

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return err
		}

		sites, err := conn.ListSitesInfo(ctx, appID)
		if err != nil {
			return err
		}

		result := mapModels(sites, func(site db.SiteInfo) apiSite {
			return c.makeAPISite(app, site)
		})

		ctx.JSON(http.StatusOK, response{Result: result})
		return nil
	})

	switch {
	case errors.Is(err, models.ErrAppNotFound):
		ctx.JSON(http.StatusNotFound, response{Error: err})

	case err != nil:
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (c *Controller) handleSiteCreate(ctx *gin.Context) {
	appID := ctx.Param("app-id")

	var request struct {
		Name string `json:"name" binding:"required,dnsLabel"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return err
		}

		if _, ok := app.Config.ResolveSite(request.Name); !ok {
			return models.ErrUndefinedSite
		}

		site := models.NewSite(c.Clock.Now().UTC(), appID, request.Name)
		info, err := conn.CreateSiteIfNotExist(ctx, site)
		if err != nil {
			return err
		}

		result := c.makeAPISite(app, *info)

		ctx.JSON(http.StatusOK, response{Result: result})
		return nil
	})

	switch {
	case errors.Is(err, models.ErrAppNotFound):
		ctx.JSON(http.StatusNotFound, response{Error: err})

	case err != nil:
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (c *Controller) handleSiteUpdate(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	siteName := ctx.Param("site-name")

	var request struct {
		DeploymentName *string `json:"deploymentName,omitempty" binding:"omitempty"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return err
		}

		site, err := conn.GetSiteByName(ctx, appID, siteName)
		if err != nil {
			return err
		}

		if request.DeploymentName != nil {
			deployment, err := conn.GetDeploymentByName(ctx, appID, *request.DeploymentName)
			if err != nil {
				return err
			}

			if deployment.UploadedAt == nil {
				return models.ErrDeploymentNotUploaded
			}

			err = conn.AssignDeploymentSite(ctx, deployment, site.ID)
			if err != nil {
				return err
			}
		}

		info, err := conn.GetSiteInfo(ctx, appID, site.ID)
		if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, response{Result: c.makeAPISite(app, *info)})
		return nil
	})

	if err != nil {
		if errors.Is(err, models.ErrDeploymentNotFound) {
			ctx.JSON(http.StatusNotFound, response{Error: err})
		} else if errors.Is(err, models.ErrDeploymentNotUploaded) {
			ctx.JSON(http.StatusBadRequest, response{Error: err})
		} else {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
	}
}

package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type apiSite struct {
	*models.Site
	URL          string     `json:"url"`
	LastDeployAt *time.Time `json:"uploadedAt"`
}

func (c *Controller) makeAPISite(app *models.App, site db.SiteInfo) apiSite {
	hostID := site.AppID
	if app.Config.DefaultSite != site.Name {
		hostID = site.Name + "." + site.AppID
	}

	return apiSite{
		Site:         site.Site,
		URL:          c.Config.HostPattern.MakeURL(hostID),
		LastDeployAt: site.LastDeployedAt,
	}
}

func (c *Controller) handleSiteList(ctx *gin.Context) {
	appID := ctx.Param("app-id")

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return err
		}

		sites, err := conn.ListSites(ctx, appID)
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

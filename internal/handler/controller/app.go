package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type apiApp struct {
	*models.App
	URL string `json:"url"`
}

func (c *Controller) makeAPIApp(app *models.App) apiApp {
	return apiApp{
		App: app,
		URL: c.Config.HostPattern.MakeURL(app.ID),
	}
}

func (c *Controller) handleAppCreate(ctx *gin.Context) {
	var request struct {
		ID string `json:"id" binding:"required"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		app := models.NewApp(c.Clock.Now().UTC(), request.ID)

		err := conn.CreateApp(ctx, app)
		if errors.Is(err, models.ErrUsedAppID) {
			ctx.JSON(http.StatusBadRequest, response{Error: err})
			return db.ErrRollback
		} else if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, response{Result: c.makeAPIApp(app)})
		return nil
	})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (c *Controller) handleAppGet(ctx *gin.Context) {
	id := ctx.Param("app-id")

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		app, err := conn.GetApp(ctx, id)
		if errors.Is(err, models.ErrAppNotFound) {
			ctx.JSON(http.StatusNotFound, response{Error: err})
			return db.ErrRollback
		} else if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, response{Result: c.makeAPIApp(app)})
		return nil
	})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (c *Controller) handleAppList(ctx *gin.Context) {
	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		apps, err := conn.ListApps(ctx)
		if err != nil {
			return err
		}

		result := mapModels(apps, c.makeAPIApp)

		ctx.JSON(http.StatusOK, response{Result: result})
		return nil
	})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

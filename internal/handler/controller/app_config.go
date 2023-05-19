package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

func (c *Controller) handleAppConfigGet(ctx *gin.Context) {
	id := ctx.Param("app-id")

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		app, err := conn.GetApp(ctx, id)
		if errors.Is(err, models.ErrAppNotFound) {
			ctx.JSON(http.StatusNotFound, response{Error: err})
			return db.ErrRollback
		} else if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, response{Result: app.Config})
		return nil
	})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (c *Controller) handleAppConfigSet(ctx *gin.Context) {
	id := ctx.Param("app-id")

	var request struct {
		Config *config.AppConfig `json:"config" validate:"required"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	request.Config.SetDefaults()

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		app, err := conn.UpdateAppConfig(ctx, id, request.Config)
		if errors.Is(err, models.ErrAppNotFound) {
			ctx.JSON(http.StatusNotFound, response{Error: err})
			return db.ErrRollback
		} else if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, response{Result: app.Config})
		return nil
	})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

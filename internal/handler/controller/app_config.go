package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
)

func (c *Controller) handleAppConfigGet(ctx *gin.Context) {
	id := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzReadApp(id)) {
		return
	}

	config, err := tx(ctx, c.DB, func(conn db.Conn) (*config.AppConfig, error) {
		app, err := conn.GetApp(ctx, id)
		if err != nil {
			return nil, err
		}
		return app.Config, nil
	})

	writeResponse(ctx, config, err)
}

func (c *Controller) handleAppConfigSet(ctx *gin.Context) {
	id := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzWriteApp(id)) {
		return
	}

	var request struct {
		Config *config.AppConfig `json:"config" binding:"required"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	request.Config.SetDefaults()
	if err := config.ValidateAppConfig(request.Config); err != nil {
		ctx.JSON(http.StatusNotFound, response{Error: err})
		return
	}

	config, err := tx(ctx, c.DB, func(conn db.Conn) (*config.AppConfig, error) {
		app, err := conn.UpdateAppConfig(ctx, id, request.Config)
		if err != nil {
			return nil, err
		}

		return app.Config, nil
	})

	writeResponse(ctx, config, err)
}

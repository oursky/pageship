package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type apiApp struct {
	*models.App
	URL string `json:"url"`
}

func (c *Controller) makeAPIApp(app *models.App) *apiApp {
	return &apiApp{
		App: app,
		URL: c.Config.HostPattern.MakeURL(
			c.Config.HostIDScheme.Make(app.ID, ""),
		),
	}
}

func (c *Controller) handleAppCreate(ctx *gin.Context) {
	if !c.requireAuthn(ctx) {
		return
	}

	var request struct {
		ID string `json:"id" binding:"required,dnsLabel"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	if _, reserved := c.Config.ReservedApps[request.ID]; reserved {
		writeResponse(ctx, nil, models.ErrAppUsedID)
		return
	}

	userID := ctx.GetString(contextUserID)
	app, err := tx(ctx, c.DB, func(conn db.Conn) (*apiApp, error) {
		app := models.NewApp(c.Clock.Now().UTC(), request.ID)

		err := conn.CreateApp(ctx, app)
		if err != nil {
			return nil, err
		}

		err = conn.AssignAppUser(ctx, app.ID, userID)
		if err != nil {
			return nil, err
		}

		return c.makeAPIApp(app), nil
	})

	writeResponse(ctx, app, err)
}

func (c *Controller) handleAppGet(ctx *gin.Context) {
	id := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzReadApp(id)) {
		return
	}

	app, err := tx(ctx, c.DB, func(conn db.Conn) (*apiApp, error) {
		app, err := conn.GetApp(ctx, id)
		if err != nil {
			return nil, err
		}

		return c.makeAPIApp(app), nil
	})

	writeResponse(ctx, app, err)
}

func (c *Controller) handleAppList(ctx *gin.Context) {
	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx) {
		return
	}

	userID := ctx.GetString(contextUserID)
	apps, err := tx(ctx, c.DB, func(conn db.Conn) ([]*apiApp, error) {
		apps, err := conn.ListApps(ctx, userID)
		if err != nil {
			return nil, err
		}

		return mapModels(apps, c.makeAPIApp), nil
	})

	writeResponse(ctx, apps, err)
}

func (c *Controller) handleAppUserList(ctx *gin.Context) {
	id := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzReadApp(id)) {
		return
	}

	users, err := tx(ctx, c.DB, func(conn db.Conn) ([]*apiUser, error) {
		users, err := conn.ListAppUsers(ctx, id)
		if err != nil {
			return nil, err
		}

		return mapModels(users, c.makeAPIUser), nil
	})

	writeResponse(ctx, users, err)
}

func (c *Controller) handleAppUserAdd(ctx *gin.Context) {
	appID := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzWriteApp(appID)) {
		return
	}

	var request struct {
		UserID string `json:"userID" binding:"required"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	result, err := tx(ctx, c.DB, func(conn db.Conn) (*struct{}, error) {
		user, err := conn.GetUser(ctx, request.UserID)
		if err != nil {
			return nil, err
		}

		err = conn.AssignAppUser(ctx, appID, user.ID)
		if err != nil {
			return nil, err
		}

		return &struct{}{}, nil
	})

	writeResponse(ctx, result, err)
}

func (c *Controller) handleAppUserDelete(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	userID := ctx.Param("user-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzWriteApp(appID)) {
		return
	}

	if userID == ctx.GetString(contextUserID) {
		writeResponse(ctx, nil, models.ErrDeleteCurrentUser)
		return
	}

	result, err := tx(ctx, c.DB, func(conn db.Conn) (*struct{}, error) {
		err := conn.UnassignAppUser(ctx, appID, userID)
		if err != nil {
			return nil, err
		}

		return &struct{}{}, nil
	})

	writeResponse(ctx, result, err)
}

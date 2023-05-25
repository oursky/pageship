package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type apiUser struct {
	*models.User
}

func (c *Controller) makeAPIUser(u *models.User) *apiUser {
	return &apiUser{User: u}
}

func (c *Controller) handleMe(ctx *gin.Context) {
	if !c.requireAuthn(ctx) {
		return
	}

	user, err := tx(ctx, c.DB, func(conn db.Conn) (*apiUser, error) {
		user, err := conn.GetUser(ctx, ctx.GetString(contextUserID))
		if err != nil {
			return nil, err
		}

		return c.makeAPIUser(user), nil
	})

	writeResponse(ctx, user, err)
}

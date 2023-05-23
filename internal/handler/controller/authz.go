package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type authzAction interface {
	authz()
}

type authzReadApp string

func (authzReadApp) authz() {}

type authzWriteApp string

func (authzWriteApp) authz() {}

func (c *Controller) requireAuthz(ctx *gin.Context, actions ...authzAction) bool {
	authn_, ok := ctx.Get(contextAuthnInfo)
	if !ok {
		writeResponse(ctx, nil, models.ErrInvalidCredentials)
		return false
	}
	authn := authn_.(*authnInfo)

	err := db.WithTx(ctx, c.DB, func(c db.Conn) error {
		for _, a := range actions {
			switch a := a.(type) {
			case authzReadApp:
				if err := c.IsAppAccessible(ctx, string(a), authn.UserID); err != nil {
					return err
				}
			case authzWriteApp:
				if err := c.IsAppAccessible(ctx, string(a), authn.UserID); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		writeResponse(ctx, nil, err)
		return false
	}

	return true
}

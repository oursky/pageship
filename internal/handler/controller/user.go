package controller

import (
	"net/http"

	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type apiUser struct {
	*models.User
}

func (c *Controller) makeAPIUser(u *models.User) *apiUser {
	return &apiUser{User: u}
}

func (c *Controller) handleMe(w http.ResponseWriter, r *http.Request) {
	userID := authn(r).UserID
	user, err := tx(r.Context(), c.DB, func(conn db.Conn) (*apiUser, error) {
		user, err := conn.GetUser(r.Context(), userID)
		if err != nil {
			return nil, err
		}

		return c.makeAPIUser(user), nil
	})

	writeResponse(w, user, err)
}

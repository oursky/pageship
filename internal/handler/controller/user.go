package controller

import (
	"net/http"

	"github.com/oursky/pageship/internal/models"
)

type apiUser struct {
	*models.User
}

func (c *Controller) makeAPIUser(u *models.User) *apiUser {
	return &apiUser{User: u}
}

func (c *Controller) handleMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	respond(w, func() (any, error) {
		user, err := c.DB.GetUser(r.Context(), userID)
		if err != nil {
			return nil, err
		}

		return c.makeAPIUser(user), nil
	})
}

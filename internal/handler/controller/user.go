package controller

import "github.com/oursky/pageship/internal/models"

type apiUser struct {
	*models.User
}

func (c *Controller) makeAPIUser(u *models.User) *apiUser {
	return &apiUser{User: u}
}

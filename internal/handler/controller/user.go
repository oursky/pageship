package controller

import (
	"net/http"

	"github.com/oursky/pageship/internal/models"
)

type apiUser struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Credentials []models.CredentialID `json:"credentials"`
}

func (c *Controller) handleMe(w http.ResponseWriter, r *http.Request) {
	authn := get[*authnInfo](r)
	writeResponse(w, &apiUser{
		ID:          authn.Subject,
		Name:        authn.Name,
		Credentials: authn.CredentialIDs,
	}, nil)
}

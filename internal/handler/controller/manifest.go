package controller

import (
	"net/http"
)

type apiManifest struct {
	Version             string `json:"version"`
	CustomDomainMessage string `json:"customDomainMessage,omitempty"`
}

func (c *Controller) handleManifest(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, &apiManifest{
		Version:             c.Config.ServerVersion,
		CustomDomainMessage: c.Config.CustomDomainMessage,
	}, nil)
}

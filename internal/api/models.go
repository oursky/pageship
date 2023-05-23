package api

import (
	"github.com/oursky/pageship/internal/models"
)

type APIApp struct {
	*models.App
	URL string `json:"url"`
}

type APISite struct {
	*models.Site
	URL            string  `json:"url"`
	DeploymentName *string `json:"deploymentName"`
}

type APIDeployment struct {
	*models.Deployment
}

type APIUser struct {
	*models.User
}

type SitePatchRequest struct {
	DeploymentName *string `json:"deploymentName,omitempty"`
}

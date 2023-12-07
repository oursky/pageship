package api

import (
	"github.com/oursky/pageship/internal/models"
)

type APIManifest struct {
	Version             string `json:"version"`
	CustomDomainMessage string `json:"customDomainMessage,omitempty"`
}

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
	SiteName *string `json:"siteName"`
	URL      *string `json:"url"`
}

type APIDomain struct {
	Domain             *models.Domain
	DomainVerification *models.DomainVerification
}

type APIUser struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Credentials []models.CredentialID `json:"credentials"`
}

type SitePatchRequest struct {
	DeploymentName *string `json:"deploymentName,omitempty"`
}

package api

import (
	"time"

	"github.com/oursky/pageship/internal/models"
)

type APIApp struct {
	*models.App
	URL string `json:"url"`
}

type APISite struct {
	*models.Site
	URL          string     `json:"url"`
	LastDeployAt *time.Time `json:"uploadedAt" db:"uploaded_at"`
}

type APIDeployment struct {
	*models.Deployment
}

type DeploymentPatchRequest struct {
	Status *models.DeploymentStatus `json:"status,omitempty"`
}

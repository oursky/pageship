package api

import "github.com/oursky/pageship/internal/models"

type DeploymentPatchRequest struct {
	Status *models.DeploymentStatus `json:"status,omitempty"`
}

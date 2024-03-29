package models

import "time"

type Site struct {
	ID           string     `json:"id" db:"id"`
	AppID        string     `json:"appID" db:"app_id"`
	Name         string     `json:"name" db:"name"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt    *time.Time `json:"deletedAt" db:"deleted_at"`
	DeploymentID *string    `json:"deploymentID" db:"deployment_id"`
}

func NewSite(now time.Time, appID string, name string) *Site {
	return &Site{
		ID:           newID("site"),
		AppID:        appID,
		Name:         name,
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil,
		DeploymentID: nil,
	}
}

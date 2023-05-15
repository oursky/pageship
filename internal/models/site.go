package models

import "time"

type Site struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Name               string
	AppID              string
	ActiveDeploymentID string
}

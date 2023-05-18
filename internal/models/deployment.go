package models

import (
	"time"
)

type DeploymentStatus string

const (
	DeploymentStatusPending  DeploymentStatus = "PENDING"
	DeploymentStatusActive   DeploymentStatus = "ACTIVE"
	DeploymentStatusInactive DeploymentStatus = "INACTIVE"
)

type Deployment struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	AppID           string
	EnvironmentName string

	Status           DeploymentStatus
	StorageKeyPrefix string
	Metadata         any
}

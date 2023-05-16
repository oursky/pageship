package models

import (
	"time"

	"github.com/oursky/pageship/internal/config"
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
	Metadata         *DeploymentMetadata
}

type DeploymentMetadata struct {
	Config   *config.ServerConfig
	RootFile *FileMeta
}

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oursky/pageship/internal/config"
)

type DeploymentStatus string

const (
	DeploymentStatusPending  DeploymentStatus = "PENDING"
	DeploymentStatusActive   DeploymentStatus = "ACTIVE"
	DeploymentStatusInactive DeploymentStatus = "INACTIVE"
)

func (s DeploymentStatus) IsValid() bool {
	switch s {
	case DeploymentStatusPending, DeploymentStatusActive, DeploymentStatusInactive:
		return true
	}
	return false
}

type Deployment struct {
	ID        string     `json:"id" db:"id"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time `json:"deletedAt" db:"deleted_at"`

	AppID  string `json:"appID" db:"app_id"`
	SiteID string `json:"siteID" db:"site_id"`

	Status           DeploymentStatus    `json:"status" db:"-"`
	StorageKeyPrefix string              `json:"-" db:"storage_key_prefix"`
	Metadata         *DeploymentMetadata `json:"metadata" db:"metadata"`
	UploadedAt       *time.Time          `json:"uploadedAt" db:"uploaded_at"`
}

func NewDeployment(
	now time.Time,
	appID string,
	siteID string,
	storageKeyPrefix string,
	metadata *DeploymentMetadata,
) *Deployment {
	id := newID("deployment")
	return &Deployment{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
		AppID:     appID,
		SiteID:    siteID,

		Status:           DeploymentStatusPending,
		StorageKeyPrefix: fmt.Sprintf("%s%s/%s/%s", storageKeyPrefix, appID, siteID, id),
		Metadata:         metadata,
		UploadedAt:       nil,
	}
}

func (d *Deployment) SetStatus(siteDeploymentID *string) {
	switch {
	case d.UploadedAt == nil:
		d.Status = DeploymentStatusPending
	case siteDeploymentID != nil && d.ID == *siteDeploymentID:
		d.Status = DeploymentStatusActive
	default:
		d.Status = DeploymentStatusInactive
	}
}

type DeploymentMetadata struct {
	Files  []FileEntry       `json:"files"`
	Config config.SiteConfig `json:"config"`
}

func (m *DeploymentMetadata) Scan(val any) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
func (m *DeploymentMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

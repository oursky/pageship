package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oursky/pageship/internal/config"
)

type Deployment struct {
	ID        string     `json:"id" db:"id"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time `json:"deletedAt" db:"deleted_at"`

	Name  string `json:"name" db:"name"`
	AppID string `json:"appID" db:"app_id"`

	StorageKeyPrefix string              `json:"-" db:"storage_key_prefix"`
	Metadata         *DeploymentMetadata `json:"metadata" db:"metadata"`
	UploadedAt       *time.Time          `json:"uploadedAt" db:"uploaded_at"`
	ExpireAt         *time.Time          `json:"expireAt" db:"expire_at"`
}

func NewDeployment(
	now time.Time,
	name string,
	appID string,
	storageKeyPrefix string,
	metadata *DeploymentMetadata,
) *Deployment {
	id := newID("deployment")
	return &Deployment{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
		Name:      name,
		AppID:     appID,

		StorageKeyPrefix: fmt.Sprintf("%s%s/%s", storageKeyPrefix, appID, id),
		Metadata:         metadata,
		UploadedAt:       nil,
		ExpireAt:         nil,
	}
}

func (d *Deployment) IsExpired(now time.Time) bool {
	return d.ExpireAt != nil && !now.Before(*d.ExpireAt)
}

func (d *Deployment) CheckAlive(now time.Time) error {
	if d.UploadedAt == nil {
		// Not yet uploaded
		return ErrDeploymentNotUploaded
	}
	if d.IsExpired(now) {
		// Expired
		return ErrDeploymentExpired
	}
	return nil
}

type DeploymentMetadata struct {
	Files  []FileEntry       `json:"files,omitempty"`
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

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	ID        string     `json:"id" db:"id"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time `json:"deletedAt" db:"deleted_at"`
	Name      string     `json:"name" db:"name"`
}

func NewUser(now time.Time, name string) *User {
	return &User{
		ID:        newID("user"),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
		Name:      name,
	}
}

type UserCredential struct {
	ID        CredentialID        `json:"id" db:"id"`
	CreatedAt time.Time           `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time           `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time          `json:"deletedAt" db:"deleted_at"`
	UserID    string              `json:"userID" db:"user_id"`
	Data      *UserCredentialData `json:"data" db:"data"`
}

func NewUserCredential(now time.Time, userID string, id CredentialID, data *UserCredentialData) *UserCredential {
	return &UserCredential{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
		UserID:    userID,
		Data:      data,
	}
}

type UserCredentialData struct {
	KeyFingerprint string `json:"keyFingerprint,omitempty"`
}

func (d *UserCredentialData) Scan(val any) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, d)
	case string:
		return json.Unmarshal([]byte(v), d)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

func (d *UserCredentialData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

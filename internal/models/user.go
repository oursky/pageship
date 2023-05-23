package models

import "time"

type User struct {
	ID        string     `json:"id" db:"id"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time `json:"deletedAt" db:"deleted_at"`
	Name      string     `json:"name" db:"name"`
}

type UserCredential struct {
	ID        string              `json:"id" db:"id"`
	CreatedAt time.Time           `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time           `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time          `json:"deletedAt" db:"deleted_at"`
	UserID    string              `json:"userID" db:"user_id"`
	Data      *UserCredentialData `json:"data" db:"data"`
}

type UserCredentialData struct{}

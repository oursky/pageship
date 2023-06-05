package models

import (
	"time"

	"github.com/oursky/pageship/internal/config"
)

type App struct {
	ID          string            `json:"id" db:"id"`
	CreatedAt   time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time         `json:"updatedAt" db:"updated_at"`
	DeletedAt   *time.Time        `json:"deletedAt" db:"deleted_at"`
	OwnerUserID string            `json:"ownerUserID" db:"owner_user_id"`
	Config      *config.AppConfig `json:"config" db:"config"`
}

func NewApp(now time.Time, id string, ownerUserID string) *App {
	config := config.DefaultAppConfig()
	config.SetDefaults()

	return &App{
		ID:          id,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
		OwnerUserID: ownerUserID,
		Config:      &config,
	}
}

func (a *App) CredentialIDs() []UserCredentialID {
	credIDs := []UserCredentialID{UserCredentialUserID(a.OwnerUserID)}
	for _, r := range a.Config.Team {
		if id := UserCredentialIDFromSubject(&r.AccessSubject); id != nil {
			credIDs = append(credIDs, *id)
		}
	}
	return credIDs
}

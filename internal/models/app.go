package models

import (
	"time"

	"github.com/oursky/pageship/internal/config"
)

type App struct {
	ID        string            `json:"id" db:"id"`
	CreatedAt time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time         `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time        `json:"deletedAt" db:"deleted_at"`
	Config    *config.AppConfig `json:"config" db:"config"`
}

func NewApp(id string, now time.Time) *App {
	config := config.DefaultAppConfig()

	return &App{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
		Config:    &config,
	}
}

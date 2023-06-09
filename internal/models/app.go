package models

import (
	"sort"
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

func (a *App) CredentialIndexKeys() []CredentialndexKey {
	m := make(map[CredentialndexKey]struct{})

	collectIndexKeys(m, &config.CredentialMatcher{PageshipUser: a.OwnerUserID})
	for _, r := range a.Config.Team {
		collectIndexKeys(m, &r.CredentialMatcher)
	}

	var keys []CredentialndexKey
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func collectIndexKeys(keys map[CredentialndexKey]struct{}, m *config.CredentialMatcher) {
	for _, k := range MakeCredentialMatcherIndexKeys(m) {
		keys[k] = struct{}{}
	}
}

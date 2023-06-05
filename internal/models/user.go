package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/oursky/pageship/internal/config"
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

type CredentialID string

func CredentialIDFromSubject(s *config.AccessSubject) *CredentialID {
	var id CredentialID
	switch {
	case s.PageshipUser != "":
		id = CredentialUserID(s.PageshipUser)
	case s.GitHubUser != "":
		id = CredentialGitHubUser(s.GitHubUser)
	case s.GitHubRepositoryActions != "":
		id = CredentialGitHubRepositoryActions(s.GitHubRepositoryActions)
	default:
		return nil
	}
	return &id
}

func CredentialUserID(id string) CredentialID {
	return CredentialID(id)
}

func CredentialGitHubUser(username string) CredentialID {
	return CredentialID("github:" + username)
}

func CredentialGitHubRepositoryActions(repo string) CredentialID {
	return CredentialID("github-repo-actions:" + repo)
}

func (i CredentialID) Name() string {
	kind, data, found := strings.Cut(string(i), ":")
	if !found {
		return string(i)
	}

	switch kind {
	case "github":
		name := data
		return name

	default:
		return string(i)
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

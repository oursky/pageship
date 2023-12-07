package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DomainVerification struct {
	ID           string     `json:"id" db:"id"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt    *time.Time `json:"deletedAt" db:"deleted_at"`
	VerifiedAt   *time.Time `json:"verifiedAt" db:"verified_at"`
	Domain       string     `json:"domain" db:"domain"`
	DomainPrefix string     `json:"domainPrefix" db:"domain_prefix"`
	AppID        string     `json:"appID" db:"app_id"`
	Value        string     `json:"value" db:"value"`
}

func NewDomainVerification(now time.Time, domain string, appID string) *DomainVerification {
	uuid, _ := uuid.NewRandom()
	tokens := strings.Split(uuid.String(), "-")
	return &DomainVerification{
		ID:           newID("domain_verification"),
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil,
		Domain:       domain,
		AppID:        appID,
		DomainPrefix: tokens[0],
		Value:        tokens[1],
	}
}

func (d *DomainVerification) GetTxtRecord() (domain string, value string) {
	return fmt.Sprintf("%s._pageship.%s", d.DomainPrefix, d.Domain), d.Value
}

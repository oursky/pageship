package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type DomainVerification struct {
	ID            string     `json:"id" db:"id"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt     *time.Time `json:"deletedAt" db:"deleted_at"`
	VerifiedAt    *time.Time `json:"verifiedAt" db:"verified_at"`
	LastCheckedAt *time.Time `json:"lastCheckedAt" db:"last_checked_at"`
	WillCheckAt   *time.Time `json:"willCheckAt" db:"will_check_at"`
	Domain        string     `json:"domain" db:"domain"`
	DomainPrefix  string     `json:"domainPrefix" db:"domain_prefix"`
	AppID         string     `json:"appID" db:"app_id"`
	Value         string     `json:"value" db:"value"`
}

func NewDomainVerification(now time.Time, domain string, appID string) *DomainVerification {
	num1, err := rand.Int(rand.Reader, big.NewInt(999999))
	if err != nil {
		panic(err)
	}
	num2, err := rand.Int(rand.Reader, big.NewInt(999999))
	if err != nil {
		panic(err)
	}
	return &DomainVerification{
		ID:            newID("domain_verification"),
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     nil,
		Domain:        domain,
		AppID:         appID,
		DomainPrefix:  fmt.Sprintf("%06d", num1),
		Value:         fmt.Sprintf("%06d", num2),
		WillCheckAt:   &now,
		LastCheckedAt: nil,
	}
}

func (d *DomainVerification) GetTxtRecord() (domain string, value string) {
	return fmt.Sprintf("%s._pageship.%s", d.DomainPrefix, d.Domain), d.Value
}

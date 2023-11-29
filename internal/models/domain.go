package models

import "time"

type Domain struct {
	ID        string     `json:"id" db:"id"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time `json:"deletedAt" db:"deleted_at"`
	Domain    string     `json:"domain" db:"domain"`
	AppID     string     `json:"appID" db:"app_id"`
	SiteName  string     `json:"siteName" db:"site_name"`
}

func NewDomain(now time.Time, domain string, appID string, siteName string) *Domain {
	return &Domain{
		ID:        newID("domain"),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
		Domain:    domain,
		AppID:     appID,
		SiteName:  siteName,
	}
}

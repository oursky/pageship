package models

import "time"

type CertDataEntry struct {
	Key       string    `db:"key"`
	UpdatedAt time.Time `db:"updated_at"`
	Value     string    `db:"value"`
}

func NewCertDataEntry(key string, value string, now time.Time) *CertDataEntry {
	return &CertDataEntry{
		Key:       key,
		UpdatedAt: now,
		Value:     value,
	}
}

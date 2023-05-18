package models

import "time"

type Site struct {
	AppID     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

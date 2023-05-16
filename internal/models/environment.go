package models

import "time"

type Environment struct {
	AppID     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

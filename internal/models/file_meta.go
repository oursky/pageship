package models

import (
	"time"
)

type FileMeta struct {
	Name    string
	Size    int64
	IsDir   bool
	ModTime time.Time
	Files   []FileMeta

	ETag string
}

package models

const MaxFiles = 10000

type FileEntry struct {
	Path string `json:"path" validate:"required,max=256"` // Directory has trailing slash in path
	Size int64  `json:"size" validate:"required,gte=0"`
	Hash string `json:"hash" validate:"required,max=100"`
}

package models

const MaxFiles = 10000

type FileEntry struct {
	FilePath string `json:"-"`
	Path     string `json:"path"` // Directory has trailing slash in path
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
}

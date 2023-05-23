package models

import (
	"mime"
	"path"

	"github.com/h2non/filetype"
)

const MaxFiles = 10000

type FileEntry struct {
	Path        string `json:"path" validate:"required,max=256"` // Directory has trailing slash in path
	Size        int64  `json:"size" validate:"required,gte=0"`
	Hash        string `json:"hash" validate:"required,max=100"`
	ContentType string `json:"contentType" validate:"required,max=100"`
}

func DetectContentType(fileName string, initialBytes []byte) string {
	magic, err := filetype.Match(initialBytes)
	if err == nil && magic != filetype.Unknown {
		return magic.MIME.Value
	}

	exttype := mime.TypeByExtension(path.Ext(fileName))
	if exttype != "" {
		return exttype
	}

	return "application/octet-stream"
}

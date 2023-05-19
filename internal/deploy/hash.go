package deploy

import (
	"encoding/base64"
	"hash"

	"golang.org/x/crypto/sha3"
)

type FileHash struct {
	hash hash.Hash
}

func NewFileHash() *FileHash {
	return &FileHash{hash: sha3.New256()}
}

func (h *FileHash) Write(p []byte) (int, error) {
	return h.hash.Write(p)
}

func (h *FileHash) Sum() string {
	return base64.RawURLEncoding.EncodeToString(h.hash.Sum(nil))
}

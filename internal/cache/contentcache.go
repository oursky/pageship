package cache

import (
	"io"
	"sync"

	"github.com/dgraph-io/ristretto"
)

const (
	contentCacheSize uint64 = 1000000000
)

type ContentCache[T io.WriteCloser] struct {
	m     map[string]sync.Mutex
	cache *ristretto.Cache
	load  func(id string) (T, error)
}

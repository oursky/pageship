package cache

import (
	"io"
	"sync"

	"github.com/dgraph-io/ristretto"
)

const (
	contentCacheSize int64 = 1000000000
)

type ContentCache[T io.WriteCloser] struct {
	m     map[string]sync.Mutex
	cache *ristretto.Cache
	load  func(id string) (T, error)
}

func NewContentCache[T io.WriteCloser](load func(id string) (T, error)) (*ContentCache[T], error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     contentCacheSize,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &ContentCache[T]{cache: cache, load: load}, nil
}

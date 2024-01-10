package cache

import (
	"bytes"
	"sync"

	"github.com/dgraph-io/ristretto"
)

const (
	contentCacheSize int64 = 1000000000
)

type ContentCache struct {
	m     map[string]*sync.Mutex
	cache *ristretto.Cache
	load  func(id string) (*bytes.Buffer, error)
}

func NewContentCache(load func(id string) (*bytes.Buffer, error)) (*ContentCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     contentCacheSize,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &ContentCache{m: make(map[string]*sync.Mutex), cache: cache, load: load}, nil
}

func (c *ContentCache) getContent(id string) (*bytes.Buffer, error) {
	c.m[id].Lock()
	defer c.m[id].Unlock()

	temp, found := c.cache.Get(id)
	ce := temp.(*bytes.Buffer)
	if found {
		return ce, nil
	}

	ce, err := c.load(id)
	if err != nil {
		return bytes.NewBuffer([]byte{}), err
	}

	c.cache.Set(id, ce, int64(ce.Len()))
	return ce, nil
}

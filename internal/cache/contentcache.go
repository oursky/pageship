package cache

import (
	"io"
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
}

type contentCacheCell struct {
	hash	string
	data	*bytes.Buffer
}

func NewContentCache() (*ContentCache, error) {
	m := make(map[string]*sync.Mutex)
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     contentCacheSize,
		BufferItems: 64,
		OnExit: func(item interface{}) {
			cell := item.(contentCacheCell)
			delete(m, cell.hash)
		},
	})
	if err != nil {
		return nil, err
	}

	return &ContentCache{m: m, cache: cache}, nil
}

func (c *ContentCache) getContent(id string, r io.Reader) (*bytes.Buffer, error) {
	(*c).m[id].Lock()
	defer c.m[id].Unlock()

	temp, found := c.cache.Get(id)
	ce := temp.(contentCacheCell)
	if found {
		return ce.data, nil
	}
	
	b := make([]byte, contentCacheSize)
	_, err := r.Read(b)
	data := bytes.NewBuffer(b)
	if err != io.EOF {
		return data, err
	}

	ce = contentCacheCell {
		hash: id,
		data: data,
	}
	c.cache.Set(id, ce, int64(ce.data.Len()))
	return ce.data, nil
}

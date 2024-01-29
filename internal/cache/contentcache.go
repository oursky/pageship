package cache

import (
	"bytes"
	"io"
	"sync"

	"github.com/dgraph-io/ristretto"
)

type ContentCache struct {
	m     sync.Map
	size  int64
	cache *ristretto.Cache
}

type ContentCacheCell struct {
	id   string
	Data *bytes.Buffer
}

func NewContentCache(contentCacheSize int64, metrics bool) (*ContentCache, error) {
	var m sync.Map
	size := contentCacheSize
	cache, err := ristretto.NewCache(&ristretto.Config{
		//NumCounters is 10 times estimated max number of items in cache, as suggested in https://pkg.go.dev/github.com/dgraph-io/ristretto@v0.1.1#Config
		NumCounters: 16000, //16 MB limit / 10 KB small files = 1600 max number of items
		MaxCost:     size,
		BufferItems: 64,
		Metrics:     metrics,
		OnExit: func(item interface{}) {
			cell := item.(ContentCacheCell)
			m.Delete(cell.id)
		},
		IgnoreInternalCost: true,
	})
	if err != nil {
		return nil, err
	}

	return &ContentCache{m: m, size: size, cache: cache}, nil
}

func (c *ContentCache) GetContent(id string, r io.ReadSeeker) (ContentCacheCell, error) {
	m, _ := c.m.LoadOrStore(id, new(sync.Mutex))
	mu := m.(*sync.Mutex)
	mu.Lock()
	defer func() {
		m, _ := c.m.LoadOrStore(id, new(sync.Mutex))
		mu := m.(*sync.Mutex)
		mu.Unlock()
	}()

	temp, found := c.cache.Get(id)
	if found {
		ce := temp.(ContentCacheCell)
		return ce, nil
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return ContentCacheCell{id: "", Data: new(bytes.Buffer)}, err
	}

	ce := ContentCacheCell{
		id:   id,
		Data: bytes.NewBuffer(b),
	}
	c.cache.Set(id, ce, int64(ce.Data.Len()))
	return ce, nil
}

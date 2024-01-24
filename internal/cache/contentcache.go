package cache

import (
	"bytes"
	"io"
	"sync"

	"github.com/dgraph-io/ristretto"
)

type ContentCache struct {
	mm    *sync.Mutex
	m     map[string]*sync.Mutex
	size  int64
	cache *ristretto.Cache
}

type ContentCacheCell struct {
	hash string
	Data *bytes.Buffer
}

func NewContentCache(contentCacheSize int64) (*ContentCache, error) {
	mm := new(sync.Mutex)
	m := make(map[string]*sync.Mutex)
	size := contentCacheSize
	cache, err := ristretto.NewCache(&ristretto.Config{
		//NumCounters is 10 times estimated max number of items in cache, as suggested in https://pkg.go.dev/github.com/dgraph-io/ristretto@v0.1.1#Config
		NumCounters: 1e7,
		MaxCost:     size,
		BufferItems: 64,
		OnExit: func(item interface{}) {
			mm.Lock()
			defer mm.Unlock()
			cell := item.(ContentCacheCell)
			delete(m, cell.hash)
		},
	})
	if err != nil {
		return nil, err
	}

	return &ContentCache{mm: mm, m: m, size: size, cache: cache}, nil
}

func (c *ContentCache) GetContent(id string, r io.Reader) (ContentCacheCell, error) {
	c.mm.Lock()
	c.m[id].Lock()
	c.mm.Unlock()
	defer func() {
		c.mm.Lock()
		c.m[id].Unlock()
		c.mm.Unlock()
	}()

	temp, found := c.cache.Get(id)
	ce := temp.(ContentCacheCell)
	if found {
		return ce, nil
	}

	b := make([]byte, c.size)
	_, err := r.Read(b)
	data := bytes.NewBuffer(b)
	if err != io.EOF {
		return ContentCacheCell{hash: "", Data: new(bytes.Buffer)}, err
	}

	ce = ContentCacheCell{
		hash: id,
		Data: data,
	}
	c.cache.Set(id, ce, int64(ce.Data.Len()))
	return ce, nil
}

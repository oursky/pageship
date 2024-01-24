package cache

import (
	"bytes"
	"io"
	"sync"
	"fmt"

	"github.com/dgraph-io/ristretto"
)

type ContentCache struct {
	mm    *sync.Mutex
	m     map[string]*sync.Mutex
	size  int64
	cache *ristretto.Cache
}

type ContentCacheCell struct {
	id   string
	Data *bytes.Buffer
}

func NewContentCache(contentCacheSize int64, metrics bool) (*ContentCache, error) {
	mm := new(sync.Mutex)
	m := make(map[string]*sync.Mutex)
	size := contentCacheSize
	cache, err := ristretto.NewCache(&ristretto.Config{
		//NumCounters is 10 times estimated max number of items in cache, as suggested in https://pkg.go.dev/github.com/dgraph-io/ristretto@v0.1.1#Config
		NumCounters: 1e7,
		MaxCost:     size,
		BufferItems: 64,
		Metrics:     metrics,
		OnExit: func(item interface{}) {
			mm.Lock()
			defer mm.Unlock()
			cell := item.(ContentCacheCell)
			delete(m, cell.id)
		},
	})
	if err != nil {
		return nil, err
	}

	return &ContentCache{mm: mm, m: m, size: size, cache: cache}, nil
}

func (c *ContentCache) GetContent(id string, r io.Reader) (ContentCacheCell, error) {
	c.mm.Lock()
	if m, ok := c.m[id]; ok {
		m.Lock()
	} else {
		c.m[id] = new(sync.Mutex)
		c.m[id].Lock()
	}
	c.mm.Unlock()
	defer func() {
		c.mm.Lock()
		if m, ok := c.m[id]; ok {
			m.Unlock()
		} else {
			c.m[id] = new(sync.Mutex)
			c.m[id].Unlock()
		}
		c.mm.Unlock()
	}()

	temp, found := c.cache.Get(id)
	if found {
		ce := temp.(ContentCacheCell)
		return ce, nil
	}

	b := make([]byte, c.size)
	_, err := r.Read(b)
	if err != nil { //io.EOF?
		return ContentCacheCell{id: "", Data: new(bytes.Buffer)}, err
	}

	ce := ContentCacheCell{
		id:   id,
		Data: bytes.NewBuffer(bytes.Trim(b, string([]byte{0x0}))),
	}
	fmt.Printf("setting... %d\n", ce.Data.Len())
	c.cache.Set(id, ce, int64(ce.Data.Len()))
	fmt.Printf("set\n")
	return ce, nil
}

package cache

import (
	"net/http"

	"github.com/dgraph-io/ristretto"
)

type ContentCache struct {
	size int64
	*ristretto.Cache
	debug bool
}

type Response struct {
	Header     http.Header
	Body       []byte
	StatusCode int
}

func NewContentCache(contentCacheSize int64, debug bool) (*ContentCache, error) {
	size := contentCacheSize
	nc := size / 1000
	cache, err := ristretto.NewCache(&ristretto.Config{
		//NumCounters is 10 times estimated max number of items in cache, as suggested in https://pkg.go.dev/github.com/dgraph-io/ristretto@v0.1.1#Config
		NumCounters:        nc, //limit / 10 KB small files * 10
		MaxCost:            size,
		BufferItems:        64,
		Metrics:            debug,
		IgnoreInternalCost: true,
	})
	if err != nil {
		return nil, err
	}

	return &ContentCache{size, cache, debug}, nil
}

func (c *ContentCache) Get(key string) (*Response, bool) {
	r, b := c.Cache.Get(key)
	return r.(*Response), b
}

func (c *ContentCache) Set(key string, value *Response) {
	c.Cache.Set(key, value, int64(len(value.Body)))
}

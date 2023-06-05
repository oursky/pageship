package cache

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/simplelru"
)

const (
	cacheSize int           = 100
	cacheTTL  time.Duration = time.Second * 1
)

type Cache[T any] struct {
	m     sync.Mutex
	ttl   time.Duration
	cache *simplelru.LRU[string, *TTLCell[T]]
	load  func(id string) (T, error)
}

func NewCache[T any](size int, ttl time.Duration, load func(id string) (T, error)) (*Cache[T], error) {
	cache, err := simplelru.NewLRU[string, *TTLCell[T]](cacheSize, nil)
	if err != nil {
		return nil, err
	}

	return &Cache[T]{cache: cache, ttl: ttl, load: load}, nil
}

func (c *Cache[T]) getCell(id string) *TTLCell[T] {
	c.m.Lock()
	defer c.m.Unlock()

	ce, ok := c.cache.Get(id)
	if ok {
		return ce
	}

	ce = NewTTLCell(id, c.ttl, c.load)
	c.cache.Add(id, ce)
	return ce
}

func (c *Cache[T]) Load(id string) (T, error) {
	cell := c.getCell(id)
	return cell.Load()
}

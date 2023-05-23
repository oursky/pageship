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
	cache *simplelru.LRU[string, *cell[T]]
}

func NewCache[T any](size int, ttl time.Duration) (*Cache[T], error) {
	cache, err := simplelru.NewLRU[string, *cell[T]](cacheSize, nil)
	if err != nil {
		return nil, err
	}

	return &Cache[T]{cache: cache, ttl: ttl}, nil
}

func (c *Cache[T]) getCell(id string) *cell[T] {
	c.m.Lock()
	defer c.m.Unlock()

	ce, ok := c.cache.Get(id)
	if ok {
		return ce
	}

	ce = new(cell[T])
	c.cache.Add(id, ce)
	return ce
}

func (c *Cache[T]) Load(
	id string,
	load func(id string) (*T, error),
) (*T, error) {
	cell := c.getCell(id)
	return cell.Load(func() (*T, error) { return load(id) })
}

type cell[T any] struct {
	m        sync.RWMutex
	expireAt time.Time
	value    *T
	err      error
}

func (c *cell[T]) Load(fn func() (*T, error)) (*T, error) {
	if value, err, ok := c.loadCached(); ok {
		return value, err
	}
	return c.loadNew(fn)
}

func (c *cell[T]) checkCachedValue() (*T, error, bool) {
	if c.value == nil {
		return nil, nil, false
	}
	if time.Now().After(c.expireAt) {
		return nil, nil, false
	}
	return c.value, c.err, true
}

func (c *cell[T]) loadCached() (*T, error, bool) {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.checkCachedValue()
}

func (c *cell[T]) loadNew(fn func() (*T, error)) (*T, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if value, err, ok := c.checkCachedValue(); ok {
		return value, err
	}

	c.value, c.err = fn()
	c.expireAt = time.Now().Add(cacheTTL)
	return c.value, c.err
}

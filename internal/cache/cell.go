package cache

import (
	"sync"
	"time"
)

type TTLCell[T any] struct {
	id   string
	ttl  time.Duration
	load func(id string) (T, error)

	m        sync.RWMutex
	expireAt time.Time
	loaded   bool
	value    T
	err      error
}

func NewTTLCell[T any](
	id string,
	ttl time.Duration,
	load func(id string) (T, error),
) *TTLCell[T] {
	return &TTLCell[T]{id: id, ttl: ttl, load: load}
}

func (c *TTLCell[T]) Load() (T, error) {
	if value, err, ok := c.loadCached(); ok {
		return value, err
	}
	return c.loadNew()
}

func (c *TTLCell[T]) checkCachedValue() (value T, err error, ok bool) {
	if !c.loaded {
		return
	}
	if time.Now().After(c.expireAt) {
		return
	}
	return c.value, c.err, true
}

func (c *TTLCell[T]) loadCached() (T, error, bool) {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.checkCachedValue()
}

func (c *TTLCell[T]) loadNew() (T, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if value, err, ok := c.checkCachedValue(); ok {
		return value, err
	}

	c.value, c.err = c.load(c.id)
	c.loaded = true
	c.expireAt = time.Now().Add(cacheTTL)
	return c.value, c.err
}

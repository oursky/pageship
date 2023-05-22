package site

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/simplelru"
)

const (
	cacheSize int           = 100
	cacheTTL  time.Duration = time.Second * 1
)

type siteCache struct {
	m     sync.Mutex
	cache *simplelru.LRU[string, *siteCacheCell]
}

func newSiteCache() (*siteCache, error) {
	cache, err := simplelru.NewLRU[string, *siteCacheCell](cacheSize, nil)
	if err != nil {
		return nil, err
	}

	return &siteCache{cache: cache}, nil
}

func (c *siteCache) getCell(matchedID string) *siteCacheCell {
	c.m.Lock()
	defer c.m.Unlock()

	cell, ok := c.cache.Get(matchedID)
	if ok {
		return cell
	}

	cell = new(siteCacheCell)
	c.cache.Add(matchedID, cell)
	return cell
}

func (c *siteCache) Load(
	matchedID string,
	load func(matchedID string) (*Descriptor, error),
) (*Descriptor, error) {
	cell := c.getCell(matchedID)
	return cell.Load(func() (*Descriptor, error) { return load(matchedID) })
}

type siteCacheCell struct {
	m        sync.RWMutex
	expireAt time.Time
	value    *Descriptor
	err      error
}

func (c *siteCacheCell) Load(fn func() (*Descriptor, error)) (*Descriptor, error) {
	if value, err, ok := c.loadCached(); ok {
		return value, err
	}
	return c.loadNew(fn)
}

func (c *siteCacheCell) checkCachedValue() (*Descriptor, error, bool) {
	if c.value == nil {
		return nil, nil, false
	}
	if time.Now().After(c.expireAt) {
		return nil, nil, false
	}
	return c.value, c.err, true
}

func (c *siteCacheCell) loadCached() (*Descriptor, error, bool) {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.checkCachedValue()
}

func (c *siteCacheCell) loadNew(fn func() (*Descriptor, error)) (*Descriptor, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if value, err, ok := c.checkCachedValue(); ok {
		return value, err
	}

	c.value, c.err = fn()
	c.expireAt = time.Now().Add(cacheTTL)
	return c.value, c.err
}

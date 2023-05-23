package site

import (
	"time"

	"github.com/oursky/pageship/internal/cache"
)

const (
	cacheSize int           = 100
	cacheTTL  time.Duration = time.Second * 1
)

func newSiteCache() (*cache.Cache[Descriptor], error) {
	return cache.NewCache[Descriptor](cacheSize, cacheTTL)
}

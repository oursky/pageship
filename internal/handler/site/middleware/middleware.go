package middleware

import (
	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/handler/site"
)

func Default(cc *cache.ContentCache) []site.Middleware {
	cacheContext := NewCacheContext(cc)

	return []site.Middleware{
		RedirectCustomDomain,
		CanonicalizePath,
		RouteSPA,
		IndexPage,
		cacheContext.Cache,
	}
}

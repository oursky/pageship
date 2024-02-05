package middleware

import "github.com/oursky/pageship/internal/handler/site"

func Default(cc ContentCacheType) []site.Middleware{
	cacheContext := NewCacheContext(cc)

	return []site.Middleware{
		RedirectCustomDomain,
		CanonicalizePath,
		RouteSPA,
		IndexPage,
		cacheContext.Cache,
	}
}

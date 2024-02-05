package middleware

import (
	"net/http"

	"github.com/oursky/pageship/internal/site"
)

type Middleware func(*site.Descriptor, http.Handler) http.Handler

func ApplyMiddleware(site *site.Descriptor, middlewares []Middleware, handler http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](site, handler)
	}
	return handler
}

func Default(cc ContentCacheType) []Middleware{
	cacheContext := NewCacheContext(cc)

	return []Middleware{
		RedirectCustomDomain,
		CanonicalizePath,
		RouteSPA,
		IndexPage,
		cacheContext.Cache,
	}
}

package site

import (
	"net/http"

	"github.com/oursky/pageship/internal/site"
)

type Middleware func(site.FS, http.Handler) http.Handler

func applyMiddleware(fs site.FS, middlewares []Middleware, handler http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](fs, handler)
	}
	return handler
}

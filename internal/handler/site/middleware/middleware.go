package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/handler/site"
)

var Default = []site.Middleware{
	RedirectCustomDomain,
	CanonicalizePath,
	RouteSPA,
	IndexPage,
	cm.Compression, //need to create cm
}

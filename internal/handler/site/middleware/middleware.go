package middleware

import "github.com/oursky/pageship/internal/handler/site"

var Default = []site.Middleware{
	CanonicalizePath,
	RouteSPA,
	IndexPage,
}

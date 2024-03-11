package middleware

import (
	"github.com/oursky/pageship/internal/handler/site"
)

var Default = []site.Middleware{
	RedirectCustomDomain,
	CanonicalizePath,
	RouteSPA,
	IndexPage,
	NotFound,
	compression,
}

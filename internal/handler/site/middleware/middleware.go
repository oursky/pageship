package middleware

import (
	"github.com/oursky/pageship/internal/handler/site"
)

var Default = []site.Middleware{
	RedirectCustomDomain,
	CanonicalizePath,
	RouteSPA,
	IndexPage,
	compression,
}

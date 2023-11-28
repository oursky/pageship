package middleware

import (
	"net/http"

	"github.com/oursky/pageship/internal/site"
)

func RedirectCustomDomain(site *site.Descriptor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if site.Domain != "" && r.Host != site.Domain {
			// Site with custom domain must be accessed through the custom domain.
			url := *r.URL
			url.Host = site.Domain
			http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

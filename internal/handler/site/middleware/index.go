package middleware

import (
	"net/http"

	"github.com/oursky/pageship/internal/site"
)

func IndexPage(fs site.FS, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const indexPage = "index.html"

		info, err := fs.Stat(r.URL.Path)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if info.IsDir {
			r.URL.Path = r.URL.Path + indexPage
		}
		next.ServeHTTP(w, r)
	})
}

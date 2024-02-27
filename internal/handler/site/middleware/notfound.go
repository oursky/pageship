package middleware

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/oursky/pageship/internal/site"
)

func NotFound(site *site.Descriptor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const NotFoundPage = "404.html"
		notFound := false

		for {
			fmt.Println(" ", r.URL.Path)
			_, err := site.FS.Stat(r.URL.Path)
			if os.IsNotExist(err) {
				notFound = true
				if path.Dir(r.URL.Path) == "/" && path.Base(r.URL.Path) == NotFoundPage {
					http.NotFound(w, r)
					return
				}
				if path.Base(r.URL.Path) == NotFoundPage {
					r.URL.Path = path.Join(path.Dir(path.Dir(r.URL.Path)), NotFoundPage)
				} else {
					r.URL.Path = path.Join(path.Dir(r.URL.Path), NotFoundPage)
				}
			} else {
				break
			}
		}

		if notFound {
			w.WriteHeader(404)
		}
		next.ServeHTTP(w, r)
	})
}

package middleware

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"time"

	internalhttputil "github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/site"
)

const NotFoundPage = "404.html"

func NotFound(site *site.Descriptor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFound := false

		for {
			_, err := site.FS.Stat(r.URL.Path)
			if errors.Is(err, fs.ErrNotExist) {
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
			writer := internalhttputil.NewTimeoutResponseWriter(w, 10*time.Second)
			rsc, _ := site.FS.Open(r.Context(), r.URL.Path)
			b, _ := io.ReadAll(rsc)
			writer.Write(b)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

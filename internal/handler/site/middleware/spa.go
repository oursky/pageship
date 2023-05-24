package middleware

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/oursky/pageship/internal/site"
)

// RouteSPA routes non-existing files to nearest parent directory
func RouteSPA(fs site.FS, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlpath := r.URL.Path
		for {
			_, err := fs.Stat(urlpath)
			if os.IsNotExist(err) {
				if urlpath == "/" {
					// Reached root; stop
					break
				}
				urlpath = path.Dir("/" + strings.TrimSuffix(urlpath, "/"))
				if urlpath != "/" {
					urlpath += "/"
				}
				continue
			} else if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			break
		}
		r.URL.Path = urlpath
		next.ServeHTTP(w, r)
	})
}

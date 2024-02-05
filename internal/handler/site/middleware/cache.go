package middleware

import (
	"bytes"
	"net/http"
	"path"
	"time"
	"net/http/httptest"
	"io"

	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/site"
)

type ContentCacheKey struct {
	hash        string
	compression string
}

type ContentCacheType = *cache.ContentCache[ContentCacheKey, *bytes.Buffer, io.ReadSeeker]

type CacheContext struct {
	cc ContentCacheType
}

func NewCacheContext(cc ContentCacheType) CacheContext {
	return CacheContext{cc: cc}
}

func (ctx *CacheContext) Cache(site *site.Descriptor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info, err := site.FS.Stat(r.URL.Path)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		compression := "no"
		if contains(r.Header["Accept-Encoding"], "*") || contains(r.Header["Accept-Encoding"], "br") {
			compression := "br"
		} else if contains(r.Header["Accept-Encoding"], "gzip") {
			compression := "gz"
		}

		value, found := ctx.cc.GetContent(ContentCacheKey{hash: info.Hash, compression: compression})
		if found {
			reader := bytes.NewReader(value.Bytes())
			writer := httputil.NewTimeoutResponseWriter(w, 10*time.Second)
			http.ServeContent(writer, r, path.Base(r.URL.Path), info.ModTime, reader)
			return
		}

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)
		reader := bytes.NewReader(rec.Body.Bytes())

		compression = "no"
		if contains(rec.Header["Content-Encoding"], "*") || contains(rec.Header["Content-Encoding"], "br") {
			compression = "br"
		} else if contains(rec.Header["Content-Encoding"], "gzip") {
			compression = "gz"
		}

		ctx.cc.SetContent(ContentCacheKey{hash: info.Hash, compression: compression}, reader)
		writer := httputil.NewTimeoutResponseWriter(w, 10*time.Second)
		http.ServeContent(writer, r, path.Base(r.URL.Path), info.ModTime, reader)
	})
}

//https://stackoverflow.com/questions/10485743/contains-method-for-a-slice
func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

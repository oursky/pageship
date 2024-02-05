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

		value, found := ctx.cc.GetContent(ContentCacheKey{hash: info.Hash, compression: "none"}) //TODO: find compression type
		if found {
			reader := bytes.NewReader(value.Bytes())
			writer := httputil.NewTimeoutResponseWriter(w, 10*time.Second)
			http.ServeContent(writer, r, path.Base(r.URL.Path), info.ModTime, reader)
			return
		}

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		reader := bytes.NewReader(rec.Body.Bytes())
		ctx.cc.SetContent(ContentCacheKey{hash: info.Hash, compression: "none"}, reader)
		writer := httputil.NewTimeoutResponseWriter(w, 10*time.Second)
		http.ServeContent(writer, r, path.Base(r.URL.Path), info.ModTime, reader)
	})
}


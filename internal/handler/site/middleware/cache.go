package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"time"
	"fmt"

	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/site"
	"github.com/dgraph-io/ristretto"
	"github.com/go-chi/chi/v5/middleware"
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
	return CacheContext{cc: cc, vc: varyCache}
}

func (ctx *CacheContext) Cache(site *site.Descriptor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info, err := site.FS.Stat(r.URL.Path)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		
		keyString := fmt.sprintf("%v", ContentCacheKey{hash: info.Hash, compression: r.Header().Get("Accept-Encoding"))
		value, found := ctx.cc.Get(keyString)
		if found {
			for k, v := range(value.Header) {
				w.Header()[k] = v
			}
			w.WriteHeader(value.StatusCode)
			writer := httputil.NewTimeoutResponseWriter(w, 10*time.Second)
			writer.Write(value.Body)
			return
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		b := new(bytes.Buffer)
		ww.Tee(b)

		next.ServeHTTP(ww, r)

		value := cache.Response { Header: ww.Header(), Body: b.Bytes(), StatusCode: ww.Status() }
		ctx.cc.Set(keystring, value)
	})
}

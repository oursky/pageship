package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/site"
)

type ContentCacheKey struct {
	hash        string
	compression string
}

type CacheContext struct {
	cc *cache.ContentCache
}

func NewCacheContext(cc *cache.ContentCache) CacheContext {
	return CacheContext{cc: cc}
}

func (ctx *CacheContext) Cache(site *site.Descriptor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info, err := site.FS.Stat(r.URL.Path)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		key := ContentCacheKey{hash: info.Hash, compression: r.Header.Get("Accept-Encoding")}
		keyString := fmt.Sprintf("%s,%s", key.hash, key.compression)
		value, found := ctx.cc.Get(keyString)
		if found {
			for k, v := range value.Header {
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

		setValue := cache.Response{Header: ww.Header(), Body: b.Bytes(), StatusCode: ww.Status()}
		ctx.cc.Set(keyString, &setValue)
	})
}

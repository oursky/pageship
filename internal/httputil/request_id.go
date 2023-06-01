package httputil

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := makeID()
		r = r.WithContext(context.WithValue(r.Context(), middleware.RequestIDKey, id))
		w.Header().Set(middleware.RequestIDHeader, id)
		next.ServeHTTP(w, r)
	})
}

var idEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func makeID() string {
	var b [12]byte
	binary.BigEndian.PutUint64(b[:], uint64(time.Now().UnixMilli()))
	_, _ = rand.Read(b[8:])

	return idEncoding.EncodeToString(b[:])
}

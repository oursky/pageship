package middleware

import (
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/site"
)

func Compression(site *site.Descriptor, next http.Handler) http.Handler {
	c := middleware.NewCompressor(5)
	c.SetEncoder("br", func(w io.Writer, level int) io.Writer {
		return brotli.NewWriterV2(w, level)
	})
	return c.Handler(next)
}

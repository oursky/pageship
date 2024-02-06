package middleware

import (
	"net/http"
	"io"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/site"
)

type CompressionMiddleware struct {
	compressor middleware.Compressor
}

func NewCompressionMiddleware() CompressionMiddleware {
	return compressionMiddleware{compressor: NewCompressor(5)}
}

func (cm *compressionMiddleware) Compression(site *site.Descriptor, next http.Handler) http.Handler {
	return cm.compressor.Handler(next)
}

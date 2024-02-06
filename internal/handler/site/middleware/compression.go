package middleware

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/site"
)

type CompressionMiddleware struct {
	compressor middleware.Compressor
}

func NewCompressionMiddleware() CompressionMiddleware {
	return CompressionMiddleware{compressor: NewCompressor(5)}
}

func (cm *CompressionMiddleware) Compression(site *site.Descriptor, next http.Handler) http.Handler {
	return cm.compressor.Handler(next)
}

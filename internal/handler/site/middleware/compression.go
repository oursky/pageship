package middleware

import (
	"net/http"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/oursky/pageship/internal/site"
)

type CompressedResponseWriter struct {
	http.ResponseWriter
	compressor io.WriteCloser
}

func (crw CompressedResponseWriter) Write(b []byte) (int, error) {
	defer crw.compressor.Close()
	return crw.compressor.Write(b)
}

func Compression(site *site.Descriptor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		crw := CompressedResponseWriter{w, brotli.HTTPCompressor(w, r)} //chooses compression method based on Accept-Encoding header and
	                                                                    //also sets the Content-Encoding header
		next.ServeHTTP(crw, r)
	})
}


package middleware_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/stretchr/testify/assert"
)

type mockHandler struct {
	executeCount int
}

func (mh *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mh.executeCount++
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte("hello"))
}

func TestCacheGzip(t *testing.T) {
	//Setup
	h := middleware.Compression(new(mockHandler))

	//Act Assert
	req, err := http.NewRequest("GET", "endpoint", nil)
	assert.Empty(t, err)
	req.Header.Add("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	resp := rec.Result()
	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	gzreader, _ := gzip.NewReader(resp.Body)
	defer gzreader.Close()
	b, err := io.ReadAll(gzreader)

	assert.Empty(t, err)
	assert.Equal(t, "hello", string(b))
}

func TestCacheBrotli(t *testing.T) {
	//Setup
	h := middleware.Compression(new(mockHandler))

	//Act Assert
	req, err := http.NewRequest("GET", "endpoint", nil)
	assert.Empty(t, err)
	req.Header.Add("Accept-Encoding", "br")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	resp := rec.Result()
	assert.Equal(t, "br", resp.Header.Get("Content-Encoding"))

	brreader := brotli.NewReader(resp.Body)
	defer resp.Body.Close()
	b, err := io.ReadAll(brreader)

	assert.Empty(t, err)
	assert.Equal(t, "hello", string(b))
}

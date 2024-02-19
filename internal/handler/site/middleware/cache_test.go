package middleware_test //black box testing

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/site"
	"github.com/stretchr/testify/assert"
)

type mockHandler struct {
	executeCount *int
}

func (mh mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	(*mh.executeCount)++
	w.WriteHeader(200)
	w.Write([]byte("hello"))
}

type mockFS struct{}

func (m mockFS) Stat(path string) (*site.FileInfo, error) {
	return &site.FileInfo{
		IsDir:       false,
		ModTime:     time.Now(),
		Size:        0,
		ContentType: "",
		Hash:        "",
	}, nil
}

type mockRSCloser struct {
	io.ReadSeeker
}

func (m mockRSCloser) Close() error {
	return nil
}

func (m mockFS) Open(ctx context.Context, path string) (io.ReadSeekCloser, error) {
	return mockRSCloser{bytes.NewReader([]byte{})}, nil
}

func TestCache(t *testing.T) {
	//Setup
	executeCount := 0
	mh := mockHandler{executeCount: &executeCount}
	contentCache, err := cache.NewContentCache(1<<24, true)
	assert.Empty(t, err)
	cacheContext := middleware.NewCacheContext(contentCache)

	sc := config.DefaultSiteConfig()
	mockSiteDescriptor := site.Descriptor{
		ID:     "",
		Domain: "",
		Config: &sc,
		FS:     mockFS{},
	}
	h := cacheContext.Cache(&mockSiteDescriptor, mh)

	//Act Assert
	req, err := http.NewRequest("GET", "endpoint", nil)
	assert.Empty(t, err)
	rec := httptest.NewRecorder()
	assert.Equal(t, 0, executeCount)
	h.ServeHTTP(rec, req)
	assert.Equal(t, 1, executeCount)
	h.ServeHTTP(rec, req)
	assert.Equal(t, 1, executeCount)
	h.ServeHTTP(rec, req)
	assert.Equal(t, 1, executeCount)
}

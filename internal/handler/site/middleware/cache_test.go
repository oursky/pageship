package middleware_test //black box testing

import (
	"net/http"
	"testing"
	"io"
	"time"
	"context"
	"bytes"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/site"
	"github.com/oursky/pageship/internal/config"
)

type mockHandler struct {
	executeCount int
}

func (mh *mockHandler) serve(w http.ResponseWriter, r *http.Request) {
	mh.executeCount++
}

type mockFS struct{}

func (m mockFS) Stat(path string) (*site.FileInfo, error) {
	return &site.FileInfo {
		IsDir: false,
		ModTime: time.Now(),
		Size: 0,
		ContentType: "",
		Hash: "",
	}, nil
}

type mockRSCloser struct{
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
	mh := mockHandler{}
	load := func(r io.ReadSeeker) (*bytes.Buffer, int64, error) {
		b, err := io.ReadAll(r)
		nb := bytes.NewBuffer(b)
		if err != nil {
			return nb, 0, err
		}
		return nb, int64(nb.Len()), nil 
	}
	contentCache, err := cache.NewContentCache[middleware.ContentCacheKey](1<<24, true, load)
	assert.Empty(t, err)
	cacheContext := middleware.NewCacheContext(contentCache)

	sc := config.DefaultSiteConfig()
	mockSiteDescriptor := site.Descriptor{
		ID: "",
		Domain: "",
		Config: &sc,
		FS: mockFS{},
	}
	h := cacheContext.Cache(&mockSiteDescriptor, http.HandlerFunc(mh.serve))

	//Act Assert
	req, err := http.NewRequest("GET", "endpoint", bytes.NewBuffer([]byte("body")))
	assert.Empty(t, err)
	rec := httptest.NewRecorder()
	assert.Equal(t, 0, mh.executeCount)
	h.ServeHTTP(rec, req)
	assert.Equal(t, 1, mh.executeCount)
	h.ServeHTTP(rec, req)
	assert.Equal(t, 1, mh.executeCount)
	h.ServeHTTP(rec, req)
	assert.Equal(t, 1, mh.executeCount)
}

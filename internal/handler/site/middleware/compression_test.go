package middleware_test //black box testing

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/site"
	"github.com/stretchr/testify/assert"
)

type mockHandler struct {
	executeCount int
}

func (mh *mockHandler) serve(w http.ResponseWriter, r *http.Request) {
	mh.executeCount++
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
	mh := mockHandler{}
	sc := config.DefaultSiteConfig()
	mockSiteDescriptor := site.Descriptor{
		ID:     "",
		Domain: "",
		Config: &sc,
		FS:     mockFS{},
	}
	h := middleware.Compression(&mockSiteDescriptor, http.HandlerFunc(mh.serve))

	//Act Assert
	req, err := http.NewRequest("GET", "endpoint", bytes.NewBuffer([]byte("body")))
	assert.Empty(t, err)
	req.Header.Add("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	rec.Header().Add("Content-Type", "text/plain")
	h.ServeHTTP(rec, req)
	assert.Equal(t, "gzip", rec.Result().Header.Get("Content-Encoding"))

	req, err = http.NewRequest("GET", "endpoint", bytes.NewBuffer([]byte("body")))
	assert.Empty(t, err)
	req.Header.Add("Accept-Encoding", "br")
	rec = httptest.NewRecorder()
	rec.Header().Add("Content-Type", "text/plain")
	h.ServeHTTP(rec, req)
	assert.Equal(t, "br", rec.Result().Header.Get("Content-Encoding"))
}

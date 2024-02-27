package middleware_test

import (
	"bytes"
	"context"
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"
	"time"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/site"
)

type mockHandler struct {
	publicFS site.FS
}

func (mh mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rsc, _ := mh.publicFS.Open(r.Context(), r.URL.Path)
	http.ServeContent(w, r, path.Base(r.URL.Path), time.Now(), rsc)
}

type RSCAdapter struct {
	*bytes.Reader
}

func (rsca RSCAdapter) Close() error {
	return nil
}

type FSAdapter struct {
	embed.FS
}

func (fa FSAdapter) Open(c context.Context, s string) (io.ReadSeekCloser, error) {
	b := []byte{}
	f, _ := fa.FS.Open(s)
	f.Read(b)
	return RSCAdapter{bytes.NewReader(b)}, nil
}

func (fa FSAdapter) Stat(string) (*site.FileInfo, error) {
	return &site.FileInfo{
		IsDir:       false,
		ModTime:     time.Now(),
		Size:        0,
		ContentType: "",
		Hash:        "",
	}, nil
}

//go:embed testdata/testrootwith404
var mockFS embed.FS

func TestRootWith404(t *testing.T) {
	mh := mockHandler{}
	sc := config.DefaultSiteConfig()
	mockSiteDescriptor := site.Descriptor{
		ID:     "",
		Domain: "",
		Config: &sc,
		FS:     FSAdapter{mockFS},
	}
	h := middleware.NotFound(&mockSiteDescriptor, mh)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, &http.Request{
		URL: &url.URL{
			Path: "/",
		},
	})

}

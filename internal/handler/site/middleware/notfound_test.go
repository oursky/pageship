package middleware_test

import (
	"bytes"
	"context"
	"embed"
	"fmt"
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
	"github.com/stretchr/testify/assert"
)

type mockHandler struct {
	publicFS site.FS
	bruh     string
}

func (mh mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mh.bruh = r.URL.Path
	fmt.Println("AAAAAAAAAAAAAAA")
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
	b := make([]byte, 1000)
	f, _ := fa.FS.Open(s)
	f.Read(b)
	return RSCAdapter{bytes.NewReader(b)}, nil
}

func (fa FSAdapter) Stat(s string) (*site.FileInfo, error) {
	f, err := fa.FS.Open(s)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	return &site.FileInfo{
		IsDir:       st.IsDir(),
		ModTime:     st.ModTime(),
		Size:        st.Size(),
		ContentType: "",
		Hash:        "",
	}, err
}

//go:embed testdata/testrootwith404
var testrootwith404FS embed.FS

func TestRootWith404(t *testing.T) {
	mh := mockHandler{FSAdapter{testrootwith404FS}, ""}
	sc := config.DefaultSiteConfig()
	mockSiteDescriptor := site.Descriptor{
		ID:     "",
		Domain: "",
		Config: &sc,
		FS:     FSAdapter{testrootwith404FS},
	}
	h := middleware.NotFound(&mockSiteDescriptor, mh)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, &http.Request{
		URL: &url.URL{
			Path: "/index.html",
		},
	})
	res := rec.Result()
	assert.Equal(t, 404, res.StatusCode)
	bo, _ := io.ReadAll(res.Body)
	assert.Equal(t, "rootwith404_404", string(bo))
	assert.Equal(t, "help", mh.bruh)

	h.ServeHTTP(rec, &http.Request{
		URL: &url.URL{
			Path: "/404.html",
		},
	})
	res = rec.Result()
	assert.Equal(t, 200, res.StatusCode)
	bo, _ = io.ReadAll(res.Body)
	assert.Equal(t, "rootwith404_404", string(bo))
	assert.Equal(t, "help", mh.bruh)

	h.ServeHTTP(rec, &http.Request{
		URL: &url.URL{
			Path: "/nonexistant.html",
		},
	})
	res = rec.Result()
	assert.Equal(t, 404, res.StatusCode)
	bo, _ = io.ReadAll(res.Body)
	assert.Equal(t, "rootwith404_404", string(bo))
	assert.Equal(t, "help", mh.bruh)

	h.ServeHTTP(rec, &http.Request{
		URL: &url.URL{
			Path: "/nonexistant/index.html",
		},
	})
	res = rec.Result()
	assert.Equal(t, 404, res.StatusCode)
	bo, _ = io.ReadAll(res.Body)
	assert.Equal(t, "rootwith404_404", string(bo))
}

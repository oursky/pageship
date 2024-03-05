package middleware_test

import (
	"bytes"
	"context"
	"embed"
	"io"
	"io/fs"
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

type notFoundMockHandler struct {
	publicFS site.FS
}

func (mh notFoundMockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	subdir string
}

func (fa FSAdapter) Open(c context.Context, s string) (io.ReadSeekCloser, error) {
	f, _ := fa.FS.Open(path.Join(fa.subdir, s))
	b, _ := io.ReadAll(f)
	return RSCAdapter{bytes.NewReader(b)}, nil
}

func (fa FSAdapter) Stat(s string) (*site.FileInfo, error) {
	st, err := fs.Stat(fa.FS, path.Join(fa.subdir, s))
	if err != nil {
		return nil, err
	}
	return &site.FileInfo{
		IsDir:       st.IsDir(),
		ModTime:     st.ModTime(),
		Size:        st.Size(),
		ContentType: "",
		Hash:        "",
	}, nil
}

var default404 = "404 page not found\n"

func AssertResponse(t *testing.T, h http.Handler, p string, sc int, cont string) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, &http.Request{
		URL: &url.URL{
			Path: p,
		},
	})
	res := rec.Result()
	assert.Equal(t, sc, res.StatusCode)
	bo, _ := io.ReadAll(res.Body)
	assert.Equal(t, cont, string(bo))
}

//go:embed testdata/testrootwith404
var testrootwith404FS embed.FS

func TestRootWith404(t *testing.T) {
	mh := notFoundMockHandler{FSAdapter{testrootwith404FS, "testdata/testrootwith404"}}
	sc := config.DefaultSiteConfig()
	mockSiteDescriptor := site.Descriptor{
		ID:     "",
		Domain: "",
		Config: &sc,
		FS:     FSAdapter{testrootwith404FS, "testdata/testrootwith404"},
	}
	h := middleware.NotFound(&mockSiteDescriptor, mh)

	AssertResponse(t, h, "/index.html", 404, "testrootwith404_404")
	AssertResponse(t, h, "/404.html", 200, "testrootwith404_404")
	AssertResponse(t, h, "/nonexistant.html", 404, "testrootwith404_404")
	AssertResponse(t, h, "/nonexistant/index.html", 404, "testrootwith404_404")
}

//go:embed testdata/testrootno404
var testrootno404FS embed.FS

func TestRootno404(t *testing.T) {
	mh := notFoundMockHandler{FSAdapter{testrootno404FS, "testdata/testrootno404"}}
	sc := config.DefaultSiteConfig()
	mockSiteDescriptor := site.Descriptor{
		ID:     "",
		Domain: "",
		Config: &sc,
		FS:     FSAdapter{testrootno404FS, "testdata/testrootno404"},
	}
	h := middleware.NotFound(&mockSiteDescriptor, mh)

	AssertResponse(t, h, "/index.html", 200, "testrootno404_index")
	AssertResponse(t, h, "/404.html", 404, default404)
	AssertResponse(t, h, "/nonexistant.html", 404, default404)
	AssertResponse(t, h, "/nonexistant/index.html", 404, default404)
}

//go:embed testdata/testsubdirno404
var testsubdirno404FS embed.FS

func TestSubdirNo404(t *testing.T) {
	mh := notFoundMockHandler{FSAdapter{testsubdirno404FS, "testdata/testsubdirno404"}}
	sc := config.DefaultSiteConfig()
	mockSiteDescriptor := site.Descriptor{
		ID:     "",
		Domain: "",
		Config: &sc,
		FS:     FSAdapter{testsubdirno404FS, "testdata/testsubdirno404"},
	}
	h := middleware.NotFound(&mockSiteDescriptor, mh)

	AssertResponse(t, h, "/subdir/index.html", 200, "testsubdirno404_index")
	AssertResponse(t, h, "/subdir/404.html", 404, default404)
	AssertResponse(t, h, "/subdir/nonexistant.html", 404, default404)
	AssertResponse(t, h, "/subdir/nonexistant/index.html", 404, default404)
}

//go:embed testdata/testsubdirwith404
var testsubdirwith404FS embed.FS

func TestSubdirWith404(t *testing.T) {
	mh := notFoundMockHandler{FSAdapter{testsubdirwith404FS, "testdata/testsubdirwith404"}}
	sc := config.DefaultSiteConfig()
	mockSiteDescriptor := site.Descriptor{
		ID:     "",
		Domain: "",
		Config: &sc,
		FS:     FSAdapter{testsubdirwith404FS, "testdata/testsubdirwith404"},
	}
	h := middleware.NotFound(&mockSiteDescriptor, mh)

	AssertResponse(t, h, "/subdir/index.html", 404, "testsubdirwith404_404")
	AssertResponse(t, h, "/404.html", 200, "testsubdirwith404_404")
	AssertResponse(t, h, "/subdir/nonexistant.html", 404, "testsubdirwith404_404")
	AssertResponse(t, h, "/subdir/nonexistant/index.html", 404, "testsubdirwith404_404")
}

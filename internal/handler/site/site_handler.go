package site

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/site"
)

type SiteHandler struct {
	desc     *site.Descriptor
	publicFS site.FS
	next     http.Handler
	cc       *cache.ContentCache
}

func NewSiteHandler(desc *site.Descriptor, middlewares []Middleware) *SiteHandler {
	cc, err := cache.NewContentCache(1 << 24) //16 MiB
	if err != nil {
		cc = nil
	}

	h := &SiteHandler{
		desc:     desc,
		publicFS: site.SubFS(desc.FS, path.Clean("/"+desc.Config.Public)),
		cc:       cc,
	}

	publicDesc := *desc
	publicDesc.FS = site.SubFS(desc.FS, path.Clean("/"+desc.Config.Public))
	h.next = applyMiddleware(&publicDesc, middlewares, http.HandlerFunc(h.serveFile))
	return h
}

func (h *SiteHandler) ID() string {
	return h.desc.ID
}

// ref http.fileHandler.ServeHTTP
func (h *SiteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const options = http.MethodOptions + ", " + http.MethodGet + ", " + http.MethodHead

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		r = r.WithContext(withSiteContext(r.Context()))

		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = "/" + r.URL.Path
		}
		h.next.ServeHTTP(w, r)

	case http.MethodOptions:
		w.Header().Set("Allow", options)

	default:
		w.Header().Set("Allow", options)
		http.Error(w, "read-only", http.StatusMethodNotAllowed)
	}
}

func (h *SiteHandler) serveFile(w http.ResponseWriter, r *http.Request) {
	info, err := h.publicFS.Stat(r.URL.Path)
	if os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if info.ContentType != "" {
		w.Header().Set("Content-Type", info.ContentType)
	}
	if info.Hash != "" {
		w.Header().Set("ETag", fmt.Sprintf(`"%s"`, info.Hash))
		w.Header().Set("Cache-Control", "public, max-age=31536000, no-cache")
	}

	lReader := &lazyReader{
		fs:   h.publicFS,
		path: r.URL.Path,
		ctx:  r.Context(),
	}
	defer lReader.Close()
	var reader = io.ReadSeeker(lReader)

	if info.Hash != "" {
		cell, err := h.cc.GetContent(info.Hash, reader)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		} else {
			reader = bytes.NewReader(cell.Data.Bytes())
		}
	}
	writer := httputil.NewTimeoutResponseWriter(w, 10*time.Second)
	http.ServeContent(writer, r, path.Base(r.URL.Path), info.ModTime, reader)
}

type lazyReader struct {
	fs     site.FS
	path   string
	ctx    context.Context
	reader io.ReadSeekCloser
}

func (r *lazyReader) init() error {
	if r.reader != nil {
		return nil
	}

	reader, err := r.fs.Open(r.ctx, r.path)
	if err != nil {
		return err
	}

	r.reader = reader
	return nil
}

func (r *lazyReader) Close() error {
	if r.reader != nil {
		return r.reader.Close()
	}
	return nil
}

func (r *lazyReader) Read(p []byte) (n int, err error) {
	if err := r.init(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func (r *lazyReader) Seek(offset int64, whence int) (int64, error) {
	if err := r.init(); err != nil {
		return 0, err
	}
	return r.reader.Seek(offset, whence)
}

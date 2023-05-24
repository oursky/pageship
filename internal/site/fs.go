package site

import (
	"context"
	"io"
	"path"
	"time"

	"github.com/oursky/pageship/internal/config"
)

type FileInfo struct {
	IsDir       bool
	ModTime     time.Time
	Size        int64
	ContentType string
	Hash        string
}

type FS interface {
	Stat(path string) (*FileInfo, error)
	Open(ctx context.Context, path string) (io.ReadSeekCloser, error)
}

type Descriptor struct {
	ID     string
	Config *config.SiteConfig
	FS     FS
}

type subFS struct {
	fs  FS
	dir string
}

func SubFS(fs FS, dir string) FS { return &subFS{fs: fs, dir: dir} }

func (s *subFS) Stat(p string) (*FileInfo, error) {
	return s.fs.Stat(path.Join(s.dir, p))
}

func (s *subFS) Open(ctx context.Context, p string) (io.ReadSeekCloser, error) {
	return s.fs.Open(ctx, path.Join(s.dir, p))
}

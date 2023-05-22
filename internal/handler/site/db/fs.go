package db

import (
	"context"
	"io"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/storage"
)

type storageFS struct {
	ctx     context.Context
	storage *storage.Storage
	modTime time.Time
	prefix  string
	fileMap map[string]models.FileEntry
	files   []models.FileEntry
}

func newStorageFS(ctx context.Context, storage *storage.Storage, deployment *models.Deployment) *storageFS {
	files := deployment.Metadata.Files
	fileMap := make(map[string]models.FileEntry)
	for _, entry := range deployment.Metadata.Files {
		fileMap[entry.Path] = entry
	}

	return &storageFS{
		ctx:     ctx,
		storage: storage,
		modTime: *deployment.UploadedAt,
		prefix:  deployment.StorageKeyPrefix,
		fileMap: fileMap,
		files:   files,
	}
}

func (f *storageFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	if name[0] == '.' { // "./x" -> "/x", "." -> ""
		name = name[1:]
	}
	if name[0] != '/' { // "/x" -> "/", "" -> "/"
		name = "/" + name
	}

	if entry, ok := f.fileMap[name]; ok {
		return &file{fs: f, entry: entry}, nil
	}
	if entry, ok := f.fileMap[name+"/"]; ok {
		return &file{fs: f, entry: entry}, nil
	}

	return nil, &fs.PathError{
		Op:   "open",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

type file struct {
	fs            *storageFS
	entry         models.FileEntry
	readDirOffset int
	reader        io.ReadSeekCloser
}

func (f *file) Info() (fs.FileInfo, error) {
	return f, nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *file) IsDir() bool {
	return f.entry.Path[len(f.entry.Path)-1] == '/'
}

func (f *file) ModTime() time.Time {
	return f.fs.modTime
}

func (f *file) Mode() fs.FileMode {
	var mode fs.FileMode = 0777
	if f.IsDir() {
		mode |= fs.ModeDir
	}
	return mode
}

func (f *file) Type() fs.FileMode {
	return f.Mode() & fs.ModeType
}

func (f *file) Name() string {
	return path.Base(f.entry.Path)
}

func (f *file) Size() int64 {
	return f.entry.Size
}

func (*file) Sys() any {
	return nil
}

func (f *file) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.IsDir() {
		return nil, &fs.PathError{
			Op:   "open",
			Path: f.entry.Path,
			Err:  fs.ErrInvalid,
		}
	}

	var entries []fs.DirEntry
	for _, e := range f.fs.files[f.readDirOffset:] {
		f.readDirOffset++

		if !strings.HasPrefix(e.Path, f.entry.Path) {
			continue
		}

		p := strings.TrimPrefix(e.Path, f.entry.Path)
		if p == "" { // this dir
			continue
		}
		if p[len(p)-1] == '/' { // trim trailing slash
			p = p[:len(p)-1]
		}

		if strings.Contains(p, "/") { // in child dirs
			continue
		}

		entries = append(entries, &file{fs: f.fs, entry: e})
		if n > 0 && len(entries) == n {
			break
		}
	}

	if n > 0 && f.readDirOffset == len(f.fs.files) && len(entries) == 0 {
		return nil, io.EOF
	}
	return entries, nil
}

func (f *file) Read(p []byte) (int, error) {
	if f.reader == nil {
		key := f.fs.prefix + f.entry.Path
		reader, err := f.fs.storage.OpenRead(f.fs.ctx, key)
		if err != nil {
			return 0, err
		}

		f.reader = reader
	}
	return f.reader.Read(p)
}

func (f *file) Close() error {
	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

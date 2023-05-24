package db

import (
	"context"
	"io"
	"io/fs"
	"time"

	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/site"
	"github.com/oursky/pageship/internal/storage"
)

type storageFS struct {
	storage *storage.Storage
	modTime time.Time
	prefix  string
	fileMap map[string]models.FileEntry
	files   []models.FileEntry
}

func newStorageFS(storage *storage.Storage, deployment *models.Deployment) site.FS {
	files := deployment.Metadata.Files
	fileMap := make(map[string]models.FileEntry)
	for _, entry := range deployment.Metadata.Files {
		fileMap[entry.Path] = entry
	}

	return &storageFS{
		storage: storage,
		modTime: *deployment.UploadedAt,
		prefix:  deployment.StorageKeyPrefix,
		fileMap: fileMap,
		files:   files,
	}
}

func (f *storageFS) lookup(path string) (models.FileEntry, bool) {
	if entry, ok := f.fileMap[path]; ok {
		return entry, true
	}

	if entry, ok := f.fileMap[path+"/"]; ok {
		return entry, true
	}

	return models.FileEntry{}, false
}

func (f *storageFS) Stat(path string) (*site.FileInfo, error) {
	entry, ok := f.lookup(path)
	if !ok {
		return nil, &fs.PathError{
			Op:   "open",
			Path: path,
			Err:  fs.ErrNotExist,
		}
	}

	return &site.FileInfo{
		IsDir:       entry.Path[len(entry.Path)-1] == '/',
		ModTime:     f.modTime,
		Size:        entry.Size,
		ContentType: entry.ContentType,
		Hash:        entry.Hash,
	}, nil
}

func (f *storageFS) Open(ctx context.Context, path string) (io.ReadSeekCloser, error) {
	entry, ok := f.lookup(path)
	if !ok {
		return nil, &fs.PathError{
			Op:   "open",
			Path: path,
			Err:  fs.ErrNotExist,
		}
	}

	key := f.prefix + entry.Path
	reader, err := f.storage.OpenRead(ctx, key)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

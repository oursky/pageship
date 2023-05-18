package deploy

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/oursky/pageship/internal/models"
	"golang.org/x/crypto/sha3"
)

var ErrTooManyFiles = errors.New("too many files collected")

func CollectFileList(fsys fs.FS) ([]models.FileEntry, error) {
	var entries []models.FileEntry
	var walker fs.WalkDirFunc
	walker = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		if d.Type()&fs.ModeSymlink == fs.ModeSymlink {
			return fs.WalkDir(fsys, path, walker)
		}

		entry, err := makeFileEntry(fsys, path, d)
		if err != nil {
			return err
		}
		entries = append(entries, entry)
		if len(entries) > models.MaxFiles {
			return ErrTooManyFiles
		}
		return nil
	}

	err := fs.WalkDir(fsys, ".", walker)
	fmt.Printf("%+v %s\n", entries, err)
	return entries, err
}

func makeFileEntry(fsys fs.FS, filePath string, d fs.DirEntry) (models.FileEntry, error) {
	info, err := d.Info()
	if err != nil {
		return models.FileEntry{}, err
	}

	path := "/" + filepath.ToSlash(filePath)
	size := info.Size()
	hash := ""
	if info.IsDir() {
		size = 0
		path += "/"
	} else {
		file, err := fsys.Open(filePath)
		if err != nil {
			return models.FileEntry{}, err
		}
		h := sha3.New256()
		_, err = io.Copy(h, file)
		if err != nil {
			return models.FileEntry{}, err
		}
		hash = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	}

	return models.FileEntry{
		FilePath: filePath,
		Path:     path,
		Size:     size,
		Hash:     hash,
	}, nil

}

package deploy

import (
	"archive/tar"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/oursky/pageship/internal/models"
)

var ErrTooManyFiles error = Error("too many files collected")

func CollectFileList(fsys fs.FS, now time.Time, tarfile *os.File) ([]models.FileEntry, error) {
	comp, err := zstd.NewWriter(tarfile)
	if err != nil {
		return nil, err
	}
	defer comp.Close()

	writer := tar.NewWriter(comp)
	defer writer.Close()

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

		entry, err := addFile(fsys, path, d, now, writer)
		if err != nil {
			return err
		}
		entries = append(entries, entry)
		if len(entries) > models.MaxFiles {
			return ErrTooManyFiles
		}
		return nil
	}

	err = fs.WalkDir(fsys, ".", walker)
	return entries, err
}

func addFile(fsys fs.FS, filePath string, d fs.DirEntry, now time.Time, writer *tar.Writer) (models.FileEntry, error) {
	info, err := d.Info()
	if err != nil {
		return models.FileEntry{}, err
	}

	header := tar.Header{
		Name:    "/" + filepath.ToSlash(filePath),
		ModTime: now,
		Size:    info.Size(),
	}
	if info.IsDir() {
		header.Typeflag = tar.TypeDir
		header.Size = 0
		header.Name += "/"
	} else {
		header.Typeflag = tar.TypeReg
	}
	writer.WriteHeader(&header)

	hash := ""
	if !info.IsDir() {
		file, err := fsys.Open(filePath)
		if err != nil {
			return models.FileEntry{}, err
		}

		h := NewFileHash()
		_, err = io.Copy(writer, io.TeeReader(file, h))
		if err != nil {
			return models.FileEntry{}, err
		}
		hash = h.Sum()
	}

	return models.FileEntry{
		FilePath: filePath,
		Path:     header.Name,
		Size:     header.Size,
		Hash:     hash,
	}, nil
}

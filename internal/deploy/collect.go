package deploy

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/oursky/pageship/internal/models"
)

var ErrTooManyFiles error = Error("too many files collected")

type Collector struct {
	files   []models.FileEntry
	modTime time.Time

	closed bool
	comp   *zstd.Encoder
	writer *tar.Writer
}

func NewCollector(modTime time.Time, tarfile *os.File) (coll *Collector, err error) {
	coll = &Collector{
		files:   nil,
		modTime: modTime,
		closed:  false,
	}
	defer func() {
		if err != nil {
			coll.Close()
			coll = nil
		}
	}()

	coll.comp, err = zstd.NewWriter(tarfile, zstd.WithWindowSize(zstdWindowSize))
	if err != nil {
		return
	}

	coll.writer = tar.NewWriter(coll.comp)

	return
}

func (c *Collector) Close() {
	if c.closed {
		return
	}

	if c.writer != nil {
		c.writer.Close()
	}
	if c.comp != nil {
		c.comp.Close()
	}
	c.closed = true
}

func (c *Collector) Files() []models.FileEntry {
	return c.files
}

func (c *Collector) AddDir(path string) {
	header := &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     path,
		ModTime:  c.modTime,
		Size:     0,
	}
	c.writer.WriteHeader(header)

	c.files = append(c.files, models.FileEntry{
		Path:        header.Name,
		Size:        header.Size,
		Hash:        "",
		ContentType: "",
	})
}

func (c *Collector) AddFile(path string, data []byte) error {
	h := NewFileHash()
	_, err := io.Copy(h, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	hash := h.Sum()

	header := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     path,
		ModTime:  c.modTime,
		Size:     int64(len(data)),
	}
	c.writer.WriteHeader(header)
	c.writer.Write(data)

	c.files = append(c.files, models.FileEntry{
		Path:        header.Name,
		Size:        header.Size,
		Hash:        hash,
		ContentType: models.DetectContentType(header.Name, data),
	})
	return nil
}

func (c *Collector) Collect(fsys fs.FS, dir string) error {
	var walker fs.WalkDirFunc
	walker = func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Type()&fs.ModeSymlink == fs.ModeSymlink {
			return fs.WalkDir(fsys, p, walker)
		}

		entry, err := addFile(fsys, p, d, dir, c.modTime, c.writer)
		if err != nil {
			return err
		}

		c.files = append(c.files, entry)
		if len(c.files) > models.MaxFiles {
			return ErrTooManyFiles
		}
		return nil
	}

	return fs.WalkDir(fsys, ".", walker)
}

func addFile(fsys fs.FS, filePath string, d fs.DirEntry, dir string, modTime time.Time, writer *tar.Writer) (models.FileEntry, error) {
	info, err := d.Info()
	if err != nil {
		return models.FileEntry{}, err
	}

	path := path.Join(dir, filepath.ToSlash(filePath))
	header := tar.Header{
		Name:    path,
		ModTime: modTime,
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
	contentType := ""
	if !info.IsDir() {
		file, err := fsys.Open(filePath)
		if err != nil {
			return models.FileEntry{}, err
		}

		initialBytes := make([]byte, 512)
		n, _ := io.ReadFull(file, initialBytes)
		initialBytes = initialBytes[:n]
		contentType = models.DetectContentType(header.Name, initialBytes)

		fileData := io.MultiReader(bytes.NewBuffer(initialBytes), file)
		h := NewFileHash()
		_, err = io.Copy(writer, io.TeeReader(fileData, h))
		if err != nil {
			return models.FileEntry{}, err
		}
		hash = h.Sum()
	}

	return models.FileEntry{
		Path:        header.Name,
		Size:        header.Size,
		Hash:        hash,
		ContentType: contentType,
	}, nil
}

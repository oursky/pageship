package deploy

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/oursky/pageship/internal/models"
)

var ErrUnexpectedFile error = Error("unexpected file")
var ErrUnexpectedFileSize error = Error("unexpected file size")
var ErrMissingFile error = Error("missing file")

const zstdWindowSize = 1024 * 1024 * 1 // 1MB
const zstdMaxMemory = 1024 * 1024 * 1  // 1MB

func ExtractFiles(r io.Reader, files []models.FileEntry, handle func(models.FileEntry, io.Reader) error) error {
	pending := make(map[string]models.FileEntry)
	for _, entry := range files {
		pending[entry.Path] = entry
	}

	decomp, err := zstd.NewReader(r, zstd.WithDecoderMaxMemory(zstdMaxMemory))
	if err != nil {
		return err
	}
	defer decomp.Close()

	tr := tar.NewReader(decomp)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break // End of archive
		} else if err != nil {
			return err
		}

		file, ok := pending[hdr.Name]
		if !ok {
			return fmt.Errorf("%w: %s", ErrUnexpectedFile, hdr.Name)
		}
		if hdr.Size != file.Size {
			return fmt.Errorf("%w: %s", ErrUnexpectedFileSize, hdr.Name)
		}

		if err := handle(file, tr); err != nil {
			return fmt.Errorf("%s: %w", hdr.Name, err)
		}

		delete(pending, hdr.Name)
	}

	if len(pending) > 0 {
		missing := ""
		for path := range pending {
			missing = path
		}
		return fmt.Errorf("%w: %s", ErrMissingFile, missing)
	}

	return nil
}

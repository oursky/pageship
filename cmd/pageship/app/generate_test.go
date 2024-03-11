package app

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata
var testfs embed.FS

type FSAdapter struct {
	embed.FS
	subdir string
}

func (fa FSAdapter) Open(s string) (fs.File, error) {
	return fa.FS.Open(path.Join(fa.subdir, s))
}

func (fa FSAdapter) ReadFile(s string) ([]byte, error) {
	return fa.FS.ReadFile(path.Join(fa.subdir, s))
}

func TestGenerate(t *testing.T) {
	Version = "1.2.3"
	fsa := FSAdapter{testfs, "testdata"}

	fs.WalkDir(fsa, ".", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path)
		return nil
	})

	s, err := generateContent(fsa)
	assert.Empty(t, err)
	assert.Equal(t, `FROM ghcr.io/oursky/pageship:v1.2.3
EXPOSE 8000
COPY dist /var/pageship

# INSTRUCTIONS:
# 1. install docker (if it is not installed yet)
# 2. open a terminal and navigate to folder containing pageship.toml
# 3. run in terminal:
#      pageship generate dockerfile
# 4. build the image:
#      docker build -t IMAGETAG .
# 5. run the container:
#      docker run -d --name CONTAINERNAME -p PORT:8000 IMAGETAG
# 6. visit in browser (URL):
#      localhost:PORT`, s)
}

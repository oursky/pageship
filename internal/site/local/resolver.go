package local

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/site"
)

func NewMultiSiteResolver(fs fs.FS, defaultSite string, sites map[string]config.SitesConfigEntry) site.Resolver {
	if len(sites) == 0 {
		return &resolverAdhoc{fs: fs, defaultSite: defaultSite}
	}

	return &resolverStatic{
		fs:          fs,
		defaultSite: defaultSite,
		sites:       sites,
	}
}

func checkSiteFS(fsys fs.FS) error {
	f, err := fsys.Open(".")
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("not directory: %s", fi.Name())
	}
	return nil
}

func loadConfig(fs fs.FS) (*config.Config, error) {
	loader := config.NewLoader(config.SiteConfigName)

	conf := config.DefaultConfig()
	if err := loader.Load(fs, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

type siteFS struct{ fs fs.FS }

func (f siteFS) Stat(path string) (*site.FileInfo, error) {
	i, err := fs.Stat(f.fs, fsPath(path))
	if err != nil {
		return nil, err
	}

	return &site.FileInfo{
		IsDir:       i.IsDir(),
		ModTime:     i.ModTime(),
		Size:        i.Size(),
		ContentType: "",
		Hash:        "",
	}, nil
}

func (f siteFS) Open(ctx context.Context, path string) (io.ReadSeekCloser, error) {
	file, err := f.fs.Open(fsPath(path))
	if err != nil {
		return nil, err
	}
	return file.(io.ReadSeekCloser), nil
}

func fsPath(url string) string {
	if url == "/" {
		return "."
	} else {
		return strings.TrimPrefix(strings.TrimSuffix(url, "/"), "/")
	}
}

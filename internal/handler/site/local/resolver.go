package local

import (
	"fmt"
	"io/fs"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/handler/site"
)

func NewMultiSiteResolver(fs fs.FS, sites map[string]config.SitesConfigEntry) site.Resolver {
	if len(sites) == 0 {
		return &resolverAdhoc{fs: fs}
	}

	return &resolverStatic{
		fs:    fs,
		sites: sites,
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

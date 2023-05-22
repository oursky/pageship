package local

import (
	"fmt"
	"io/fs"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/handler/site"
)

func NewMultiSiteResolver(fs fs.FS, conf *config.SitesConfig) site.Resolver {
	if len(conf.Sites) == 0 {
		return &resolverAdhoc{fs: fs, defaultSite: conf.DefaultSite}
	}

	return &resolverStatic{
		fs:          fs,
		defaultSite: conf.DefaultSite,
		sites:       conf.Sites,
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

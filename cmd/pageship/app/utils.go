package app

import (
	"os"

	"github.com/oursky/pageship/internal/config"
)

func loadConfig(dir string) (*config.Config, error) {
	loader := config.NewLoader(config.SiteConfigName)

	conf := config.DefaultConfig()
	if err := loader.Load(os.DirFS(dir), &conf); err != nil {
		return nil, err
	}

	conf.SetDefaults()
	return &conf, nil
}

func tryLoadAppID() string {
	conf, err := loadConfig(".")
	if err != nil {
		return ""
	}
	return conf.App.ID
}

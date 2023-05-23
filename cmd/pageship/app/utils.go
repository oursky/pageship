package app

import (
	"os"

	"github.com/oursky/pageship/internal/config"
)

func tryLoadAppID() string {
	loader := config.NewLoader(config.SiteConfigName)

	conf := config.DefaultConfig()
	if err := loader.Load(os.DirFS("."), &conf); err != nil {
		Debug("Failed to load config: %s", err)
		return ""
	}
	return conf.ID
}

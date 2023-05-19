package config

const SiteConfigName = "pageship"

const DefaultSite = "app"

type SiteConfig struct {
	Public string `json:"public" validate:"required"`
}

func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		Public: ".",
	}
}

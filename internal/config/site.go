package config

const SiteConfigName = "pageship"

const DefaultSite = "main"

type SiteConfig struct {
	Public string `json:"public" validate:"required"`
}

func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		Public: ".",
	}
}

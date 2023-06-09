package config

const SiteConfigName = "pageship"

const DefaultSite = "main"

type SiteConfig struct {
	Public string              `json:"public" pageship:"required"`
	Access []CredentialMatcher `json:"access" pageship:"omitempty"`
}

func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		Public: ".",
	}
}

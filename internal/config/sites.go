package config

const SitesConfigName = "sites"

type SitesConfig struct {
	HostPattern string
	DefaultSite string
	Sites       map[string]SitesConfigEntry
}

type SitesConfigEntry struct {
	Context string
}

func DefaultSitesConfig() *SitesConfig {
	return &SitesConfig{
		HostPattern: `(?:(.+)\.)?localhost`,
		DefaultSite: DefaultSite,
		Sites:       nil,
	}
}

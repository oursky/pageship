package config

const SitesConfigName = "sites"

type SitesConfig struct {
	HostPattern string                      `json:"hostPattern"`
	DefaultSite string                      `json:"defaultSite"`
	Sites       map[string]SitesConfigEntry `json:"sites,omitempty"`
}

type SitesConfigEntry struct {
	Context string `json:"context"`
}

func DefaultSitesConfig() *SitesConfig {
	return &SitesConfig{
		HostPattern: DefaultHostPattern,
		DefaultSite: DefaultSite,
		Sites:       make(map[string]SitesConfigEntry),
	}
}

package config

const SitesConfigName = "sites"

type SitesConfig struct {
	Sites map[string]SitesConfigEntry `json:"sites,omitempty"`
}

type SitesConfigEntry struct {
	Context string `json:"context"`
}

func DefaultSitesConfig() *SitesConfig {
	return &SitesConfig{
		Sites: make(map[string]SitesConfigEntry),
	}
}

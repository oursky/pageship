package config

const SitesConfigName = "sites"

type SitesConfig struct {
	Sites map[string]SitesConfigEntry `json:"sites,omitempty"`
}

type SitesConfigEntry struct {
	Context string `json:"context"`
	Domain  string `json:"domain"`
}

func DefaultSitesConfig() *SitesConfig {
	return &SitesConfig{
		Sites: make(map[string]SitesConfigEntry),
	}
}

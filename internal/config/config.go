package config

type Config struct {
	ID string `json:"id"`

	AppConfig
	Site SiteConfig `json:"site"`
}

func DefaultConfig() Config {
	return Config{
		AppConfig: DefaultAppConfig(),
		Site:      DefaultSiteConfig(),
	}
}

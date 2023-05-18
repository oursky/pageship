package config

type Config struct {
	AppConfig
	Site SiteConfig
}

func DefaultConfig() Config {
	return Config{
		AppConfig: DefaultAppConfig(),
		Site:      DefaultSiteConfig(),
	}
}

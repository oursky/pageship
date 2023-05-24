package config

type Config struct {
	App  AppConfig  `json:"app"`
	Site SiteConfig `json:"site"`
}

func DefaultConfig() Config {
	return Config{
		App:  DefaultAppConfig(),
		Site: DefaultSiteConfig(),
	}
}

func (c *Config) SetDefaults() {
	c.App.SetDefaults()
}

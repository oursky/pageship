package config

type Config struct {
	ID string `json:"id" validate:"required,dnsLabel"`

	AppConfig `mapstructure:",squash"`
	Site      SiteConfig `json:"site"`
}

func DefaultConfig() Config {
	return Config{
		AppConfig: DefaultAppConfig(),
		Site:      DefaultSiteConfig(),
	}
}

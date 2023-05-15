package config

type ServerConfig struct {
	Site SiteConfig
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Site: SiteConfig{
			Public: ".",
		},
	}
}

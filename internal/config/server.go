package config

type ServerConfig struct {
	Root string `mapstructure:"-"`
	Site SiteConfig
}

func DefaultServerConfig(root string) *ServerConfig {
	return &ServerConfig{
		Root: root,
		Site: SiteConfig{
			Public: root,
		},
	}
}

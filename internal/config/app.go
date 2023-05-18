package config

type AppConfig struct {
	ID                 string
	DefaultEnvironment string
	Environments       map[string]EnvironmentConfig
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		ID:                 "",
		DefaultEnvironment: DefaultEnvironment,
	}
}

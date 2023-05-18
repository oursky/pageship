package config

const DefaultEnvironment = "main"

type EnvironmentConfig struct {
	SitePattern string `json:"sitePattern,omitempty" validate:"max=64"`
}

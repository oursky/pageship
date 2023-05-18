package config

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type AppConfig struct {
	DefaultEnvironment string              `json:"defaultEnvironment" validate:"required,dnsLabel"`
	Environments       []EnvironmentConfig `json:"environments" validate:"max=10"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		DefaultEnvironment: DefaultEnvironment,
		Environments:       make([]EnvironmentConfig, 0),
	}
}

func (c *AppConfig) SetDefaults() {
	if c.Environments == nil {
		c.Environments = make([]EnvironmentConfig, 0)
	}

	foundDefault := false
	for i, env := range c.Environments {
		env.SetDefaults()
		c.Environments[i] = env
	}

	if !foundDefault {
		defaultEnv := EnvironmentConfig{Name: c.DefaultEnvironment}
		defaultEnv.SetDefaults()
		c.Environments = append(c.Environments, defaultEnv)
	}
}

func (c *AppConfig) ResolveSite(site string) (resolved EnvironmentConfig, ok bool) {
	for _, env := range c.Environments {
		pattern, err := env.CompileSitePattern()
		if err != nil {
			break
		}
		if pattern.MatchString(site) {
			resolved = env
			ok = true
			return
		}
	}
	return
}

func (c *AppConfig) Scan(val any) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
func (c *AppConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

package config

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type AppConfig struct {
	DefaultSite  string              `json:"defaultSite" pageship:"required,dnsLabel"`
	Environments []EnvironmentConfig `json:"environments" pageship:"max=10"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		DefaultSite:  DefaultSite,
		Environments: make([]EnvironmentConfig, 0),
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

		pattern, err := env.CompileSitePattern()
		if err == nil && pattern.MatchString(c.DefaultSite) {
			foundDefault = true
		}
	}

	if !foundDefault && len(c.Environments) == 0 {
		defaultEnv := EnvironmentConfig{Name: c.DefaultSite}
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

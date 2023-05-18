package config

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type AppConfig struct {
	DefaultEnvironment string                       `json:"defaultEnvironment" validate:"required,max=64"`
	Environments       map[string]EnvironmentConfig `json:"environments" validate:"max=10,dive,keys,max=64,endkeys"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		DefaultEnvironment: DefaultEnvironment,
		Environments:       make(map[string]EnvironmentConfig),
	}
}

func (c *AppConfig) SetDefaults() {
	if c.Environments == nil {
		c.Environments = make(map[string]EnvironmentConfig)
	}
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

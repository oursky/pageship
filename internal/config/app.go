package config

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type AppConfig struct {
	ID          string          `json:"id" pageship:"required,dnsLabel"`
	DefaultSite string          `json:"defaultSite" pageship:"required,dnsLabel"`
	Sites       []AppSiteConfig `json:"sites" pageship:"max=10"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		DefaultSite: DefaultSite,
		Sites:       make([]AppSiteConfig, 0),
	}
}

func (c *AppConfig) SetDefaults() {
	if c.Sites == nil {
		c.Sites = make([]AppSiteConfig, 0)
	}

	foundDefault := false
	for i, s := range c.Sites {
		c.Sites[i] = s

		pattern, err := s.CompilePattern()
		if err == nil && pattern.MatchString(c.DefaultSite) {
			foundDefault = true
		}
	}

	if !foundDefault && len(c.Sites) == 0 {
		defaultSite := AppSiteConfig{Name: c.DefaultSite}
		c.Sites = append(c.Sites, defaultSite)
	}
}

func (c *AppConfig) ResolveSite(site string) (resolved AppSiteConfig, ok bool) {
	for _, s := range c.Sites {
		pattern, err := s.CompilePattern()
		if err != nil {
			break
		}
		if pattern.MatchString(site) {
			resolved = s
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

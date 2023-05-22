package config

import "regexp"

type EnvironmentConfig struct {
	Name        string `json:"name" validate:"required,dnsLabel"`
	SitePattern string `json:"sitePattern,omitempty" validate:"max=100,regexp"`
}

func (c *EnvironmentConfig) SetDefaults() {
	if c.SitePattern == "" {
		c.SitePattern = regexp.QuoteMeta(c.Name)
	}
}

func (c *EnvironmentConfig) CompileSitePattern() (*regexp.Regexp, error) {
	return regexp.Compile("^" + c.SitePattern + "$")
}

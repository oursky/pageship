package config

import "regexp"

type AppSiteConfig struct {
	Name    string `json:"name" pageship:"excluded_with=Pattern,dnsLabel"`
	Pattern string `json:"pattern,omitempty" pageship:"excluded_with=Name,max=100,regexp"`
}

func (c *AppSiteConfig) CompilePattern() (*regexp.Regexp, error) {
	pattern := c.Pattern
	if c.Name != "" {
		pattern = regexp.QuoteMeta(c.Name)
	}
	return regexp.Compile("^" + pattern + "$")
}

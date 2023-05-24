package config

type AppDeploymentsConfig struct {
	Accessible bool   `json:"accessible"`
	TTL        string `json:"ttl" pageship:"omitempty,duration"`
}

func (c *AppDeploymentsConfig) SetDefaults() {
	if c.TTL == "" {
		c.TTL = "24h"
	}
}

package config

type AppDomainConfig struct {
	Domain string `json:"domain" pageship:"required,max=200,hostname_port,lowercase"`
	Site   string `json:"site" pageship:"required,dnsLabel"`
}

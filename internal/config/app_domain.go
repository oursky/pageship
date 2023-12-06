package config

type AppDomainConfig struct {
	Domain string `json:"domain" pageship:"required,max=200,hostname_port|hostname_rfc1123,lowercase"`
	Site   string `json:"site" pageship:"required,dnsLabel"`
}

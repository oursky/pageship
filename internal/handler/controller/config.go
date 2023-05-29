package controller

import "github.com/oursky/pageship/internal/config"

type Config struct {
	MaxDeploymentSize int64
	StorageKeyPrefix  string
	HostIDScheme      config.HostIDScheme
	HostPattern       *config.HostPattern
	TokenAuthority    string
	TokenSigningKey   []byte
}

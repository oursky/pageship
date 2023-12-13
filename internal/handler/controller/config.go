package controller

import (
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/watch"
)

type Config struct {
	MaxDeploymentSize         int64
	StorageKeyPrefix          string
	HostIDScheme              config.HostIDScheme
	HostPattern               *config.HostPattern
	ReservedApps              map[string]struct{}
	TokenAuthority            string
	TokenSigningKey           []byte
	ACL                       *watch.File[config.ACL]
	DomainVerificationEnabled bool

	ServerVersion       string
	CustomDomainMessage string
}

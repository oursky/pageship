package local

import (
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/domain"
)

func NewResolver(defaultSite string, sites map[string]config.SitesConfigEntry) (domain.Resolver, error) {
	if len(sites) == 0 {
		// Custom domain resolution is not supported for ad-hoc sites.
		return &domain.ResolverNull{}, nil
	}

	return newResolverStatic(defaultSite, sites)
}

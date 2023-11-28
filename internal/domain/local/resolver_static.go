package local

import (
	"context"
	"fmt"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/domain"
)

type resolverStatic struct {
	domains map[string]string
}

func newResolverStatic(defaultSite string, sites map[string]config.SitesConfigEntry) (*resolverStatic, error) {
	domains := make(map[string]string)
	for id, site := range sites {
		if site.Domain == "" {
			continue
		}

		if _, exists := domains[site.Domain]; exists {
			return nil, fmt.Errorf("duplicated domain: %q", site.Domain)
		}

		if defaultSite != "-" && id == defaultSite {
			id = ""
		}
		domains[site.Domain] = id
	}

	return &resolverStatic{domains: domains}, nil
}

func (h *resolverStatic) Kind() string { return "static config" }

func (h *resolverStatic) Resolve(ctx context.Context, hostname string) (string, error) {
	id, ok := h.domains[hostname]
	if !ok {
		return "", domain.ErrDomainNotFound
	}

	return id, nil
}

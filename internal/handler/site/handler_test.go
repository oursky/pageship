package site_test

import (
	"context"
	"errors"
	"testing"

	"github.com/oursky/pageship/internal/domain"
	sitehandler "github.com/oursky/pageship/internal/handler/site"
	"github.com/oursky/pageship/internal/site"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockDomainResolver struct {
	domains map[string]string
}

func (*mockDomainResolver) Kind() string { return "mock" }

func (r *mockDomainResolver) Resolve(ctx context.Context, hostname string) (string, error) {
	if id, ok := r.domains[hostname]; ok {
		return id, nil
	}
	return "", domain.ErrDomainNotFound
}

type mockSiteResolver struct {
	wildcard bool
	sites    map[string]string
}

func (r *mockSiteResolver) IsWildcard() bool { return r.wildcard }

func (*mockSiteResolver) Kind() string { return "mock" }

func (r *mockSiteResolver) Resolve(ctx context.Context, matchedID string) (*site.Descriptor, error) {
	if id, ok := r.sites[matchedID]; ok {
		return &site.Descriptor{ID: id}, nil
	}
	return nil, site.ErrSiteNotFound
}

func TestHandleResolution(t *testing.T) {
	hostPattern := "http://*.pageship.local"
	domainResolver := &mockDomainResolver{
		domains: map[string]string{
			"example.com":     "example",
			"dev.example.com": "dev.example",
		},
	}
	siteResolver := &mockSiteResolver{
		sites: map[string]string{
			"example":     "example/main",
			"dev.example": "example/dev",
			"test":        "test/main",
			"dev.test":    "test/dev",
		},
	}

	handler, err := sitehandler.NewHandler(context.Background(), zap.NewNop(),
		domainResolver, siteResolver, sitehandler.HandlerConfig{HostPattern: hostPattern})
	assert.NoError(t, err)

	resolve := func(host string) any {
		desc, err := handler.ResolveSite(host)
		if errors.Is(err, site.ErrSiteNotFound) {
			return nil
		} else if err != nil {
			panic(err)
		}
		return desc
	}

	assert.Equal(t, resolve("example.com"), &site.Descriptor{ID: "example/main"})
	assert.Equal(t, resolve("example.com:8001"), &site.Descriptor{ID: "example/main"})
	assert.Equal(t, resolve("dev.example.com"), &site.Descriptor{ID: "example/dev"})
	assert.Equal(t, resolve("example.local"), nil)

	assert.Equal(t, resolve("test.pageship.local"), &site.Descriptor{ID: "test/main"})
	assert.Equal(t, resolve("dev.test.pageship.local"), &site.Descriptor{ID: "test/dev"})
	assert.Equal(t, resolve("dev.test.pageship.local:8001"), &site.Descriptor{ID: "test/dev"})
	assert.Equal(t, resolve("staging.test.pageship.local"), nil)
	assert.Equal(t, resolve("pageship.local"), nil)
	assert.Equal(t, resolve("main.pageship.local"), nil)
}

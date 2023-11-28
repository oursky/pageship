package local

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/site"
)

type resolverStatic struct {
	fs          fs.FS
	defaultSite string
	sites       map[string]config.SitesConfigEntry
}

func (h *resolverStatic) Kind() string { return "static config" }

func (h *resolverStatic) IsWildcard() bool { return false }

func (h *resolverStatic) Resolve(ctx context.Context, matchedID string) (*site.Descriptor, error) {
	if !site.CheckDefaultSite(&matchedID, h.defaultSite) {
		return nil, site.ErrSiteNotFound
	}

	entry, ok := h.sites[matchedID]
	if !ok {
		return nil, site.ErrSiteNotFound
	}

	fsys, err := fs.Sub(h.fs, entry.Context)
	if err != nil {
		return nil, fmt.Errorf("make context fs: %w", err)
	}

	if err := checkSiteFS(fsys); err != nil {
		return nil, fmt.Errorf("check context fs: %w", err)
	}

	config, err := loadConfig(fsys)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &site.Descriptor{
		ID:     matchedID,
		Domain: entry.Domain,
		Config: &config.Site,
		FS:     siteFS{fs: fsys},
	}, nil
}

package local

import (
	"fmt"
	"io/fs"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/handler/site"
)

type resolverStatic struct {
	fs    fs.FS
	sites map[string]config.SitesConfigEntry
}

func (h *resolverStatic) Kind() string { return "static config" }

func (h *resolverStatic) Resolve(matchedID string) (*site.Descriptor, error) {
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
		AppID:    config.ID,
		SiteName: matchedID,
		Files:    nil,
		Config:   &config.Site,
		FS:       fsys,
	}, nil
}

package local

import (
	"fmt"
	"io/fs"

	"github.com/oursky/pageship/internal/handler/site"
)

type resolverSingle struct {
	fs fs.FS
}

func NewSingleSiteResolver(fs fs.FS) site.Resolver {
	return &resolverSingle{fs: fs}
}

func (h *resolverSingle) Kind() string { return "single site" }

func (h *resolverSingle) Resolve(matchedID string) (*site.Descriptor, error) {
	if err := checkSiteFS(h.fs); err != nil {
		return nil, fmt.Errorf("check context fs: %w", err)
	}

	config, err := loadConfig(h.fs)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &site.Descriptor{
		AppID:    config.ID,
		SiteName: matchedID,
		Files:    nil,
		Config:   &config.Site,
		FS:       h.fs,
	}, nil
}

package local

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/oursky/pageship/internal/handler/site"
	"github.com/spf13/viper"
)

type resolverAdhoc struct {
	fs          fs.FS
	defaultSite string
}

func (h *resolverAdhoc) Kind() string { return "ad-hoc" }

func (h *resolverAdhoc) Resolve(ctx context.Context, matchedID string) (*site.Descriptor, error) {
	if !site.CheckDefaultSite(&matchedID, h.defaultSite) {
		return nil, site.ErrSiteNotFound
	}

	dnsLabels := strings.Split(path.Clean(matchedID), ".")
	pathSegments := make([]string, len(dnsLabels))
	for i, label := range dnsLabels {
		pathSegments[len(pathSegments)-1-i] = label
	}

	fsys, err := fs.Sub(h.fs, path.Join(pathSegments...))
	if err != nil {
		return nil, fmt.Errorf("make context fs: %w", err)
	}

	if err := checkSiteFS(fsys); err != nil {
		if os.IsNotExist(err) {
			return nil, site.ErrSiteNotFound
		}
		return nil, fmt.Errorf("check context fs: %w", err)
	}

	config, err := loadConfig(fsys)
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		// Explicitly handle config not found for ad-hoc sites
		return nil, site.ErrSiteNotFound
	} else if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &site.Descriptor{
		ID:     matchedID,
		Config: &config.Site,
		FS:     fsys,
	}, nil
}

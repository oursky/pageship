package site

import (
	"context"
	"errors"
	"io/fs"

	"github.com/oursky/pageship/internal/config"
)

var ErrSiteNotFound = errors.New("site not found")

type Descriptor struct {
	ID     string
	Config *config.SiteConfig
	FS     fs.FS
}

type Resolver interface {
	Kind() string
	Resolve(ctx context.Context, matchedID string) (*Descriptor, error)
}

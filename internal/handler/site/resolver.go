package site

import (
	"errors"
	"io/fs"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
)

var ErrSiteNotFound = errors.New("site not found")

type Descriptor struct {
	AppID    string
	SiteName string
	Files    []models.FileEntry
	Config   *config.SiteConfig
	FS       fs.FS
}

type Resolver interface {
	Kind() string
	Resolve(matchedID string) (*Descriptor, error)
}

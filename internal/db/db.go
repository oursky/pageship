package db

import (
	"context"
	"fmt"
	neturl "net/url"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
)

type Factory func(url neturl.URL) (DB, error)

var FactoryMap = make(map[string]Factory)

func New(url string) (DB, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}

	factory, ok := FactoryMap[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("unknown scheme: %s", u.Scheme)
	}

	return factory(*u)
}

type DB interface {
	BeginTx(ctx context.Context) (Conn, error)
}

type Conn interface {
	Tx() *sqlx.Tx
	AppsDB
}

type AppsDB interface {
	CreateApp(ctx context.Context, id string) (*models.App, error)
	GetApp(ctx context.Context, id string) (*models.App, error)
	ListApps(ctx context.Context) ([]*models.App, error)
	UpdateAppConfig(ctx context.Context, id string, config *config.AppConfig) (*models.App, error)

	EnsureSite(ctx context.Context, appID string, siteName string) (*models.Site, error)

	CreateDeployment(
		ctx context.Context,
		appID string,
		siteID string,
		storageKeyPrefix string,
		files []models.FileEntry,
		config *config.SiteConfig,
	) (*models.Deployment, error)
}

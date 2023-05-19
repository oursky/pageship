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
	CreateApp(ctx context.Context, app *models.App) error
	GetApp(ctx context.Context, id string) (*models.App, error)
	ListApps(ctx context.Context) ([]*models.App, error)
	UpdateAppConfig(ctx context.Context, id string, config *config.AppConfig) (*models.App, error)

	CreateSiteIfNotExist(ctx context.Context, site *models.Site) error

	CreateDeployment(ctx context.Context, deployment *models.Deployment) error
	GetDeployment(ctx context.Context, appID string, siteName string, id string) (*models.Deployment, error)
}

package db

import (
	"context"
	"fmt"
	neturl "net/url"
	"time"

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
	SitesDB
	DeploymentsDB
}

type AppsDB interface {
	CreateApp(ctx context.Context, app *models.App) error
	GetApp(ctx context.Context, id string) (*models.App, error)
	ListApps(ctx context.Context) ([]*models.App, error)
	UpdateAppConfig(ctx context.Context, id string, config *config.AppConfig) (*models.App, error)
}

type SitesDB interface {
	CreateSiteIfNotExist(ctx context.Context, site *models.Site) error
	ListSites(ctx context.Context, appID string) ([]SiteInfo, error)
}

type DeploymentsDB interface {
	CreateDeployment(ctx context.Context, deployment *models.Deployment) error
	ListDeployments(ctx context.Context, appID string, siteName string) ([]*models.Deployment, error)
	GetDeployment(ctx context.Context, appID string, siteName string, id string) (*models.Deployment, error)
	MarkDeploymentUploaded(ctx context.Context, now time.Time, deployment *models.Deployment) error
	ActivateSiteDeployment(ctx context.Context, deployment *models.Deployment) error
	DeactivateSiteDeployment(ctx context.Context, deployment *models.Deployment) error
	GetActiveSiteDeployment(ctx context.Context, appID string, siteName string) (*models.Deployment, error)
}

type SiteInfo struct {
	*models.Site
	ActiveDeploymentID *string    `db:"deployment_id"`
	LastDeployedAt     *time.Time `db:"last_deployed_at"`
}

package db

import (
	"context"
	"fmt"
	neturl "net/url"
	"time"

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
	Rollback() error
	Commit() error

	AppsDB
	SitesDB
	DeploymentsDB
	UserDB
}

type AppsDB interface {
	CreateApp(ctx context.Context, app *models.App) error
	GetApp(ctx context.Context, id string) (*models.App, error)
	ListApps(ctx context.Context, userID string) ([]*models.App, error)
	UpdateAppConfig(ctx context.Context, id string, config *config.AppConfig) (*models.App, error)
}

type SitesDB interface {
	CreateSiteIfNotExist(ctx context.Context, site *models.Site) (*SiteInfo, error)
	GetSiteByName(ctx context.Context, appID string, siteName string) (*models.Site, error)
	GetSiteInfo(ctx context.Context, appID string, id string) (*SiteInfo, error)
	ListSitesInfo(ctx context.Context, appID string) ([]SiteInfo, error)
}

type DeploymentsDB interface {
	CreateDeployment(ctx context.Context, deployment *models.Deployment) error
	GetDeployment(ctx context.Context, appID string, id string) (*models.Deployment, error)
	GetDeploymentByName(ctx context.Context, appID string, name string) (*models.Deployment, error)
	ListDeployments(ctx context.Context, appID string) ([]DeploymentInfo, error)
	MarkDeploymentUploaded(ctx context.Context, now time.Time, deployment *models.Deployment) error
	AssignDeploymentSite(ctx context.Context, deployment *models.Deployment, siteID string) error
	UnassignDeploymentSite(ctx context.Context, deployment *models.Deployment, siteID string) error
	GetSiteDeployment(ctx context.Context, appID string, siteName string) (*models.Deployment, error)
	GetDeploymentSiteNames(ctx context.Context, deployment *models.Deployment) ([]string, error)
	SetDeploymentExpiry(ctx context.Context, deployment *models.Deployment) error
	DeleteExpiredDeployments(ctx context.Context, now time.Time, expireBefore time.Time) (int64, error)
}

type UserDB interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByCredential(ctx context.Context, id models.UserCredentialID) (*models.User, error)
	CreateUserWithCredential(ctx context.Context, user *models.User, credential *models.UserCredential) error
	AssignAppUser(ctx context.Context, appID string, userID string) error
	UnassignAppUser(ctx context.Context, appID string, userID string) error
	ListAppUsers(ctx context.Context, appID string) ([]*models.User, error)
	IsAppAccessible(ctx context.Context, appID string, userID string) error
}

type DeploymentInfo struct {
	*models.Deployment
	FirstSiteName *string `db:"site_name"`
}

type SiteInfo struct {
	*models.Site
	DeploymentName *string `db:"deployment_name"`
}

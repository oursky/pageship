package db

import (
	"context"
	"fmt"
	neturl "net/url"
	"time"

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
	BeginTx(ctx context.Context) (Tx, error)
	Locker(ctx context.Context) (LockerDB, error)
	DBQuery
}

type Tx interface {
	Rollback() error
	Commit() error
	DBQuery
}

type DBQuery interface {
	AppsDB
	SitesDB
	DeploymentsDB
	DomainsDB
	DomainVerificationDB
	UserDB
	CertificateDB
}

type AppsDB interface {
	CreateApp(ctx context.Context, app *models.App) error
	GetApp(ctx context.Context, id string) (*models.App, error)
	ListApps(ctx context.Context, credentialIDs []models.CredentialID) ([]*models.App, error)
	UpdateAppConfig(ctx context.Context, app *models.App) error
}

type SitesDB interface {
	CreateSiteIfNotExist(ctx context.Context, site *models.Site) (*SiteInfo, error)
	GetSiteByName(ctx context.Context, appID string, siteName string) (*models.Site, error)
	GetSiteInfo(ctx context.Context, appID string, id string) (*SiteInfo, error)
	ListSitesInfo(ctx context.Context, appID string) ([]SiteInfo, error)
	SetSiteDeployment(ctx context.Context, site *models.Site) error
}

type DeploymentsDB interface {
	CreateDeployment(ctx context.Context, deployment *models.Deployment) error
	GetDeployment(ctx context.Context, appID string, id string) (*models.Deployment, error)
	GetDeploymentByName(ctx context.Context, appID string, name string) (*models.Deployment, error)
	ListDeployments(ctx context.Context, appID string) ([]DeploymentInfo, error)
	MarkDeploymentUploaded(ctx context.Context, now time.Time, deployment *models.Deployment) error
	GetSiteDeployment(ctx context.Context, appID string, siteName string) (*models.Deployment, error)
	GetDeploymentSiteNames(ctx context.Context, deployment *models.Deployment) ([]string, error)
	SetDeploymentExpiry(ctx context.Context, deployment *models.Deployment) error
	DeleteExpiredDeployments(ctx context.Context, now time.Time, expireBefore time.Time) (int64, error)
}

type DomainsDB interface {
	CreateDomain(ctx context.Context, domain *models.Domain) error
	GetDomainByName(ctx context.Context, domain string) (*models.Domain, error)
	GetDomainBySite(ctx context.Context, appID string, siteName string) (*models.Domain, error)
	DeleteDomain(ctx context.Context, id string, now time.Time) error
	ListDomains(ctx context.Context, appID string) ([]*models.Domain, error)
}

type DomainVerificationDB interface {
	CreateDomainVerification(ctx context.Context, domainVerification *models.DomainVerification) error
	ScheduleDomainVerificationAt(ctx context.Context, id string, time time.Time) error
	GetDomainVerificationByName(ctx context.Context, domain string, appID string) (*models.DomainVerification, error)
	DeleteDomainVerification(ctx context.Context, id string, now time.Time) error
	ListDomainVerifications(ctx context.Context, appID string) ([]*models.DomainVerification, error)
	ListLeastRecentlyCheckedDomain(ctx context.Context, now time.Time, isVerified bool, count uint) ([]*models.DomainVerification, error)
	LabelDomainVerificationAsVerified(ctx context.Context, id string, now time.Time, nextVerifyAt time.Time) error
	LabelDomainVerificationAsInvalid(ctx context.Context, id string, now time.Time) error
}

type UserDB interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetCredential(ctx context.Context, id models.CredentialID) (*models.UserCredential, error)
	CreateUser(ctx context.Context, user *models.User) error
	AddCredential(ctx context.Context, credential *models.UserCredential) error
	UpdateCredentialData(ctx context.Context, cred *models.UserCredential) error
	ListCredentialIDs(ctx context.Context, userID string) ([]models.CredentialID, error)
}

type CertificateDB interface {
	GetCertDataEntry(ctx context.Context, key string) (*models.CertDataEntry, error)
	SetCertDataEntry(ctx context.Context, entry *models.CertDataEntry) error
	DeleteCertificateData(ctx context.Context, key string) error
	ListCertificateData(ctx context.Context, prefix string) ([]string, error)
}

type LockerDB interface {
	Close() error
	Lock(ctx context.Context, name string) error
	Unlock(ctx context.Context, name string) error
}

type DeploymentInfo struct {
	*models.Deployment
	FirstSiteName *string `db:"site_name"`
}

type SiteInfo struct {
	*models.Site
	DeploymentName *string `db:"deployment_name"`
}

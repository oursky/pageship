package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/site"
	"github.com/oursky/pageship/internal/storage"
)

type Resolver struct {
	DB           db.DB
	Storage      *storage.Storage
	HostIDScheme config.HostIDScheme
}

func (r *Resolver) Kind() string { return "database" }

func (r *Resolver) resolveDeployment(
	ctx context.Context,
	app *models.App,
	siteName string,
) (*models.Deployment, string, error) {
	deployment, err := r.DB.GetSiteDeployment(ctx, app.ID, siteName)

	if deployment != nil {
		// Site found, check site name
		_, ok := app.Config.ResolveSite(siteName)
		if !ok {
			return nil, "", site.ErrSiteNotFound
		}
	}

	if errors.Is(err, models.ErrDeploymentNotFound) {
		// Site not found, check deployment with same name
		deploymentName := siteName
		siteName = ""

		if len(app.Config.Deployments.Access) == 0 {
			return nil, "", site.ErrSiteNotFound
		}

		deployment, err = r.DB.GetDeploymentByName(ctx, app.ID, deploymentName)
		if errors.Is(err, models.ErrDeploymentNotFound) {
			return nil, "", site.ErrSiteNotFound
		} else if err != nil {
			return nil, "", err
		}

		sites, cerr := r.DB.GetDeploymentSiteNames(ctx, deployment)
		if cerr != nil {
			return nil, "", cerr
		}
		if len(sites) > 0 {
			// Deployments assigned to site must be accessed through site
			return nil, "", site.ErrSiteNotFound
		}

	} else if err != nil {
		return nil, "", err
	}

	return deployment, siteName, nil
}

func (h *Resolver) IsWildcard() bool { return false }

func (r *Resolver) Resolve(ctx context.Context, matchedID string) (*site.Descriptor, error) {
	appID, siteName := r.HostIDScheme.Split(matchedID)

	app, err := r.DB.GetApp(ctx, appID)
	if errors.Is(err, models.ErrAppNotFound) {
		return nil, site.ErrSiteNotFound
	}

	if !site.CheckDefaultSite(&siteName, app.Config.DefaultSite) {
		return nil, site.ErrSiteNotFound
	}

	deployment, siteName, err := r.resolveDeployment(ctx, app, siteName)
	if errors.Is(err, models.ErrDeploymentNotFound) {
		return nil, site.ErrSiteNotFound
	} else if err != nil {
		return nil, err
	}

	if err := deployment.CheckAlive(time.Now().UTC()); err != nil {
		return nil, site.ErrSiteNotFound
	}

	config := deployment.Metadata.Config
	if siteName == "" {
		// Not assigned to site; use preview deployment access rules
		config.Access = app.Config.Deployments.Access
	}

	domain, err := r.DB.GetDomainBySite(ctx, app.ID, siteName)
	if errors.Is(err, models.ErrDomainNotFound) {
		domain = nil
	} else if err != nil {
		return nil, err
	}

	domainName := ""
	if domain != nil {
		domainName = domain.Domain
	}

	id := strings.Join([]string{deployment.AppID, siteName, deployment.ID}, "/")
	desc := &site.Descriptor{
		ID:     id,
		Config: &config,
		Domain: domainName,
		FS:     newStorageFS(r.Storage, deployment),
	}

	return desc, nil
}

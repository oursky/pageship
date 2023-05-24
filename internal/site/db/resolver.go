package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/site"
	"github.com/oursky/pageship/internal/storage"
)

type resolver struct {
	db      db.DB
	storage *storage.Storage
}

func NewResolver(db db.DB, storage *storage.Storage) site.Resolver {
	return &resolver{db: db, storage: storage}
}

func (r *resolver) Kind() string { return "database" }

func (r *resolver) resolveDeployment(
	ctx context.Context,
	c db.Conn,
	app *models.App,
	siteName string,
) (*models.Deployment, error) {
	deployment, err := c.GetSiteDeployment(ctx, app.ID, siteName)

	if deployment != nil {
		// Site found, check site name
		_, ok := app.Config.ResolveSite(siteName)
		if !ok {
			return nil, site.ErrSiteNotFound
		}
	}

	if errors.Is(err, models.ErrDeploymentNotFound) {
		// Site not found, check deployment with same name
		if app.Config.Deployments.Accessible {
			deployment, err = c.GetDeploymentByName(ctx, app.ID, siteName)

			sites, cerr := c.GetDeploymentSiteNames(ctx, deployment)
			if cerr != nil {
				return nil, cerr
			}
			if len(sites) > 0 {
				// Deployments assigned to site must be accessed through site
				return nil, site.ErrSiteNotFound
			}
		}
	}

	if errors.Is(err, models.ErrDeploymentNotFound) {
		return nil, site.ErrSiteNotFound
	} else if err != nil {
		return nil, err
	}
	return deployment, nil
}

func (r *resolver) Resolve(ctx context.Context, matchedID string) (*site.Descriptor, error) {
	siteName, appID, hasSite := strings.Cut(matchedID, ".")
	if !hasSite {
		appID = siteName
		siteName = ""
	}

	var desc *site.Descriptor
	err := db.WithTx(ctx, r.db, func(c db.Conn) error {
		app, err := c.GetApp(ctx, appID)
		if errors.Is(err, models.ErrAppNotFound) {
			return site.ErrSiteNotFound
		}

		if !site.CheckDefaultSite(&siteName, app.Config.DefaultSite) {
			return site.ErrSiteNotFound
		}

		deployment, err := r.resolveDeployment(ctx, c, app, siteName)
		if errors.Is(err, models.ErrDeploymentNotFound) {
			return site.ErrSiteNotFound
		} else if err != nil {
			return err
		}

		if err := deployment.CheckAlive(time.Now().UTC()); err != nil {
			return site.ErrSiteNotFound
		}

		id := strings.Join([]string{deployment.AppID, siteName, deployment.ID}, "/")
		desc = &site.Descriptor{
			ID:     id,
			Config: &deployment.Metadata.Config,
			FS:     newStorageFS(r.storage, deployment),
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return desc, nil
}

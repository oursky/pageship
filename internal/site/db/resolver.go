package db

import (
	"context"
	"errors"
	"strings"

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

		_, ok := app.Config.ResolveSite(siteName)
		if !ok {
			return site.ErrSiteNotFound
		}

		deployment, err := c.GetSiteDeployment(ctx, appID, siteName)
		if errors.Is(err, models.ErrDeploymentNotFound) {
			return site.ErrSiteNotFound
		} else if err != nil {
			return err
		}

		if deployment.UploadedAt == nil {
			return errors.New("deployment not yet uploaded")
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
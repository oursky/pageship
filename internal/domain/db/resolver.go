package db

import (
	"context"
	"errors"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/domain"
	"github.com/oursky/pageship/internal/models"
)

type Resolver struct {
	DB           db.DB
	HostIDScheme config.HostIDScheme
}

func (r *Resolver) Kind() string { return "database" }

func (r *Resolver) Resolve(ctx context.Context, hostname string) (string, error) {
	dom, err := r.DB.GetDomainByName(ctx, hostname)
	if errors.Is(err, models.ErrDomainNotFound) {
		return "", domain.ErrDomainNotFound
	}

	app, err := r.DB.GetApp(ctx, dom.AppID)
	if errors.Is(err, models.ErrAppNotFound) {
		return "", domain.ErrDomainNotFound
	}

	sub := dom.SiteName
	if dom.SiteName == app.Config.DefaultSite {
		sub = ""
	}
	id := r.HostIDScheme.Make(dom.AppID, sub)

	return id, nil
}

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/models"
)

func (q query[T]) CreateDomain(ctx context.Context, domain *models.Domain) error {
	result, err := sqlx.NamedExecContext(ctx, q.ext, `
		INSERT INTO domain_association (id, created_at, updated_at, deleted_at, domain, app_id, site_name)
			VALUES (:id, :created_at, :updated_at, :deleted_at, :domain, :app_id, :site_name)
			ON CONFLICT (domain) WHERE deleted_at IS NULL DO NOTHING
	`, domain)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return models.ErrDomainUsedName
	}

	return nil
}

func (q query[T]) GetDomainByName(ctx context.Context, domainName string) (*models.Domain, error) {
	var domain models.Domain

	err := sqlx.GetContext(ctx, q.ext, &domain, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.domain, d.app_id, d.site_name FROM domain_association d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			WHERE d.domain = $1 AND d.deleted_at IS NULL
	`, domainName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDomainNotFound
	} else if err != nil {
		return nil, err
	}

	return &domain, nil
}

func (q query[T]) GetDomainBySite(ctx context.Context, appID string, siteName string) (*models.Domain, error) {
	var domain models.Domain

	err := sqlx.GetContext(ctx, q.ext, &domain, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.domain, d.app_id, d.site_name FROM domain_association d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			WHERE d.app_id = $1 AND d.site_name = $2 AND d.deleted_at IS NULL
	`, appID, siteName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDomainNotFound
	} else if err != nil {
		return nil, err
	}

	return &domain, nil
}

func (q query[T]) DeleteDomain(ctx context.Context, id string, now time.Time) error {
	_, err := q.ext.ExecContext(ctx, `
		UPDATE domain_association SET deleted_at = $1 WHERE id = $2
	`, now, id)
	if err != nil {
		return err
	}

	return nil
}

func (q query[T]) ListDomains(ctx context.Context, appID string) ([]*models.Domain, error) {
	var domains []*models.Domain
	err := sqlx.SelectContext(ctx, q.ext, &domains, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.domain, d.app_id, d.site_name FROM domain_association d
			WHERE d.app_id = $1 AND d.deleted_at IS NULL
			ORDER BY d.domain, d.created_at
	`, appID)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

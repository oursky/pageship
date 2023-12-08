package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/models"
)

func (q query[T]) CreateDomainVerification(ctx context.Context, domainVerification *models.DomainVerification) error {
	result, err := sqlx.NamedExecContext(ctx, q.ext, `
		INSERT INTO domain_verification (id, created_at, updated_at, deleted_at, domain, app_id, value, verified_at, domain_prefix)
        VALUES (:id, :created_at, :updated_at, :deleted_at, :domain, :app_id, :value, :verified_at, :domain_prefix)
			ON CONFLICT (domain, app_id) WHERE deleted_at IS NULL DO NOTHING
	`, domainVerification)
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

func (q query[T]) GetDomainVerificationByName(ctx context.Context, domainName string) (*models.DomainVerification, error) {
	var domainVerification models.DomainVerification

	err := sqlx.GetContext(ctx, q.ext, &domainVerification, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.domain, d.app_id, d.value, d.verified_at, d.domain_prefix FROM domain_verification d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			WHERE d.domain = ? AND d.deleted_at IS NULL
	`, domainName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDomainNotFound
	} else if err != nil {
		return nil, err
	}

	return &domainVerification, nil

}

func (q query[T]) DeleteDomainVerification(ctx context.Context, id string, now time.Time) error {
	_, err := q.ext.ExecContext(ctx, `
		UPDATE domain_verification SET deleted_at = ? WHERE id = ?
	`, now, id)
	if err != nil {
		return err
	}

	return nil
}

func (q query[T]) ListDomainVerifications(ctx context.Context, appID *string, count *uint, isVerified *bool) ([]*models.DomainVerification, error) {
	var domainVerifications []*models.DomainVerification
	stmt := `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.domain, d.app_id, d.value, d.verified_at, d.domain_prefix
            FROM domain_verification d
			WHERE %s
			ORDER BY d.domain, d.updated_at
	`
	where := "d.deleted_at IS NULL"
	if appID != nil {
		where += " AND d.app_id = ?"
	}
	if isVerified != nil {
		if *isVerified {
			where += " AND d.verified_at IS NOT NULL"
		} else {
			where += " AND d.verified_at IS NULL"
		}
	}
	stmt = fmt.Sprintf(stmt, where)
	if count != nil {
		stmt += fmt.Sprintf(" LIMIT %d", *count)
	}
	err := sqlx.SelectContext(ctx, q.ext, &domainVerifications, stmt, appID)
	if err != nil {
		return nil, err
	}
	return domainVerifications, err
}

func (q query[T]) UpdateDomainVerification(ctx context.Context, domainVerification *models.DomainVerification) error {
	_, err := sqlx.NamedExecContext(ctx, q.ext, `
    UPDATE domain_verification SET created_at = :created_at,
    updated_at = :updated_at,
    deleted_at = :deleted_at,
    domain = :domain,
    app_id = :app_id,
    value = :value,
    verified_at = :verified_at,
    domain_prefix = :domain_prefix
    where id = :id
    `, domainVerification)
	return err
}

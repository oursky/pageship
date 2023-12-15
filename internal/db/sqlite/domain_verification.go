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
		INSERT INTO domain_verification (id, created_at, updated_at, deleted_at, domain, app_id, value, verified_at, domain_prefix, will_check_at, last_checked_at)
        VALUES (:id, :created_at, :updated_at, :deleted_at, :domain, :app_id, :value, :verified_at, :domain_prefix, :will_check_at, :last_checked_at)
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

func (q query[T]) GetDomainVerificationByName(ctx context.Context, domainName string, appId string) (*models.DomainVerification, error) {
	var domainVerification models.DomainVerification

	err := sqlx.GetContext(ctx, q.ext, &domainVerification, `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.verified_at, d.last_checked_at, d.will_check_at, d.domain, d.domain_prefix, d.app_id, d.value,
        d.verified_at, d.domain_prefix, d.will_check_at, d.last_checked_at
        FROM domain_verification d
			JOIN app a ON (a.id = d.app_id AND a.deleted_at IS NULL)
			WHERE d.domain = ? AND d.deleted_at IS NULL AND d.app_id = ?
	`, domainName, appId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDomainNotFound
	} else if err != nil {
		return nil, err
	}

	return &domainVerification, nil

}

func (q query[T]) DeleteDomainVerification(ctx context.Context, id string, now time.Time) error {
	_, err := q.ext.ExecContext(ctx, `
		UPDATE domain_verification SET deleted_at = ?, updated_at = ? WHERE id = ?
	`, now, now, id)
	if err != nil {
		return err
	}

	return nil
}

func (q query[T]) ListDomainVerifications(ctx context.Context, appID string) ([]*models.DomainVerification, error) {
	var domainVerifications []*models.DomainVerification
	stmt := `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.verified_at, d.last_checked_at, d.will_check_at, d.domain, d.domain_prefix, d.app_id, d.value,
            d.verified_at, d.domain_prefix, d.will_check_at, d.last_checked_at
            FROM domain_verification d
			WHERE d.deleted_at IS NULL AND d.app_id = ?
			ORDER BY d.domain, d.created_at
	`
	err := sqlx.SelectContext(ctx, q.ext, &domainVerifications, stmt, appID)
	if err != nil {
		return nil, err
	}
	return domainVerifications, err
}

func (q query[T]) LabelDomainVerificationAsVerified(ctx context.Context, id string, now time.Time, nextVerifyAt time.Time) error {
	_, err := q.ext.ExecContext(ctx, `
    UPDATE domain_verification SET
        updated_at = ?,
        verified_at = ?,
        last_checked_at = ?,
        will_check_at = ?
        WHERE id = ?
    `, now, now, now, nextVerifyAt, id)
	return err
}

func (q query[T]) LabelDomainVerificationAsInvalid(ctx context.Context, id string, now time.Time) error {
	_, err := q.ext.ExecContext(ctx, `
    UPDATE domain_verification SET
        updated_at = ?,
        verified_at = ?,
        last_checked_at = ?,
        will_check_at = ?
        WHERE id = ?
    `, now, nil, now, nil, id)
	return err
}

func (q query[T]) ListLeastRecentlyCheckedDomain(ctx context.Context, time time.Time, isVerified bool, count uint) ([]*models.DomainVerification, error) {
	var domainVerifications []*models.DomainVerification
	stmt := `
		SELECT d.id, d.created_at, d.updated_at, d.deleted_at, d.verified_at, d.last_checked_at, d.will_check_at, d.domain, d.domain_prefix, d.app_id, d.value,
            d.verified_at, d.domain_prefix, d.will_check_at, d.last_checked_at
            FROM domain_verification d
			WHERE %s
            ORDER BY d.will_check_at, d.last_checked_at NULLS FIRST
            LIMIT ?
	`
	where := "d.deleted_at IS NULL AND d.will_check_at IS NOT NULL AND d.will_check_at <= ?"
	if isVerified {
		where += " AND d.verified_at IS NOT NULL"
	} else {
		where += " AND d.verified_at IS NULL"
	}
	stmt = fmt.Sprintf(stmt, where)
	err := sqlx.SelectContext(ctx, q.ext, &domainVerifications, stmt, time, count)
	if err != nil {
		return nil, err
	}
	return domainVerifications, err
}

func (q query[T]) ScheduleDomainVerificationAt(ctx context.Context, id string, time time.Time) error {
	_, err := q.ext.ExecContext(ctx, `
    UPDATE domain_verification SET
        updated_at = ?,
        will_check_at = ?
        WHERE id = ?
    `, time, time, id)
	return err
}

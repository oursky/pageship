package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

var patternReplacer = strings.NewReplacer("_", "\\_", "%", "\\%")

const lockTimeout time.Duration = 10 * time.Minute

func (c DB) GetCertDataEntry(ctx context.Context, key string) (*models.CertDataEntry, error) {
	var entry models.CertDataEntry
	err := c.db.GetContext(ctx, &entry, `
		SELECT key, updated_at, value FROM cert_data WHERE key = ?
	`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrCertificateDataNotFound
	} else if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (c DB) SetCertDataEntry(ctx context.Context, entry *models.CertDataEntry) error {
	_, err := c.db.NamedExecContext(ctx, `
		INSERT INTO cert_data (key, updated_at, value)
			VALUES (:key, :updated_at, :value)
			ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, entry)
	if err != nil {
		return err
	}

	return nil
}

func (c DB) DeleteCertificateData(ctx context.Context, key string) error {
	result, err := c.db.ExecContext(ctx, `
		DELETE FROM cert_data WHERE key = ?
	`, key)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return models.ErrCertificateDataNotFound
	}

	return nil
}

func (c DB) ListCertificateData(ctx context.Context, prefix string) ([]string, error) {
	pattern := patternReplacer.Replace(prefix) + "%"
	var keys []string
	err := c.db.SelectContext(ctx, &keys, `
		SELECT key FROM cert_data WHERE key LIKE ? ESCAPE '\'
	`, pattern)
	if err != nil {
		return nil, err
	}

	// SQLite LIKE is case-insensitive, filter out extra keys
	n := 0
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) {
			keys[n] = k
			n++
		}
	}
	keys = keys[:n]

	return keys, nil
}
func (c DB) Locker(ctx context.Context) (db.CertificateDBLocker, error) {
	id := models.RandomID(8)
	return &locker{id: id, db: c.db}, nil
}

type locker struct {
	id string
	db *sqlx.DB
}

func (l *locker) Close() error {
	_, err := l.db.Exec("DELETE FROM cert_lock WHERE holder = ?", l.id)
	return err
}

func (l *locker) Lock(ctx context.Context, name string) error {
	releaseAt := time.Now().UTC().Add(lockTimeout)
	result, err := l.db.ExecContext(ctx, `
		INSERT INTO cert_lock (name, holder, release_at)
		VALUES (?, ?, ?)
		ON CONFLICT (name) DO UPDATE
			SET holder = excluded.holder, release_at = excluded.release_at
			WHERE release_at < DATETIME('now')
	`, name, l.id, releaseAt)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	} else if n == 0 {
		return models.ErrCertificateDataLocked
	}

	return err
}

func (l *locker) Unlock(ctx context.Context, name string) error {
	_, err := l.db.ExecContext(ctx, `
		DELETE FROM cert_lock WHERE name = ? AND holder = ?
	`, name, l.id)
	return err
}

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"hash/fnv"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

func (c DB) GetCertDataEntry(ctx context.Context, key string) (*models.CertDataEntry, error) {
	var entry models.CertDataEntry
	err := c.db.GetContext(ctx, &entry, `
		SELECT key, updated_at, value FROM cert_data WHERE key = $1
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
		DELETE FROM cert_data WHERE key = $1
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
	var keys []string
	err := c.db.SelectContext(ctx, &keys, `
		SELECT key FROM cert_data WHERE starts_with(key, $1)
	`, prefix)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (c DB) Locker(ctx context.Context) (db.CertificateDBLocker, error) {
	id := models.RandomID(8)
	conn, err := c.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	return &locker{id: id, conn: conn}, nil
}

type locker struct {
	id   string
	conn *sqlx.Conn
}

func (l *locker) Close() error {
	return l.conn.Close()
}

func (l *locker) Lock(ctx context.Context, name string) error {
	key := lockKey(name)
	_, err := l.conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, key)
	if err != nil {
		return err
	}

	return nil
}

func (l *locker) Unlock(ctx context.Context, name string) error {
	key := lockKey(name)
	_, err := l.conn.ExecContext(ctx, `SELECT pg_advisory_unlock($1)`, key)
	if err != nil {
		return err
	}

	return nil
}

func lockKey(name string) int64 {
	h := fnv.New64a()
	h.Write([]byte(name))
	return int64(h.Sum64())
}

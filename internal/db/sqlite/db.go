package sqlite

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/db"
	_ "modernc.org/sqlite"
)

func init() {
	db.FactoryMap["sqlite"] = NewSqlite
}

type query[T sqlx.ExtContext] struct{ ext T }

type DB struct {
	query[*sqlx.DB]
}

type Tx struct {
	query[*sqlx.Tx]
}

func (t Tx) Rollback() error { return t.ext.Rollback() }
func (t Tx) Commit() error   { return t.ext.Commit() }

func NewSqlite(url url.URL) (db.DB, error) {
	q := url.Query()
	q.Set("_time_format", "sqlite")
	q.Add("_pragma", "busy_timeout=10000")
	q.Add("_pragma", "journal_mode=WAL")
	q.Add("_pragma", "foreign_keys=1")
	url.RawQuery = q.Encode()

	dsn := strings.TrimPrefix(url.String(), "sqlite://")
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite DB: %w", err)
	}

	return DB{query: query[*sqlx.DB]{ext: db}}, nil
}

func (db DB) BeginTx(ctx context.Context) (db.Tx, error) {
	tx, err := db.ext.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return Tx{query: query[*sqlx.Tx]{ext: tx}}, nil
}

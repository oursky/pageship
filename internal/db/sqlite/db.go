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

type DB struct {
	db *sqlx.DB
}

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

	return DB{db: db}, nil
}

func (db DB) BeginTx(ctx context.Context) (db.Conn, error) {
	tx, err := db.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return newConn(tx), nil
}

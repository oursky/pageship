package postgres

import (
	"context"
	"fmt"
	"net/url"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/db"
)

func init() {
	db.FactoryMap["postgres"] = NewPostgres
}

type DB struct {
	db *sqlx.DB
}

func NewPostgres(url url.URL) (db.DB, error) {
	db, err := sqlx.Open("pgx", url.String())
	if err != nil {
		return nil, fmt.Errorf("open postgres DB: %w", err)
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

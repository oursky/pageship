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

type query[T sqlx.ExtContext] struct{ ext T }

type DB struct {
	query[*sqlx.DB]
}

type Tx struct {
	query[*sqlx.Tx]
}

func (t Tx) Rollback() error { return t.ext.Rollback() }
func (t Tx) Commit() error   { return t.ext.Commit() }

func NewPostgres(url url.URL) (db.DB, error) {
	db, err := sqlx.Open("pgx", url.String())
	if err != nil {
		return nil, fmt.Errorf("open postgres DB: %w", err)
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

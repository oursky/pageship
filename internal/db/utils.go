package db

import (
	"context"
	"database/sql"
	"errors"
)

var ErrRollback = errors.New("rollback tx")

func WithTx(ctx context.Context, db DB, fn func(Conn) error) error {
	conn, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer conn.Tx().Rollback()

	err = fn(conn)
	if errors.Is(err, ErrRollback) {
		err = conn.Tx().Rollback()
	}
	if err != nil {
		return err
	}

	if err := conn.Tx().Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	return nil
}

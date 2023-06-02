package db

import (
	"context"
	"database/sql"
	"errors"
)

var ErrRollback = errors.New("rollback tx")

func WithTx(ctx context.Context, db DB, fn func(Tx) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = fn(tx)
	if errors.Is(err, ErrRollback) {
		err = tx.Rollback()
	}
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	return nil
}

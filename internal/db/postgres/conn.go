package postgres

import (
	"github.com/jmoiron/sqlx"
)

type Conn struct {
	tx *sqlx.Tx
}

func newConn(tx *sqlx.Tx) Conn {
	return Conn{
		tx: tx,
	}
}

func (c Conn) Rollback() error { return c.tx.Rollback() }
func (c Conn) Commit() error   { return c.tx.Commit() }

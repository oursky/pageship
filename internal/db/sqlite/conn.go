package sqlite

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

func (c Conn) Tx() *sqlx.Tx { return c.tx }

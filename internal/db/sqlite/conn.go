package sqlite

import (
	"github.com/jmoiron/sqlx"
	"github.com/oursky/pageship/internal/time"
)

type Conn struct {
	clock time.Clock
	tx    *sqlx.Tx
}

func newConn(tx *sqlx.Tx) Conn {
	return Conn{
		clock: time.SystemClock,
		tx:    tx,
	}
}

func (c Conn) Tx() *sqlx.Tx { return c.tx }

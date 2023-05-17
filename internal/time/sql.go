package time

import (
	"database/sql"
	"time"
)

func ToSqlNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Valid: true, Time: *t}
}

func FromSqlNullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

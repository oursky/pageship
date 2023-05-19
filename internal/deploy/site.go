package deploy

import (
	"context"

	"github.com/oursky/pageship/internal/db"
)

func EnsureSite(ctx context.Context, d db.DB, appID string, siteName string) error {
	return db.WithTx(ctx, d, func(c db.Conn) error {
		return nil
	})
}

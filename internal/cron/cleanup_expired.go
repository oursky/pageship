package cron

import (
	"context"

	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/time"
	"go.uber.org/zap"
)

type CleanupExpired struct {
	Clock            time.Clock
	Schedule         string
	KeepAfterExpired time.Duration
	DB               db.DB
}

func (c *CleanupExpired) Name() string { return "cleanup-expired" }

func (c *CleanupExpired) CronSchedule() string { return c.Schedule }

func (c *CleanupExpired) Run(ctx context.Context, logger *zap.Logger) error {
	clock := c.Clock
	if clock == nil {
		clock = time.SystemClock
	}
	now := clock.Now().UTC()
	expireBefore := now.Add(-c.KeepAfterExpired)

	return db.WithTx(ctx, c.DB, func(c db.Conn) error {
		n, err := c.DeleteExpiredDeployments(ctx, now, expireBefore)
		if err != nil {
			return err
		}

		logger.Info("deleted expired deployment", zap.Int64("n", n))
		return nil
	})
}

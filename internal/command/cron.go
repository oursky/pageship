package command

import (
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CronJob interface {
	Name() string
	CronSchedule() string
	Run(ctx context.Context, logger *zap.Logger) error
}

type CronRunner struct {
	Logger *zap.Logger
	Jobs   []CronJob
}

func (s *CronRunner) Run(ctx context.Context) error {
	c := cron.New()

	for _, job := range s.Jobs {
		schedule := job.CronSchedule()
		if schedule == "" {
			continue
		}

		jobName := job.Name()
		s.Logger.Info("schedule job", zap.String("job", jobName), zap.String("schedule", schedule))

		jobLogger := s.Logger.Named(jobName)
		c.AddFunc(schedule, func() {
			defer func() {
				if err := recover(); err != nil {
					if err != nil {
						jobLogger.Error("job panic", zap.Any("error", err))
					}
				}
			}()

			err := job.Run(ctx, jobLogger)
			if err != nil {
				jobLogger.Error("job failed", zap.Error(err))
			}
		})
	}

	c.Start()
	<-ctx.Done()
	<-c.Stop().Done()
	return nil
}

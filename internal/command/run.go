package command

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type WorkFunc func(context.Context) error

func Run(logger *zap.Logger, works []WorkFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	logger.Debug("starting...")
	for _, f := range works {
		work := f
		g.Go(func() error { return work(ctx) })
	}

	go func() {
		<-sig
		logger.Debug("stopping...")
		cancel()
	}()

	_ = g.Wait()
}

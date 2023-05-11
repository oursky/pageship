package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/oursky/pageship/internal/cmd"
	"go.uber.org/zap"
)

func serve(ctx context.Context) error {
	server := http.Server{
		Addr: ":8000",
	}

	shutdown := make(chan struct{})
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		close(shutdown)
	}()

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}
	<-shutdown

	return nil
}

func main() {
	cfg := zap.NewDevelopmentConfig()
	logger, _ := cfg.Build()
	defer logger.Sync()

	cmd.Run(logger, []cmd.WorkFunc{serve})
}

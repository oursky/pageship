package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/oursky/pageship/internal/command"
	"github.com/spf13/cobra"
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

var rootCmd = &cobra.Command{
	Use:   "start server",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Sugar().Infof("%+v", config)
		command.Run(logger, []command.WorkFunc{serve})
	},
}

func Execute() error {
	return rootCmd.Execute()
}

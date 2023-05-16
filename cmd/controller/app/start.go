package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/controller"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().String("addr", ":8001", "listen address")
	viper.BindPFlags(startCmd.PersistentFlags())
}

func startServer(ctx context.Context, addr string, handler http.Handler) error {
	server := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	shutdown := make(chan struct{})
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		close(shutdown)
	}()

	logger.Info("server starting", zap.String("addr", server.Addr))

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("failed to start server", zap.Error(err))
		return fmt.Errorf("failed to start server: %w", err)
	}
	<-shutdown

	return nil
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start controller server",
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("addr")
		if !debugMode {
			gin.SetMode(gin.ReleaseMode)
		}

		ctrl := &controller.Controller{}
		handler := ctrl.Handler()

		command.Run(logger, []command.WorkFunc{
			func(ctx context.Context) error {
				return startServer(ctx, addr, handler)
			},
		})
	},
}

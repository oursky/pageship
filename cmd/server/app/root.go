package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/serve"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func loadConfig(prefix string) (*config.ServerConfig, error) {
	v := viper.New()

	var root string
	v.SetConfigName("pageship")
	v.AddConfigPath(prefix)
	if err := v.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, err
		}
		root = prefix
	} else {
		root = filepath.Dir(v.ConfigFileUsed())
	}

	conf := config.DefaultServerConfig(root)
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func start(ctx context.Context) error {
	conf, err := loadConfig(cmdConfig.Prefix)
	if err != nil {
		logger.Error("failed to load config", zap.Error(err))
		return err
	}

	handler := serve.NewHandler(conf)

	server := http.Server{
		Addr:    ":8000",
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

	err = server.ListenAndServe()
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
		command.Run(logger, []command.WorkFunc{start})
	},
}

func Execute() error {
	return rootCmd.Execute()
}

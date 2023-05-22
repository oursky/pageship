package app

import (
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	_ "github.com/oursky/pageship/internal/db/sqlite"
	"github.com/oursky/pageship/internal/handler/site"
	sitedb "github.com/oursky/pageship/internal/handler/site/db"
	"github.com/oursky/pageship/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().String("database", "", "database URL")
	startCmd.MarkPersistentFlagRequired("database")

	startCmd.PersistentFlags().String("storage-endpoint", "", "object storage endpoint")
	startCmd.MarkPersistentFlagRequired("storage-endpoint")

	startCmd.PersistentFlags().String("host-pattern", config.DefaultHostPattern, "host match pattern")

	startCmd.PersistentFlags().String("addr", ":8000", "listen address")
}

type siteLogger struct {
	log *zap.Logger
}

func (l siteLogger) Debug(format string, args ...any) {
	l.log.Sugar().Debugf(format, args...)
}

func (l siteLogger) Error(msg string, err error) {
	l.log.Error(msg, zap.Error(err))
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		database := viper.GetString("database")
		addr := viper.GetString("addr")
		storageEndpoint := viper.GetString("storage-endpoint")
		hostPattern := viper.GetString("host-pattern")

		db, err := db.New(database)
		if err != nil {
			logger.Fatal("failed to setup database", zap.Error(err))
			return
		}

		storage, err := storage.New(cmd.Context(), storageEndpoint)
		if err != nil {
			logger.Fatal("failed to setup object storage", zap.Error(err))
			return
		}

		resolver := sitedb.NewResolver(db, storage)
		handler, err := site.NewHandler(
			siteLogger{log: logger.Named("site")},
			resolver,
			site.HandlerConfig{
				HostPattern: hostPattern,
			},
		)
		if err != nil {
			logger.Fatal("failed to setup server", zap.Error(err))
			return
		}

		server := command.HTTPServer{Addr: addr, Handler: handler}

		command.Run([]command.WorkFunc{
			server.Run,
		})
	},
}

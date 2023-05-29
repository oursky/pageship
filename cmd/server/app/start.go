package app

import (
	"net/http"

	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	_ "github.com/oursky/pageship/internal/db/postgres"
	_ "github.com/oursky/pageship/internal/db/sqlite"
	"github.com/oursky/pageship/internal/handler/site"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	sitedb "github.com/oursky/pageship/internal/site/db"
	"github.com/oursky/pageship/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().String("database-url", "", "database URL")
	startCmd.MarkPersistentFlagRequired("database-url")

	startCmd.PersistentFlags().String("storage-url", "", "object storage URL")
	startCmd.MarkPersistentFlagRequired("storage-url")

	startCmd.PersistentFlags().String("host-pattern", config.DefaultHostPattern, "host match pattern")
	startCmd.PersistentFlags().String("host-id-scheme", string(config.HostIDSchemeDefault), "host ID scheme")

	startCmd.PersistentFlags().String("addr", ":8000", "listen address")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		var cmdArgs struct {
			DatabaseURL  string              `mapstructure:"database-url" validate:"url"`
			StorageURL   string              `mapstructure:"storage-url" validate:"url"`
			Addr         string              `mapstructure:"addr" validate:"hostname_port"`
			HostPattern  string              `mapstructure:"host-pattern"`
			HostIDScheme config.HostIDScheme `mapstructure:"host-id-scheme" validate:"hostidscheme"`
		}
		if err := viper.Unmarshal(&cmdArgs); err != nil {
			logger.Fatal("invalid config", zap.Error(err))
			return
		}
		if err := validate.Struct(cmdArgs); err != nil {
			logger.Fatal("invalid config", zap.Error(err))
			return
		}

		db, err := db.New(cmdArgs.DatabaseURL)
		if err != nil {
			logger.Fatal("failed to setup database", zap.Error(err))
			return
		}

		storage, err := storage.New(cmd.Context(), cmdArgs.StorageURL)
		if err != nil {
			logger.Fatal("failed to setup object storage", zap.Error(err))
			return
		}

		resolver := &sitedb.Resolver{
			HostIDScheme: cmdArgs.HostIDScheme,
			DB:           db,
			Storage:      storage,
		}
		handler, err := site.NewHandler(
			zapLogger{logger.Named("site")},
			resolver,
			site.HandlerConfig{
				HostPattern: cmdArgs.HostPattern,
				Middlewares: middleware.Default,
			},
		)
		if err != nil {
			logger.Fatal("failed to setup server", zap.Error(err))
			return
		}

		server := command.HTTPServer{
			Logger: zapLogger{logger.Named("server")},
			Server: http.Server{
				Addr:    cmdArgs.Addr,
				Handler: handler,
			},
		}

		command.Run([]command.WorkFunc{
			server.Run,
		})
	},
}

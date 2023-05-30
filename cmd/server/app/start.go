package app

import (
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	_ "github.com/oursky/pageship/internal/db/postgres"
	_ "github.com/oursky/pageship/internal/db/sqlite"
	"github.com/oursky/pageship/internal/handler/site"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/httputil"
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
	startCmd.PersistentFlags().Bool("tls", false, "use TLS")
	startCmd.PersistentFlags().String("tls-addr", ":443", "TLS listen address")
	startCmd.PersistentFlags().String("tls-acme-endpoint", "", "TLS ACME directory endpoint")
	startCmd.PersistentFlags().String("tls-acme-email", "", "TLS ACME directory account email")
	startCmd.PersistentFlags().String("tls-protect-key", "", "TLS data protection key")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		var cmdArgs struct {
			DatabaseURL     string              `mapstructure:"database-url" validate:"url"`
			StorageURL      string              `mapstructure:"storage-url" validate:"url"`
			HostPattern     string              `mapstructure:"host-pattern"`
			HostIDScheme    config.HostIDScheme `mapstructure:"host-id-scheme" validate:"hostidscheme"`
			Addr            string              `mapstructure:"addr" validate:"hostname_port"`
			TLS             bool                `mapstructure:"tls"`
			TLSAddr         string              `mapstructure:"tls-addr" validate:"hostname_port"`
			TLSACMEEndpoint string              `mapstructure:"tls-acme-endpoint"`
			TLSACMEEmail    string              `mapstructure:"tls-acme-email"`
			TLSProtectKey   string              `mapstructure:"tls-protect-key"`
		}
		if err := viper.Unmarshal(&cmdArgs); err != nil {
			logger.Fatal("invalid config", zap.Error(err))
			return
		}
		if err := validate.Struct(cmdArgs); err != nil {
			logger.Fatal("invalid config", zap.Error(err))
			return
		}

		database, err := db.New(cmdArgs.DatabaseURL)
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
			DB:           database,
			Storage:      storage,
		}
		handler, err := site.NewHandler(
			logger.Named("site"),
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

		var tls *httputil.ServerTLSConfig
		if cmdArgs.TLS {
			if cmdArgs.TLSProtectKey == "" {
				logger.Warn("TLS protect key not specified; certificate private keys would be stored in plain text.")
			}
			tls = &httputil.ServerTLSConfig{
				Storage:       db.NewCertStorage(database, cmdArgs.TLSProtectKey),
				ACMEDirectory: cmdArgs.TLSACMEEndpoint,
				ACMEEmail:     cmdArgs.TLSACMEEmail,
				Addr:          cmdArgs.TLSAddr,
				CheckDomain:   handler.CheckValidDomain,
			}
		}

		server := httputil.Server{
			Logger:  logger.Named("server"),
			Addr:    cmdArgs.Addr,
			Handler: handler,
			TLS:     tls,
		}

		command.Run([]command.WorkFunc{
			server.Run,
		})
	},
}

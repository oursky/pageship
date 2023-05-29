package app

import (
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/cron"
	"github.com/oursky/pageship/internal/db"
	_ "github.com/oursky/pageship/internal/db/postgres"
	_ "github.com/oursky/pageship/internal/db/sqlite"
	"github.com/oursky/pageship/internal/handler/controller"
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

	startCmd.PersistentFlags().Bool("migrate", false, "migrate before starting")
	startCmd.PersistentFlags().String("addr", ":8001", "listen address")

	startCmd.PersistentFlags().String("max-deployment-size", "200M", "max deployment files size")
	startCmd.PersistentFlags().String("storage-key-prefix", "", "storage key prefix")
	startCmd.PersistentFlags().String("host-pattern", config.DefaultHostPattern, "host match pattern")
	startCmd.PersistentFlags().String("host-id-scheme", string(config.HostIDSchemeDefault), "host ID scheme")

	startCmd.PersistentFlags().String("token-authority", "pageship", "auth token authority")
	startCmd.PersistentFlags().String("token-signing-key", "", "auth token signing key")

	startCmd.PersistentFlags().String("cleanup-expired-crontab", "", "cleanup expired schedule")
	startCmd.PersistentFlags().Duration("keep-after-expired", time.Hour*24, "keep-after-expired")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start controller server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		migrate := viper.GetBool("migrate")
		if migrate {
			logger.Info("migrating database before starting...")
			err := doMigrate(viper.GetString("database-url"), false)
			if err != nil {
				logger.Fatal("failed to migrate", zap.Error(err))
				return err
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cmdArgs struct {
			DatabaseURL           string              `mapstructure:"database-url" validate:"url"`
			StorageURL            string              `mapstructure:"storage-url" validate:"url"`
			Addr                  string              `mapstructure:"addr" validate:"hostname_port"`
			MaxDeploymentSize     string              `mapstructure:"max-deployment-size" validate:"size"`
			StorageKeyPrefix      string              `mapstructure:"storage-key-prefix"`
			HostPattern           string              `mapstructure:"host-pattern"`
			HostIDScheme          config.HostIDScheme `mapstructure:"host-id-scheme" validate:"hostidscheme"`
			TokenSigningKey       string              `mapstructure:"token-signing-key"`
			TokenAuthority        string              `mapstructure:"token-authority"`
			CleanupExpiredCrontab string              `mapstructure:"cleanup-expired-crontab" validate:"omitempty,cron"`
			KeepAfterExpired      time.Duration       `mapstructure:"keep-after-expired" validate:"min=0"`
		}
		if err := viper.Unmarshal(&cmdArgs); err != nil {
			logger.Fatal("invalid config", zap.Error(err))
			return
		}
		if err := validate.Struct(cmdArgs); err != nil {
			logger.Fatal("invalid config", zap.Error(err))
			return
		}

		if !debugMode {
			gin.SetMode(gin.ReleaseMode)
		}

		maxDeploymentSize, _ := humanize.ParseBytes(cmdArgs.MaxDeploymentSize)
		tokenSigningKey := cmdArgs.TokenSigningKey
		if tokenSigningKey == "" {
			logger.Warn("token signing key not specified; a temporary generated key would be used.")
			tokenSigningKey = generateSecret()
		}

		config := controller.Config{
			MaxDeploymentSize: int64(maxDeploymentSize),
			StorageKeyPrefix:  cmdArgs.StorageKeyPrefix,
			HostIDScheme:      cmdArgs.HostIDScheme,
			HostPattern:       config.NewHostPattern(cmdArgs.HostPattern),
			TokenSigningKey:   []byte(tokenSigningKey),
			TokenAuthority:    cmdArgs.TokenAuthority,
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

		ctrl := &controller.Controller{
			Logger:  logger.Named("controller"),
			Config:  config,
			Storage: storage,
			DB:      db,
		}
		server := command.HTTPServer{
			Logger: zapLogger{Logger: logger.Named("server")},
			Server: http.Server{
				Addr:    cmdArgs.Addr,
				Handler: ctrl.Handler(),
			},
		}

		cronr := command.CronRunner{
			Logger: zap.L().Named("cron"),
			Jobs: []command.CronJob{
				&cron.CleanupExpired{
					Schedule:         cmdArgs.CleanupExpiredCrontab,
					KeepAfterExpired: cmdArgs.KeepAfterExpired,
					DB:               db,
				},
			},
		}

		command.Run([]command.WorkFunc{
			server.Run,
			cronr.Run,
		})
	},
}

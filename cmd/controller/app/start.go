package app

import (
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/cron"
	"github.com/oursky/pageship/internal/db"
	_ "github.com/oursky/pageship/internal/db/sqlite"
	"github.com/oursky/pageship/internal/handler/controller"
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

	startCmd.PersistentFlags().String("addr", ":8001", "listen address")

	startCmd.PersistentFlags().String("max-deployment-size", "200M", "max deployment files size")
	startCmd.PersistentFlags().String("storage-key-prefix", "", "storage key prefix")
	startCmd.PersistentFlags().String("host-pattern", config.DefaultHostPattern, "host match pattern")
	startCmd.PersistentFlags().String("host-id-scheme", string(config.HostIDSchemeDefault), "host ID scheme")

	startCmd.PersistentFlags().String("token-authority", "pageship", "auth token authority")
	startCmd.PersistentFlags().String("token-sign-secret", "", "auth token sign secret")
	startCmd.MarkPersistentFlagRequired("token-sign-secret")

	startCmd.PersistentFlags().String("cleanup-expired-crontab", "", "cleanup expired schedule")
	startCmd.PersistentFlags().String("keep-after-expired", "24h", "keep-after-expired")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start controller server",
	Run: func(cmd *cobra.Command, args []string) {
		var cmdArgs struct {
			Database              string              `mapstructure:"database" validate:"url"`
			StorageEndpoint       string              `mapstructure:"storage-endpoint" validate:"url"`
			Addr                  string              `mapstructure:"addr" validate:"hostname_port"`
			MaxDeploymentSize     string              `mapstructure:"max-deployment-size" validate:"size"`
			StorageKeyPrefix      string              `mapstructure:"storage-key-prefix"`
			HostPattern           string              `mapstructure:"host-pattern"`
			HostIDScheme          config.HostIDScheme `mapstructure:"host-id-scheme" validate:"hostidscheme"`
			TokenSignSecret       string              `mapstructure:"token-sign-secret"`
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

		config := controller.Config{
			MaxDeploymentSize: int64(maxDeploymentSize),
			StorageKeyPrefix:  cmdArgs.StorageKeyPrefix,
			HostIDScheme:      cmdArgs.HostIDScheme,
			HostPattern:       config.NewHostPattern(cmdArgs.HostPattern),
			TokenSignSecret:   []byte(cmdArgs.TokenSignSecret),
			TokenAuthority:    cmdArgs.TokenAuthority,
		}

		db, err := db.New(cmdArgs.Database)
		if err != nil {
			logger.Fatal("failed to setup database", zap.Error(err))
			return
		}

		storage, err := storage.New(cmd.Context(), cmdArgs.StorageEndpoint)
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
			Logger:  zapLogger{Logger: logger.Named("server")},
			Addr:    cmdArgs.Addr,
			Handler: ctrl.Handler(),
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

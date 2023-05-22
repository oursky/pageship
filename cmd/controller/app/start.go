package app

import (
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
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
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start controller server",
	Run: func(cmd *cobra.Command, args []string) {
		database := viper.GetString("database")
		storageEndpoint := viper.GetString("storage-endpoint")
		addr := viper.GetString("addr")

		maxDeploymentSize, err := humanize.ParseBytes(viper.GetString("max-deployment-size"))
		if err != nil {
			logger.Fatal("invalid max deployment size", zap.Error(err))
			return
		}
		storageKeyPrefix := viper.GetString("storage-key-prefix")
		hostPattern := viper.GetString("host-pattern")

		if !debugMode {
			gin.SetMode(gin.ReleaseMode)
		}

		config := controller.Config{
			MaxDeploymentSize: int64(maxDeploymentSize),
			StorageKeyPrefix:  storageKeyPrefix,
			HostPattern:       config.NewHostPattern(hostPattern),
		}

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

		ctrl := &controller.Controller{
			Config:  config,
			Storage: storage,
			DB:      db,
		}
		server := command.HTTPServer{Addr: addr, Handler: ctrl.Handler()}

		command.Run([]command.WorkFunc{
			server.Run,
		})
	},
}

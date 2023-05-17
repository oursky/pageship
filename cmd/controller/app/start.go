package app

import (
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/db"
	_ "github.com/oursky/pageship/internal/db/sqlite"
	"github.com/oursky/pageship/internal/handler/controller"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().String("addr", ":8001", "listen address")
	viper.BindPFlags(startCmd.PersistentFlags())
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start controller server",
	Run: func(cmd *cobra.Command, args []string) {
		database := viper.GetString("database")
		addr := viper.GetString("addr")
		if database == "" {
			logger.Fatal("missing database URL")
			return
		}

		if !debugMode {
			gin.SetMode(gin.ReleaseMode)
		}

		db, err := db.New(database)
		if err != nil {
			logger.Fatal("failed to setup database", zap.Error(err))
			return
		}

		ctrl := &controller.Controller{
			DB: db,
		}
		server := command.HTTPServer{Addr: addr, Handler: ctrl.Handler()}

		command.Run(logger, []command.WorkFunc{
			server.Run,
		})
	},
}

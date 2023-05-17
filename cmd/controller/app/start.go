package app

import (
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/handler/controller"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		addr := viper.GetString("addr")
		if !debugMode {
			gin.SetMode(gin.ReleaseMode)
		}

		ctrl := &controller.Controller{}
		server := command.HTTPServer{Addr: addr, Handler: ctrl.Handler()}

		command.Run(logger, []command.WorkFunc{
			server.Run,
		})
	},
}

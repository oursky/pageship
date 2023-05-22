package app

import (
	"github.com/oursky/pageship/internal/api"
	"github.com/oursky/pageship/internal/command"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var apiClient *api.Client
var debugMode bool

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "debug mode")
	rootCmd.PersistentFlags().String("api", "localhost:8001", "server API endpoing")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	command.BindConfig(rootCmd)

	debugMode = viper.GetBool("debug")
	apiEndpoint := viper.GetString("api")
	apiClient = api.NewClient(apiEndpoint)
}

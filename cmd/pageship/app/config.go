package app

import (
	"strings"

	"github.com/oursky/pageship/internal/api"
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
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.SetEnvPrefix("PAGESHIP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	debugMode = viper.GetBool("debug")
	apiEndpoint := viper.GetString("api")
	apiClient = api.NewClient(apiEndpoint)
}

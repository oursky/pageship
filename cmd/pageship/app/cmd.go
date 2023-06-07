package app

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/oursky/pageship/internal/api"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var apiEndpoint string
var apiClient *api.Client
var debugMode bool

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "debug mode")
	rootCmd.PersistentFlags().String("api", "", "server API endpoint")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	command.BindConfig(rootCmd)

	debugMode = viper.GetBool("debug")
	apiEndpoint = viper.GetString("api")
}

func API() *api.Client {
	if apiClient == nil {
		if apiEndpoint == "" {
			conf, err := config.LoadClientConfig()
			if err != nil {
				panic(fmt.Errorf("failed to load config: %w", err))
			}
			apiEndpoint = conf.APIServer
		}

		if apiEndpoint == "" {
			api, err := Prompt("API server", func(value string) error {
				_, err := url.Parse(value)
				return err
			})
			if err != nil {
				panic(err)
			}
			apiEndpoint = api

			conf, err := config.LoadClientConfig()
			if err != nil {
				panic(fmt.Errorf("failed to load config: %w", err))
			}

			conf.APIServer = apiEndpoint
			if err := conf.Save(); err != nil {
				panic(fmt.Errorf("failed to save config: %w", err))
			}
		}

		apiClient = api.NewClient(apiEndpoint)
		apiClient.TokenFunc = func(r *http.Request) (string, error) {
			return ensureAuth(r.Context())
		}
	}
	return apiClient
}

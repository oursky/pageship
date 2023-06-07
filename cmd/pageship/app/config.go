package app

import (
	"fmt"
	"os"

	"github.com/oursky/pageship/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configResetCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show client config",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ClientConfigPath()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		conf, err := config.LoadClientConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		Info("Client config file: %s", path)

		server := conf.APIServer
		if server == "" {
			server = "<unset>"
		}
		Info("API Server: %s", server)

		if conf.GitHubUsername != "" {
			Info("GitHub username: %s", conf.GitHubUsername)
		}
		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset client config",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := Confirm("Reset client config"); err != nil {
			return err
		}

		path, err := config.ClientConfigPath()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		err = os.Remove(path)
		if os.IsNotExist(err) {
			err = nil
		}
		if err != nil {
			return fmt.Errorf("failed to reset config: %w", err)
		}

		Info("Client config reset.")
		return nil
	},
}

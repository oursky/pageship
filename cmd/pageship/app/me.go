package app

import (
	"fmt"

	"github.com/oursky/pageship/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(meCmd)
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Get user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.LoadClientConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if conf.AuthToken == "" {
			Info("Logged out.")
			return nil
		}

		me, err := apiClient.GetMe(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get user info: %w", err)
		}

		Info("Logged in as %q. (id: %q)", me.Name, me.ID)
		Info("Credentials:")
		for _, id := range me.Credentials {
			Info(" - %q", id)
		}
		return nil
	},
}

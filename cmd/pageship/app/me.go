package app

import (
	"github.com/oursky/pageship/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(meCmd)
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Get user info",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.LoadClientConfig()
		if err != nil {
			Error("Failed to load config: %s", err)
			return
		}

		if conf.AuthToken == "" {
			Info("Logged out.")
			return
		}

		me, err := apiClient.GetMe(cmd.Context())
		if err != nil {
			Error("Failed to login via SSH: %s", err)
			return
		}

		Info("Logged in as %q. (id: %q)", me.Name, me.ID)
		Info("Credentials:")
		for _, id := range me.Credentials {
			Info(" - %q", id)
		}
	},
}

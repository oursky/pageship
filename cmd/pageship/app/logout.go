package app

import (
	"fmt"

	"github.com/oursky/pageship/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout user",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.LoadClientConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		conf.AuthToken = ""
		err = conf.Save()
		if err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		Info("Logged out.")
		return nil
	},
}

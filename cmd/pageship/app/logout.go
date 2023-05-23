package app

import (
	"github.com/oursky/pageship/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout user",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.LoadClientConfig()
		if err != nil {
			Error("Failed to load config: %s", err)
			return
		}
		conf.AuthToken = ""
		err = conf.Save()
		if err != nil {
			Error("Failed to save config")
			return
		}

		Info("Logged out.")
	},
}

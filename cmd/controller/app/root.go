package app

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().String("database", "", "database URL")
}

var rootCmd = &cobra.Command{
	Use:   "pageship",
	Short: "Pageship controller",
}

func Execute() error {
	return rootCmd.Execute()
}

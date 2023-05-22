package app

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().String("database", "", "database URL")
	rootCmd.PersistentFlags().String("storage-endpoint", "", "object storage endpoint")
}

var rootCmd = &cobra.Command{
	Use:   "controller",
	Short: "Pageship controller",
}

func Execute() error {
	return rootCmd.Execute()
}

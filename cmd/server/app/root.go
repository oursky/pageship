package app

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Pageship server",
}

func Execute() error {
	return rootCmd.Execute()
}

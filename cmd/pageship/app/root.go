package app

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pageship",
	Short: "Pageship",
}

func Execute() error {
	return rootCmd.Execute()
}

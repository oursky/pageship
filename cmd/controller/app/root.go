package app

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pageship",
	Short: "Pageship controller",
}

func Execute() error {
	return rootCmd.Execute()
}

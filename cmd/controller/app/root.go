package app

import (
	"github.com/carlmjohnson/versioninfo"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "controller",
	Short:   "Pageship controller",
	Version: versioninfo.Short(),
}

func Execute() error {
	return rootCmd.Execute()
}

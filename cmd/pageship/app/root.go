package app

import (
	"github.com/carlmjohnson/versioninfo"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "pageship",
	Short:         "Pageship",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       versioninfo.Short(),
}

func Execute() error {
	return rootCmd.Execute()
}

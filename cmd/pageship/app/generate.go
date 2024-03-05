package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(generateDockerfileCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate [command]",
	Short: "Generate files",
	Args:  cobra.NoArgs, //if unknown command, will return error just like main pageship command
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help() //show help if no subcommand supplied
		return nil
	},
}

var generateDockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate dockerfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("buh")
		return nil
	},
}

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
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("guh")
		return nil
	},
}

var generateDockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate dockerfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("guh")
		return nil
	},
}

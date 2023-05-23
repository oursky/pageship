package app

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appsCmd)
}

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage apps",
	Run: func(cmd *cobra.Command, args []string) {
		apps, err := apiClient.ListApps(cmd.Context())
		if err != nil {
			Error("Failed to list apps: %s", err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 4, 4, ' ', 0)
		fmt.Fprintln(w, "ID\tENVIRONMENTS\tURL")
		for _, app := range apps {
			var envs []string
			for _, e := range app.Config.Environments {
				envs = append(envs, e.Name)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", app.ID, strings.Join(envs, ","), app.URL)
		}
		w.Flush()
	},
}

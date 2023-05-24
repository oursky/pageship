package app

import (
	"fmt"
	"os"
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
		fmt.Fprintln(w, "ID\tURL")
		for _, app := range apps {
			fmt.Fprintf(w, "%s\t%s\n", app.ID, app.URL)
		}
		w.Flush()
	},
}

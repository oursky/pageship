package app

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(sitesCmd)
	sitesCmd.PersistentFlags().String("app", "", "app ID")
}

var sitesCmd = &cobra.Command{
	Use:   "sites",
	Short: "Manage sites",
	RunE: func(cmd *cobra.Command, args []string) error {
		appID := viper.GetString("app")
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			return fmt.Errorf("app ID is not set")
		}

		sites, err := API().ListSites(cmd.Context(), appID)
		if err != nil {
			return fmt.Errorf("failed to list sites: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 4, 4, ' ', 0)
		fmt.Fprintln(w, "NAME\tURL\tDEPLOYMENT")
		for _, site := range sites {
			deployment := "-"
			if site.DeploymentName != nil {
				deployment = *site.DeploymentName
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", site.Name, site.URL, deployment)
		}
		w.Flush()
		return nil
	},
}

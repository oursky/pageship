package app

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getAppsCmd)
	getCmd.AddCommand(getSitesCmd)
	getCmd.AddCommand(getDeploymentsCmd)

	getCmd.PersistentFlags().String("app", "", "app ID")
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info",
}

var getAppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "List apps",
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

var getSitesCmd = &cobra.Command{
	Use:   "sites",
	Short: "List sites",
	Run: func(cmd *cobra.Command, args []string) {
		appID := viper.GetString("app")

		sites, err := apiClient.ListSites(cmd.Context(), appID)
		if err != nil {
			Error("Failed to list sites: %s", err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 4, 4, ' ', 0)
		fmt.Fprintln(w, "NAME\tURL\tLAST DEPLOYED AT")
		for _, site := range sites {
			lastDeployedAt := "-"
			if site.LastDeployAt != nil {
				lastDeployedAt = site.LastDeployAt.Local().Format(time.DateTime)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", site.Name, site.URL, lastDeployedAt)
		}
		w.Flush()
	},
}

var getDeploymentsCmd = &cobra.Command{
	Use:   "deployments site",
	Short: "List deployments",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appID := viper.GetString("app")
		siteName := args[0]

		deployments, err := apiClient.ListDeployments(cmd.Context(), appID, siteName)
		if err != nil {
			Error("Failed to list deployments: %s", err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 4, 4, ' ', 0)
		fmt.Fprintln(w, "ID\tCREATED AT\tSTATUS")
		for _, deployment := range deployments {
			createdAt := deployment.CreatedAt.Local().Format(time.DateTime)
			fmt.Fprintf(w, "%s\t%s\t%s\n", deployment.ID, createdAt, deployment.Status)
		}
		w.Flush()
	},
}

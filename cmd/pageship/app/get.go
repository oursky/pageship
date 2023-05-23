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

	getSitesCmd.PersistentFlags().String("app", "", "app ID")
	getDeploymentsCmd.PersistentFlags().String("app", "", "app ID")
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
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			Error("App ID is not set")
			return
		}

		sites, err := apiClient.ListSites(cmd.Context(), appID)
		if err != nil {
			Error("Failed to list sites: %s", err)
			return
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
	},
}

var getDeploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "List deployments",
	Run: func(cmd *cobra.Command, args []string) {
		appID := viper.GetString("app")
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			Error("App ID is not set")
			return
		}

		sites, err := apiClient.ListSites(cmd.Context(), appID)
		if err != nil {
			Error("Failed to list sites: %s", err)
			return
		}

		deployments, err := apiClient.ListDeployments(cmd.Context(), appID)
		if err != nil {
			Error("Failed to list deployments: %s", err)
			return
		}

		deploymentSites := map[string][]string{}
		for _, site := range sites {
			if site.DeploymentName == nil {
				continue
			}

			n := deploymentSites[*site.DeploymentName]
			n = append(n, site.Name)
			deploymentSites[*site.DeploymentName] = n
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 4, 4, ' ', 0)
		fmt.Fprintln(w, "NAME\tCREATED AT\tSITES")
		for _, deployment := range deployments {
			createdAt := deployment.CreatedAt.Local().Format(time.DateTime)
			sites := strings.Join(deploymentSites[deployment.Name], ",")
			fmt.Fprintf(w, "%s\t%s\t%s\n", deployment.Name, createdAt, sites)
		}
		w.Flush()
	},
}

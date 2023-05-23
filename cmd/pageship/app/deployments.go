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
	rootCmd.AddCommand(deploymentsCmd)
	deploymentsCmd.PersistentFlags().String("app", "", "app ID")
}

var deploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "Manage deployments",
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

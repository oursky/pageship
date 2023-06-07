package app

import (
	"fmt"
	"os"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		appID := viper.GetString("app")
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			return fmt.Errorf("app ID is not set")
		}

		deployments, err := API().ListDeployments(cmd.Context(), appID)
		if err != nil {
			return fmt.Errorf("failed to list deployments: %w", err)
		}

		now := time.Now()
		w := tabwriter.NewWriter(os.Stdout, 1, 4, 4, ' ', 0)
		fmt.Fprintln(w, "NAME\tCREATED AT\tSTATUS\tURL")
		for _, deployment := range deployments {
			createdAt := deployment.CreatedAt.Local().Format(time.DateTime)

			var status string
			switch {
			case deployment.IsExpired(now):
				status = "EXPIRED"
			case deployment.UploadedAt == nil:
				status = "PENDING"
			case deployment.SiteName != nil:
				status = "ACTIVE"
			default:
				status = "INACTIVE"
			}

			url := ""
			if deployment.URL != nil {
				url = *deployment.URL
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", deployment.Name, createdAt, status, url)
		}
		w.Flush()
		return nil
	},
}

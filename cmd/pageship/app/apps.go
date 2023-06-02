package app

import (
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/oursky/pageship/internal/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsCreateCmd)
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

var appsCreateCmd = &cobra.Command{
	Use:   "create [app-id]",
	Short: "Create app",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appID := ""
		if len(args) > 0 {
			appID = args[0]
		}
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			Error("App ID is not set")
			return
		}

		app, err := apiClient.GetApp(cmd.Context(), appID)
		if code, ok := api.ErrorStatusCode(err); ok && code == http.StatusNotFound {
			app = nil
		} else if err != nil {
			Error("Failed to get app: %s", err)
			return
		}

		if app != nil {
			Info("App %q is already created.", appID)
			return
		}

		app, err = apiClient.CreateApp(cmd.Context(), appID)
		if err != nil {
			Error("Failed to create app: %s", err)
			return
		}

		Debug("App: %+v", app)
		Info("App %q is created.", appID)
	},
}

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/oursky/pageship/internal/api"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsCreateCmd)
	appsCmd.AddCommand(appsShowCmd)
	appsCmd.AddCommand(appsConfigureCmd)
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

var appsShowCmd = &cobra.Command{
	Use:   "show [app-id]",
	Short: "Show app config",
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
		if err != nil {
			Error("Failed to get app: %s", err)
			return
		}

		jsonConf, err := json.Marshal(app.Config)
		if err != nil {
			Error("Failed to serialize config: %s", err)
			return
		}

		var rawConf map[string]any
		if err := json.Unmarshal(jsonConf, &rawConf); err != nil {
			Error("Failed to serialize config: %s", err)
			return
		}

		conf, err := toml.Marshal(map[string]any{"app": rawConf})
		if err != nil {
			Error("Failed to serialize config: %s", err)
			return
		}

		fmt.Println(string(conf))
	},
}

var appsConfigureCmd = &cobra.Command{
	Use:   "configure [deploy directory]",
	Short: "Configure app",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		conf, err := loadConfig(dir)
		if err != nil {
			Error("Failed to load config: %s", err)
			return
		}

		Info("Configuring app %q...", conf.App.ID)

		app, err := apiClient.GetApp(cmd.Context(), conf.App.ID)
		if err != nil {
			Error("Failed to get app: %s", err)
			return
		}

		app, err = apiClient.ConfigureApp(cmd.Context(), app.ID, &conf.App)
		if err != nil {
			Error("Failed to configure app: %s", err)
			return
		}

		Info("Configured app %q.", app.ID)
	},
}

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
	RunE: func(cmd *cobra.Command, args []string) error {
		apps, err := API().ListApps(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list apps: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 4, 4, ' ', 0)
		fmt.Fprintln(w, "ID\tURL")
		for _, app := range apps {
			fmt.Fprintf(w, "%s\t%s\n", app.ID, app.URL)
		}
		w.Flush()
		return nil
	},
}

var appsCreateCmd = &cobra.Command{
	Use:   "create [app-id]",
	Short: "Create app",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appID := ""
		if len(args) > 0 {
			appID = args[0]
		}
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			return fmt.Errorf("app ID is not set")
		}

		app, err := API().GetApp(cmd.Context(), appID)
		if code, ok := api.ErrorStatusCode(err); ok && code == http.StatusNotFound {
			app = nil
		} else if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}

		if app != nil {
			Info("App %q is already created.", appID)
			return nil
		}

		app, err = API().CreateApp(cmd.Context(), appID)
		if err != nil {
			return fmt.Errorf("failed to create app: %w", err)
		}

		Debug("App: %+v", app)
		Info("App %q is created.", appID)
		return nil
	},
}

var appsShowCmd = &cobra.Command{
	Use:   "show [app-id]",
	Short: "Show app config",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appID := ""
		if len(args) > 0 {
			appID = args[0]
		}
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			return fmt.Errorf("app ID is not set")
		}

		app, err := API().GetApp(cmd.Context(), appID)
		if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}

		jsonConf, err := json.Marshal(app.Config)
		if err != nil {
			return fmt.Errorf("failed to serialize config: %w", err)
		}

		var rawConf map[string]any
		if err := json.Unmarshal(jsonConf, &rawConf); err != nil {
			return fmt.Errorf("failed to serialize config: %w", err)
		}

		conf, err := toml.Marshal(map[string]any{"app": rawConf})
		if err != nil {
			return fmt.Errorf("failed to serialize config: %w", err)
		}

		fmt.Println(string(conf))
		return nil
	},
}

var appsConfigureCmd = &cobra.Command{
	Use:   "configure [deploy directory]",
	Short: "Configure app",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		conf, err := loadConfig(dir)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		Info("Configuring app %q...", conf.App.ID)

		app, err := API().GetApp(cmd.Context(), conf.App.ID)
		if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}

		oldConfig := app.Config

		app, err = API().ConfigureApp(cmd.Context(), app.ID, &conf.App)
		if err != nil {
			return fmt.Errorf("failed to configure app: %w", err)
		}

		for _, dconf := range conf.App.Domains {
			if _, exists := oldConfig.ResolveDomain(dconf.Domain); !exists {
				Info("Activating custom domain %q...", dconf.Domain)
				_, err = API().ActivateDomain(cmd.Context(), app.ID, dconf.Domain, "")
				if err != nil {
					Warn("Activation of custom domain %q failed: %s", dconf.Domain, err)
				}
			}
		}

		Info("Configured app %q.", app.ID)
		return nil
	},
}

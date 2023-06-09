package app

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/oursky/pageship/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configTemplate = template.Must(template.New("").Parse(`
[app]
id = {{.appID}}

team = []

[app.deployments]
# ttl = "24h"
# access = []

[[app.sites]]
name = "main"

# [[app.sites]]
# name = "dev"

# [[app.sites]]
# name = "staging"

[site]
public = {{.public}}

# access = []
`))

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.PersistentFlags().String("dir", ".", "target directory")
	initCmd.PersistentFlags().String("app", "", "app ID")
	initCmd.PersistentFlags().String("public", "", "static files directory")
	initCmd.PersistentFlags().Bool("register", false, "static files directory")
}

var initCmd = &cobra.Command{
	Use:   "init [--dir target directory] [--app app ID] [--public static files directory] [--register]",
	Short: "Initialize new app",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := viper.GetString("dir")
		dir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("invalid target directory: %w", err)
		}

		appID := viper.GetString("app")
		if !config.ValidateDNSLabel(appID) {
			prompt := promptui.Prompt{
				Label: "App ID",
				Validate: func(s string) error {
					if !config.ValidateDNSLabel(s) {
						return errors.New("must be a valid DNS label")
					}
					return nil
				},
			}
			result, err := prompt.Run()
			if err != nil {
				Info("Cancelled.")
				return ErrCancelled
			}
			appID = result
		}

		public := viper.GetString("public")
		if !fs.ValidPath(public) {
			prompt := promptui.Prompt{
				Label:   "Static files directory",
				Default: "dist",
				Validate: func(s string) error {
					if !fs.ValidPath(s) {
						return errors.New("must be a valid path")
					}
					return nil
				},
			}
			result, err := prompt.Run()
			if err != nil {
				Info("Cancelled.")
				return ErrCancelled
			}
			public = result
		}

		publicDir := filepath.Join(dir, public)
		if _, err := os.Stat(publicDir); os.IsNotExist(err) {
			prompt := promptui.Prompt{
				Label:     "Static files directory not found; continue",
				IsConfirm: true,
			}
			_, err := prompt.Run()
			if err != nil {
				Info("Cancelled.")
				return ErrCancelled
			}
		}

		var data bytes.Buffer
		configTemplate.Execute(&data, map[string]any{
			"appID":  strconv.Quote(appID),
			"public": strconv.Quote(public),
		})
		conf := strings.TrimSpace(data.String()) + "\n"

		err = os.WriteFile(filepath.Join(dir, config.SiteConfigName+".toml"), []byte(conf), 0644)
		if err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		Info("Done!")
		Info("Run `pageship deploy` to deploy your app now!")
		return nil
	},
}

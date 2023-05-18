package app

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/deploy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.PersistentFlags().String("site", "", "target site")
	deployCmd.PersistentFlags().BoolP("yes", "y", false, "skip confirmation")
	viper.BindPFlags(deployCmd.PersistentFlags())
}

var deployCmd = &cobra.Command{
	Use:   "deploy directory",
	Short: "Deploy site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		site := viper.GetString("site")
		yes := viper.GetBool("yes")
		fsys := os.DirFS(args[0])

		loader := config.NewLoader(config.SiteConfigName)

		conf := config.DefaultConfig()
		if err := loader.Load(fsys, &conf); err != nil {
			logger.Fatal("failed to load config", zap.Error(err))
			return
		}
		conf.SetDefaults()

		appID := conf.ID
		if site == "" {
			site = conf.DefaultEnvironment
		}

		if !config.ValidateDNSLabel(site) {
			logger.Fatal("invalid site name; site name must be a valid DNS label",
				zap.String("site", site))
			return
		}

		env, ok := conf.ResolveSite(site)
		if !ok {
			logger.Fatal("site is not defined by any environment",
				zap.String("site", site))
		}

		if !yes {
			var label string
			if site == env.Name {
				label = fmt.Sprintf("Deploy to site '%s' of app '%s'?", site, appID)
			} else {
				label = fmt.Sprintf("Deploy to site '%s' (%s) of app '%s'?", site, env.Name, appID)
			}

			prompt := promptui.Prompt{Label: label, IsConfirm: true}
			_, err := prompt.Run()
			if err != nil {
				logger.Info("cancelled", zap.Error(err))
				return
			}
		}

		logger.Info("collecting files")
		entries, err := deploy.CollectFileList(fsys)
		if err != nil {
			logger.Fatal("failed to collect files", zap.Error(err))
			return
		}
		logger.Info("setting up deployment", zap.Int("files", len(entries)))
	},
}

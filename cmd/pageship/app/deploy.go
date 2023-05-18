package app

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/oursky/pageship/internal/config"
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
		dir := args[0]

		loader := config.NewLoader(config.SiteConfigName)

		conf := config.DefaultConfig()
		if err := loader.Load(os.DirFS(dir), &conf); err != nil {
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
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Deploy to site '%s' of app '%s'?", site, appID),
				IsConfirm: true,
			}
			_, err := prompt.Run()
			if err != nil {
				logger.Info("cancelled", zap.Error(err))
				return
			}
		}

		logger.Sugar().Debug("TODO", site, env, args[0])
	},
}

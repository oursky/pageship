package app

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	DevMode bool   `mapstructure:"dev"`
	Prefix  string `mapstructure:"prefix"`
}

var cmdConfig Config
var logger *zap.Logger

func init() {
	rootCmd.PersistentFlags().Bool("dev", false, "development mode")
	rootCmd.PersistentFlags().String("prefix", "", "config directory prefix")

	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLogger)
}

func initConfig() {
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.SetEnvPrefix("PAGESHIP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&cmdConfig); err != nil {
		log.Fatalf("failed to initialize config: %s", err)
	}

	if cmdConfig.Prefix == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to get current directory: %s", err)
		}
		cmdConfig.Prefix = cwd
	}
}

func initLogger() {
	var cfg zap.Config
	if cmdConfig.DevMode {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	logger, _ = cfg.Build()
	cobra.OnFinalize(func() { logger.Sync() })
}

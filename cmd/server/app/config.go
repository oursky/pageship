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

var config Config
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

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("failed to initialize config: %s", err)
	}

	if config.Prefix == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to get current directory: %s", err)
		}
		config.Prefix = cwd
	}
}

func initLogger() {
	var cfg zap.Config
	if config.DevMode {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	logger, _ = cfg.Build()
	cobra.OnFinalize(func() { logger.Sync() })
}

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
	DebugMode bool   `mapstructure:"debug"`
	Prefix    string `mapstructure:"prefix"`
	Addr      string `mapstructure:"addr"`
}

var cmdConfig Config
var logger *zap.Logger

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "debug mode")
	rootCmd.PersistentFlags().String("prefix", "", "base directory")
	rootCmd.PersistentFlags().String("addr", ":8001", "listen address")

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
	if cmdConfig.DebugMode {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	logger, _ = cfg.Build()
	cobra.OnFinalize(func() { logger.Sync() })
}

package app

import (
	"github.com/dustin/go-humanize"
	"github.com/go-playground/validator/v10"
	"github.com/oursky/pageship/internal/command"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var validate = validator.New()
var debugMode bool

var logger *zap.Logger

func init() {
	validate.RegisterValidation("size", func(fl validator.FieldLevel) bool {
		_, err := humanize.ParseBytes(fl.Field().String())
		return err == nil
	})

	rootCmd.PersistentFlags().Bool("debug", false, "debug mode")

	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLogger)
}

func initConfig() {
	command.BindConfig(rootCmd)
	debugMode = viper.GetBool("debug")
}

func initLogger() {
	var cfg zap.Config
	if debugMode {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	logger, _ = cfg.Build()
	cobra.OnFinalize(func() { logger.Sync() })
	zap.ReplaceGlobals(logger)
}

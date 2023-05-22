package command

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func BindConfig(root *cobra.Command) {
	viper.SetEnvPrefix("PAGESHIP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	setCommandFlags(root)
}

func setCommandFlags(cmd *cobra.Command) {
	viper.BindPFlags(cmd.PersistentFlags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) {
			cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})

	for _, child := range cmd.Commands() {
		setCommandFlags(child)
	}
}

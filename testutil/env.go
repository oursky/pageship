package testutil

import (
	"strings"

	"github.com/spf13/viper"
)

func LoadTestEnvs() {
	viper.SetEnvPrefix("TEST_PAGESHIP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
}

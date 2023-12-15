package testutil

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
)

func applyDefaultEnvs() {
	dbUrl := viper.GetString("database-url")
	if dbUrl == "" {
		dbDir := path.Join(os.TempDir(), "data.local")
		err := os.Mkdir(dbDir, 0755)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}
		dbUrl = fmt.Sprintf("sqlite://%s/data.db", dbDir)
		viper.Set("database-url", dbUrl)
	}
	storageUrl := viper.GetString("storage-url")
	if storageUrl == "" {
		dir := path.Join(os.TempDir(), "data.local", "storage")
		err := os.MkdirAll(dir, 0755)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}
		storageUrl = fmt.Sprintf("file://%s", dir)
		viper.Set("storage-url", storageUrl)
	}
}

func LoadTestEnvs() {
	viper.SetEnvPrefix("TEST_PAGESHIP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	applyDefaultEnvs()
}

package config

import (
	"errors"
	"io/fs"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var ErrConfigNotFound = errors.New("config not found")

type absFS struct {
	fs.FS
}

func (fs absFS) Open(name string) (fs.File, error) {
	if path.IsAbs(name) {
		// Translate abs path to raw path name to workaround viper AddConfigPath absify the path.
		name = name[1:]
	}
	return fs.FS.Open(name)
}

type Loader struct {
	viper *viper.Viper
}

func NewLoader(name string) *Loader {
	loader := &Loader{viper: viper.NewWithOptions(viper.KeyDelimiter("/"))}

	loader.viper.SetConfigName(name)
	loader.viper.AddConfigPath("/")

	return loader
}

func (l *Loader) Load(fsys fs.FS, conf any) error {
	l.viper.SetFs(afero.FromIOFS{FS: absFS{FS: fsys}})
	defer l.viper.SetFs(nil)

	err := l.viper.ReadInConfig()
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		return ErrConfigNotFound
	} else if err != nil {
		return err
	}

	if err := l.viper.Unmarshal(conf); err != nil {
		return err
	}

	if err := validate.Struct(conf); err != nil {
		return err
	}

	return nil
}

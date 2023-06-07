package app

import (
	"errors"

	"github.com/carlmjohnson/versioninfo"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "pageship",
	Short:         "Pageship",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       versioninfo.Short(),
}

func Execute() (err error) {
	defer func() {
		if errors.Is(err, ErrCancelled) {
			err = nil
		}
	}()
	defer func() {
		if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
			} else {
				panic(e)
			}
		}
	}()
	return rootCmd.Execute()
}

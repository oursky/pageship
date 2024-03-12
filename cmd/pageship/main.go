package main

import (
	"os"

	"github.com/oursky/pageship/cmd/pageship/app"
)

var version = "dev" //https://goreleaser.com/cookbooks/using-main.version/

func main() {
	app.Version = version
	if err := app.Execute(); err != nil {
		app.Error("%s", err)
		os.Exit(1)
	}
}

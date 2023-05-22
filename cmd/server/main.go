package main

import (
	"os"

	"github.com/oursky/pageship/cmd/server/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}

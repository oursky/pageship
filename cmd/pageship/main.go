package main

import (
	"os"

	"github.com/oursky/pageship/cmd/pageship/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}

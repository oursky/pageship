package main

import (
	"os"

	"github.com/oursky/pageship/cmd/controller/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}

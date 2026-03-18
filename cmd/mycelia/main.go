package main

import (
	"os"

	"github.com/sivepanda/mycelia/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

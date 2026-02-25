package main

import (
	"os"

	"github.com/sivepanda/mycelia/internal/cli"
)

func main() {
	if err := cli.NewMycCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

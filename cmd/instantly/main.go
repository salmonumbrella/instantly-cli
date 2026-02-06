package main

import (
	"os"

	"github.com/salmonumbrella/instantly-cli/internal/cmd"
)

var exit = os.Exit

func run() int {
	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}

func main() {
	exit(run())
}

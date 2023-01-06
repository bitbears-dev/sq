package main

import (
	"os"

	"github.com/bitbears-dev/sq/cli"
)

const version = "0.1.0"

func main() {
	os.Exit(cli.NewCLI(version).Run(os.Args[1:]))
}

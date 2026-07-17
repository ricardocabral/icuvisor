// Package main is the entrypoint for the standalone icuvisor CLI.
package main

import (
	"context"
	"os"
	_ "time/tzdata"

	"github.com/ricardocabral/icuvisor/internal/cli"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(cli.RunCLI(context.Background(), cli.Options{
		Version: version,
		Args:    os.Args[1:],
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}))
}

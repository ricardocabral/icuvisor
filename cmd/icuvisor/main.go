// Package main is the entrypoint for the icuvisor MCP server binary.
package main

import (
	"context"
	"os"

	"github.com/ricardocabral/icuvisor/internal/app"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(app.RunCLI(context.Background(), app.Options{
		Version: version,
		Args:    os.Args[1:],
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}))
}

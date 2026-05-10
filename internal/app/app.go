package app

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// ErrServerNotImplemented is returned until stdio MCP wiring lands.
var ErrServerNotImplemented = errors.New("stdio server not implemented yet")

// Options contains process-level dependencies for the icuvisor CLI.
type Options struct {
	Version string
	Args    []string
	Stdout  io.Writer
	Stderr  io.Writer

	StartServer func(context.Context, ServerInfo) error
}

// ServerInfo carries process metadata needed by lower layers.
type ServerInfo struct {
	Version string
}

// Run executes the icuvisor CLI.
func Run(ctx context.Context, opts Options) error {
	version := opts.Version
	if version == "" {
		version = "dev"
	}

	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}

	args := opts.Args
	if len(args) == 0 {
		return startServer(ctx, opts.StartServer, ServerInfo{Version: version})
	}

	switch args[0] {
	case "version":
		_, err := fmt.Fprintln(stdout, version)
		if err != nil {
			return fmt.Errorf("writing version: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q (try: icuvisor version)", args[0])
	}
}

func startServer(ctx context.Context, starter func(context.Context, ServerInfo) error, info ServerInfo) error {
	if starter == nil {
		starter = defaultStartServer
	}
	if err := starter(ctx, info); err != nil {
		return err
	}
	return nil
}

func defaultStartServer(ctx context.Context, _ ServerInfo) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrServerNotImplemented
}

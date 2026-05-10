package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
)

// ErrServerNotImplemented is returned until stdio MCP wiring lands.
var ErrServerNotImplemented = errors.New("stdio server not implemented yet")

// Options contains process-level dependencies for the icuvisor CLI.
type Options struct {
	Version string
	Args    []string
	Stdout  io.Writer
	Stderr  io.Writer

	LoadConfig  func(context.Context, config.Options) (config.Config, error)
	StartServer func(context.Context, ServerInfo) error
}

// ServerInfo carries process metadata needed by lower layers.
type ServerInfo struct {
	Version string
	Config  config.Config
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
	if len(args) > 0 && args[0] == "version" {
		_, err := fmt.Fprintln(stdout, version)
		if err != nil {
			return fmt.Errorf("writing version: %w", err)
		}
		return nil
	}

	configPath, err := parseDefaultArgs(args)
	if err != nil {
		return err
	}

	return startServer(ctx, opts.LoadConfig, opts.StartServer, ServerInfo{Version: version}, configPath)
}

func parseDefaultArgs(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	if args[0] == "--config" {
		if len(args) != 2 || strings.TrimSpace(args[1]) == "" {
			return "", errors.New("missing config path; use --config /path/to/icuvisor.json")
		}
		return args[1], nil
	}
	if path, ok := strings.CutPrefix(args[0], "--config="); ok {
		if len(args) != 1 || strings.TrimSpace(path) == "" {
			return "", errors.New("missing config path; use --config /path/to/icuvisor.json")
		}
		return path, nil
	}
	return "", fmt.Errorf("unknown command %q (try: icuvisor version)", args[0])
}

func startServer(ctx context.Context, loader func(context.Context, config.Options) (config.Config, error), starter func(context.Context, ServerInfo) error, info ServerInfo, configPath string) error {
	if loader == nil {
		loader = config.Load
	}
	cfg, err := loader(ctx, config.Options{Path: configPath})
	if err != nil {
		return err
	}
	info.Config = cfg

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

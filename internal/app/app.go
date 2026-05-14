package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	mcpserver "github.com/ricardocabral/icuvisor/internal/mcp"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

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
	Version       string
	Config        config.Config
	DebugMetadata bool
	DeleteMode    safety.Mode
	Toolset       safety.Toolset
	Capability    safety.Capability
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

	return startServer(ctx, opts.LoadConfig, opts.StartServer, ServerInfo{Version: version, DebugMetadata: response.DebugMetadataFromEnv()}, configPath)
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
	info.DeleteMode = cfg.DeleteMode
	info.Toolset = cfg.Toolset
	info.Capability = safety.NewCapability(cfg.DeleteMode)

	if starter == nil {
		starter = defaultStartServer
	}
	if err := starter(ctx, info); err != nil {
		return err
	}
	return nil
}

func defaultStartServer(ctx context.Context, info ServerInfo) error {
	logger := slog.Default()
	version := strings.TrimSpace(info.Version)
	if version == "" {
		version = "dev"
	}
	info.Version = version
	logger.Info("server starting", "version", version)

	capability := info.Capability
	if capability == nil {
		capability = safety.NewCapability(info.DeleteMode)
	}
	deleteMode := safety.ParseMode(capability.Mode())
	toolset := safety.ParseToolset(info.Toolset.String())
	response.SetDeleteMode(deleteMode.String())
	response.SetToolset(toolset.String())
	safety.LogResolvedMode(logger, deleteMode)
	safety.LogResolvedToolset(logger, toolset)
	client, err := intervals.NewClient(intervals.Options{Config: info.Config, Version: info.Version})
	if err != nil {
		return err
	}
	server, err := mcpserver.NewServer(ctx, mcpserver.Options{
		Config:     info.Config,
		Version:    info.Version,
		Logger:     logger,
		Capability: capability,
		Toolset:    toolset,
		Registry: tools.NewRegistryWithOptions(client, tools.RegistryOptions{
			Version:          info.Version,
			TimezoneFallback: info.Config.Timezone,
			DebugMetadata:    info.DebugMetadata,
			Capability:       capability,
			Toolset:          toolset,
		}),
	})
	if err != nil {
		return err
	}
	return server.Run(ctx)
}

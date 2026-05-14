package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	mcpserver "github.com/ricardocabral/icuvisor/internal/mcp"
	"github.com/ricardocabral/icuvisor/internal/resources"
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

	configOpts, err := parseDefaultArgs(args)
	if err != nil {
		return err
	}

	return startServer(ctx, opts.LoadConfig, opts.StartServer, ServerInfo{Version: version, DebugMetadata: response.DebugMetadataFromEnv()}, configOpts)
}

func parseDefaultArgs(args []string) (config.Options, error) {
	var opts config.Options
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--config":
			value, next, err := requireFlagValue(args, i, "--config", "/path/to/icuvisor.json")
			if err != nil {
				return config.Options{}, err
			}
			opts.Path = value
			i = next
		case strings.HasPrefix(arg, "--config="):
			value, err := requireInlineFlagValue(arg, "--config", "/path/to/icuvisor.json")
			if err != nil {
				return config.Options{}, err
			}
			opts.Path = value
		case arg == "--transport":
			value, next, err := requireFlagValue(args, i, "--transport", "stdio|http")
			if err != nil {
				return config.Options{}, err
			}
			opts.Transport = value
			i = next
		case strings.HasPrefix(arg, "--transport="):
			value, err := requireInlineFlagValue(arg, "--transport", "stdio|http")
			if err != nil {
				return config.Options{}, err
			}
			opts.Transport = value
		case arg == "--http-bind":
			value, next, err := requireFlagValue(args, i, "--http-bind", "127.0.0.1:8765")
			if err != nil {
				return config.Options{}, err
			}
			opts.HTTPBindAddress = value
			i = next
		case strings.HasPrefix(arg, "--http-bind="):
			value, err := requireInlineFlagValue(arg, "--http-bind", "127.0.0.1:8765")
			if err != nil {
				return config.Options{}, err
			}
			opts.HTTPBindAddress = value
		default:
			return config.Options{}, fmt.Errorf("unknown command or flag %q (try: icuvisor version, --config, --transport, --http-bind)", arg)
		}
	}
	return opts, nil
}

func requireFlagValue(args []string, index int, name string, example string) (string, int, error) {
	next := index + 1
	if next >= len(args) || strings.TrimSpace(args[next]) == "" || strings.HasPrefix(args[next], "--") {
		return "", index, fmt.Errorf("missing value for %s; use %s %s", name, name, example)
	}
	return args[next], next, nil
}

func requireInlineFlagValue(arg string, name string, example string) (string, error) {
	value, _ := strings.CutPrefix(arg, name+"=")
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("missing value for %s; use %s %s", name, name, example)
	}
	return value, nil
}

func startServer(ctx context.Context, loader func(context.Context, config.Options) (config.Config, error), starter func(context.Context, ServerInfo) error, info ServerInfo, configOpts config.Options) error {
	if loader == nil {
		loader = config.Load
	}
	cfg, err := loader(ctx, configOpts)
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
	if info.Config.Transport == config.TransportHTTP && !config.HTTPBindAddressIsLoopback(info.Config.HTTPBindAddress) {
		logger.Warn("http transport non-loopback bind active", "transport", info.Config.Transport, "http_bind", info.Config.HTTPBindAddress, "security", "any host that can reach this address can connect")
	}

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
		ResourceRegistry: resources.NewRegistryWithOptions(client, resources.ResourceOptions{
			Version:          info.Version,
			TimezoneFallback: info.Config.Timezone,
			DebugMetadata:    info.DebugMetadata,
		}),
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
	if info.Config.Transport == config.TransportHTTP {
		return server.RunStreamableHTTP(ctx, info.Config.HTTPBindAddress)
	}
	return server.Run(ctx)
}

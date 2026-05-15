package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	mcpserver "github.com/ricardocabral/icuvisor/internal/mcp"
	"github.com/ricardocabral/icuvisor/internal/prompts"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

// Options contains process-level dependencies for the icuvisor CLI.
type Options struct {
	Version string
	Args    []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer

	LoadConfig  func(context.Context, config.Options) (config.Config, error)
	StartServer func(context.Context, ServerInfo) error

	SetupRunner           SetupRunner
	SetupCredentialStore  credstore.Store
	SetupPrompter         SetupPrompter
	SetupConfigExists     func(string) (bool, error)
	SetupProfileFetcher   SetupProfileFetcher
	SetupTimezoneDetector func() string
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

// RunCLI executes the icuvisor CLI, writes any error to opts.Stderr, and returns a process exit code.
func RunCLI(ctx context.Context, opts Options) int {
	stderr := opts.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	if err := Run(ctx, opts); err != nil {
		_, _ = fmt.Fprintf(stderr, "icuvisor: %v\n", err)
		return ExitCode(err)
	}
	return 0
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
	if len(args) > 0 && args[0] == "setup" {
		return runSetupCommand(ctx, opts, args[1:])
	}
	if helpRequested(args) {
		if hasCommand(args, "version") {
			return writeVersionHelp(stdout)
		}
		return writeTopLevelHelp(stdout)
	}
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

	return startServer(ctx, opts.LoadConfig, opts.StartServer, ServerInfo{Version: version}, configOpts)
}

// UsageError reports invalid CLI input that should exit with code 2.
type UsageError struct {
	message string
}

func (e UsageError) Error() string {
	return e.message
}

// ExitCode maps Run errors to process exit codes.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var usageErr UsageError
	if errors.As(err, &usageErr) {
		return 2
	}
	return 1
}

func newUsageError(format string, args ...any) UsageError {
	return UsageError{message: fmt.Sprintf(format, args...) + "\nRun 'icuvisor --help' for usage."}
}

func helpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

func hasCommand(args []string, command string) bool {
	for _, arg := range args {
		if arg == command {
			return true
		}
	}
	return false
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
		case arg == "--env-file":
			value, next, err := requireFlagValue(args, i, "--env-file", "/path/to/icuvisor.env")
			if err != nil {
				return config.Options{}, err
			}
			opts.DotEnvPath = value
			opts.DotEnvExplicit = true
			i = next
		case strings.HasPrefix(arg, "--env-file="):
			value, err := requireInlineFlagValue(arg, "--env-file", "/path/to/icuvisor.env")
			if err != nil {
				return config.Options{}, err
			}
			opts.DotEnvPath = value
			opts.DotEnvExplicit = true
		default:
			return config.Options{}, newUsageError("unknown command or flag %q (try: icuvisor version)", arg)
		}
	}
	return opts, nil
}

func requireFlagValue(args []string, index int, name string, example string) (string, int, error) {
	next := index + 1
	if next >= len(args) || strings.TrimSpace(args[next]) == "" || strings.HasPrefix(args[next], "--") {
		return "", index, newUsageError("missing value for %s; use %s %s", name, name, example)
	}
	return args[next], next, nil
}

func requireInlineFlagValue(arg string, name string, example string) (string, error) {
	value, _ := strings.CutPrefix(arg, name+"=")
	if strings.TrimSpace(value) == "" {
		return "", newUsageError("missing value for %s; use %s %s", name, name, example)
	}
	return value, nil
}

func startServer(ctx context.Context, loader func(context.Context, config.Options) (config.Config, error), starter func(context.Context, ServerInfo) error, info ServerInfo, configOpts config.Options) error {
	if loader == nil {
		loader = config.Load
		if configOpts.CredentialStore == nil {
			configOpts.CredentialStore = credstore.OSKeychain()
		}
	}
	cfg, err := loader(ctx, configOpts)
	if err != nil {
		return err
	}
	info.Config = cfg
	info.DebugMetadata = cfg.DebugMetadata
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
	safety.LogResolvedMode(logger, deleteMode)
	safety.LogResolvedToolset(logger, toolset)
	client, err := intervals.NewClient(intervals.Options{Config: info.Config, Version: info.Version})
	if err != nil {
		return err
	}
	server, err := mcpserver.NewServer(ctx, mcpserver.Options{
		Config:         info.Config,
		Version:        info.Version,
		Logger:         logger,
		Capability:     capability,
		Toolset:        toolset,
		PromptRegistry: prompts.NewRegistry(),
		ResourceRegistry: resources.NewRegistryWithOptions(client, resources.ResourceOptions{
			Version:          info.Version,
			TimezoneFallback: info.Config.Timezone,
			DebugMetadata:    info.DebugMetadata,
			DeleteMode:       deleteMode,
			Toolset:          toolset,
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

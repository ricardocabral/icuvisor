// Package cli implements the standalone command-line view over icuvisor core tools.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/tools"
	core "github.com/ricardocabral/icuvisor/pkg/icuvisor"
)

// Options configures the standalone CLI process.
type Options struct {
	Version string
	Args    []string
	Stdout  io.Writer
	Stderr  io.Writer

	LoadConfig func(context.Context, config.Options) (config.Config, error)
	NewClient  func(core.APIKeyClientOptions) (*core.Client, error)
	NewInvoker func(context.Context, core.InvokerOptions) (*core.Invoker, error)
}

// RunCLI runs the standalone CLI and returns a process exit code.
func RunCLI(ctx context.Context, opts Options) int {
	stderr := opts.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	if err := Run(ctx, opts); err != nil {
		_, _ = fmt.Fprintf(stderr, "icuvisor-cli: %v\n", err)
		return 1
	}
	return 0
}

// Run runs the standalone CLI.
func Run(ctx context.Context, opts Options) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	version := strings.TrimSpace(opts.Version)
	if version == "" {
		version = "dev"
	}
	if len(opts.Args) == 0 || isHelp(opts.Args) {
		return writeHelp(stdout)
	}
	if opts.Args[0] == "version" {
		_, err := fmt.Fprintln(stdout, version)
		return err
	}

	command := opts.Args[0]
	arguments := opts.Args[1:]
	switch command {
	case "list":
		invoker, err := newInvoker(ctx, opts, arguments)
		if err != nil {
			return err
		}
		return writeJSON(stdout, invoker.Tools())
	case "describe":
		toolName, runtimeArgs, err := splitToolName(arguments, "describe")
		if err != nil {
			return err
		}
		invoker, err := newInvoker(ctx, opts, runtimeArgs)
		if err != nil {
			return err
		}
		for _, tool := range invoker.Tools() {
			if tool.Name == toolName {
				return writeJSON(stdout, tool)
			}
		}
		return fmt.Errorf("unknown or unavailable tool %q", toolName)
	case "call":
		toolName, runtimeArgs, err := splitToolName(arguments, "call")
		if err != nil {
			return err
		}
		request, configOpts, err := parseCallArgs(runtimeArgs)
		if err != nil {
			return err
		}
		invoker, err := loadInvoker(ctx, opts, configOpts)
		if err != nil {
			return err
		}
		result, err := invoker.Invoke(ctx, toolName, request)
		if err != nil {
			slog.Default().Error("standalone tool invocation failed", "tool", toolName, "error", err)
			return fmt.Errorf("calling %s: %s", toolName, publicToolErrorMessage(err))
		}
		if result.IsError {
			return fmt.Errorf("calling %s: %s", toolName, resultErrorMessage(result))
		}
		return writeJSON(stdout, result.StructuredContent)
	default:
		return fmt.Errorf("unknown command %q; use list, describe, call, or version", command)
	}
}

func newInvoker(ctx context.Context, opts Options, args []string) (*core.Invoker, error) {
	configOpts, err := parseRuntimeArgs(args)
	if err != nil {
		return nil, err
	}
	return loadInvoker(ctx, opts, configOpts)
}

func loadInvoker(ctx context.Context, opts Options, configOpts config.Options) (*core.Invoker, error) {
	if configOpts.CredentialStore == nil {
		configOpts.CredentialStore = credstore.OSKeychain()
	}
	configOpts.Path = defaultConfigPath(configOpts.Path)
	loader := opts.LoadConfig
	if loader == nil {
		loader = config.Load
	}
	cfg, err := loader(ctx, configOpts)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	if cfg.CoachModeEnabled() {
		return nil, fmt.Errorf("standalone CLI does not support coach mode; use the MCP server")
	}
	coreConfig := coreConfigFromInternal(cfg)
	newClient := opts.NewClient
	if newClient == nil {
		newClient = core.NewAPIKeyClient
	}
	client, err := newClient(core.APIKeyClientOptions{Config: coreConfig, Version: opts.Version})
	if err != nil {
		return nil, fmt.Errorf("creating intervals client: %w", err)
	}
	registry := core.NewCoreRegistry(client, core.RegistryOptions{
		Config:           coreConfig,
		Version:          opts.Version,
		TimezoneFallback: cfg.Timezone,
		DebugMetadata:    cfg.DebugMetadata,
		DeleteMode:       core.DeleteMode(cfg.DeleteMode.String()),
		Toolset:          core.Toolset(cfg.Toolset.String()),
	})
	newInvoker := opts.NewInvoker
	if newInvoker == nil {
		newInvoker = core.NewInvoker
	}
	invoker, err := newInvoker(ctx, core.InvokerOptions{
		Config:     coreConfig,
		Registry:   registry,
		DeleteMode: core.DeleteMode(cfg.DeleteMode.String()),
		Toolset:    core.Toolset(cfg.Toolset.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("creating direct tool invoker: %w", err)
	}
	return invoker, nil
}

func coreConfigFromInternal(cfg config.Config) core.Config {
	return core.Config{
		APIKey:        cfg.APIKey,
		AthleteID:     cfg.AthleteID,
		Timezone:      cfg.Timezone,
		APIBaseURL:    cfg.APIBaseURL,
		HTTPTimeout:   cfg.HTTPTimeout,
		DeleteMode:    core.DeleteMode(cfg.DeleteMode.String()),
		Toolset:       core.Toolset(cfg.Toolset.String()),
		DebugMetadata: cfg.DebugMetadata,
	}
}

func publicToolErrorMessage(err error) string {
	if message, ok := tools.PublicErrorMessage(err); ok {
		return message
	}
	return "tool failed; check local logs"
}

func resultErrorMessage(result core.ToolResult) string {
	for _, item := range result.Content {
		if strings.TrimSpace(item.Text) != "" {
			return item.Text
		}
	}
	return "tool returned an error result"
}

func splitToolName(args []string, command string) (string, []string, error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return "", nil, fmt.Errorf("%s requires a tool name", command)
	}
	return args[0], args[1:], nil
}

func parseCallArgs(args []string) (json.RawMessage, config.Options, error) {
	var request json.RawMessage
	argumentSource := ""
	filtered := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--args":
			if argumentSource != "" {
				return nil, config.Options{}, fmt.Errorf("use only one of --args or --args-file")
			}
			value, next, err := flagValue(args, index, "--args")
			if err != nil {
				return nil, config.Options{}, err
			}
			request = json.RawMessage(value)
			argumentSource = "--args"
			index = next
		case "--args-file":
			if argumentSource != "" {
				return nil, config.Options{}, fmt.Errorf("use only one of --args or --args-file")
			}
			path, next, err := flagValue(args, index, "--args-file")
			if err != nil {
				return nil, config.Options{}, err
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return nil, config.Options{}, fmt.Errorf("reading --args-file: %w", err)
			}
			request = body
			argumentSource = "--args-file"
			index = next
		default:
			filtered = append(filtered, arg)
		}
	}
	if request == nil {
		request = json.RawMessage(`{}`)
	}
	if !json.Valid(request) {
		return nil, config.Options{}, fmt.Errorf("tool arguments must be valid JSON")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(request, &object); err != nil {
		return nil, config.Options{}, fmt.Errorf("tool arguments must be a JSON object")
	}
	configOpts, err := parseRuntimeArgs(filtered)
	return request, configOpts, err
}

func parseRuntimeArgs(args []string) (config.Options, error) {
	var opts config.Options
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--config":
			value, next, err := flagValue(args, index, "--config")
			if err != nil {
				return config.Options{}, err
			}
			opts.Path = value
			index = next
		case "--env-file":
			value, next, err := flagValue(args, index, "--env-file")
			if err != nil {
				return config.Options{}, err
			}
			opts.DotEnvPath = value
			opts.DotEnvExplicit = true
			index = next
		default:
			return config.Options{}, fmt.Errorf("unknown flag %q", arg)
		}
	}
	return opts, nil
}

func flagValue(args []string, index int, name string) (string, int, error) {
	next := index + 1
	if next >= len(args) || strings.TrimSpace(args[next]) == "" || strings.HasPrefix(args[next], "--") {
		return "", index, fmt.Errorf("missing value for %s", name)
	}
	return args[next], next, nil
}

func defaultConfigPath(path string) string {
	if strings.TrimSpace(path) != "" || strings.TrimSpace(os.Getenv(config.EnvConfigPath)) != "" {
		return path
	}
	defaultPath, err := config.DefaultPath()
	if err != nil {
		return path
	}
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	return path
}

func isHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("writing JSON: %w", err)
	}
	return nil
}

func writeHelp(w io.Writer) error {
	_, err := fmt.Fprint(w, `Invoke icuvisor tools directly without an MCP transport.

Usage:
  icuvisor-cli list [runtime flags]
  icuvisor-cli describe <tool> [runtime flags]
  icuvisor-cli call <tool> [--args <json> | --args-file <path>] [runtime flags]
  icuvisor-cli version

Commands:
  list       Print the direct-view tool catalog as JSON.
  describe   Print one tool's name, description, and JSON schemas.
  call       Invoke one tool and print its structured result as JSON.
  version    Print the icuvisor-cli version.

Runtime flags:
  --config <path>    JSON config file path. Can also be set with ICUVISOR_CONFIG.
  --env-file <path>  Env-file path to load before process env. Can also be set with ICUVISOR_ENV_FILE.

Argument flags:
  --args <json>       JSON object passed to the tool. Defaults to {}.
  --args-file <path>  File containing the JSON object passed to the tool.

Conventions:
  JSON results are written to stdout. Diagnostics and errors are written to stderr.
  The standalone CLI uses the configured local athlete only; coach mode remains MCP-only.
  API keys are loaded from the OS keychain, config, or environment and are never tool arguments.
`)
	if err != nil {
		return fmt.Errorf("writing help: %w", err)
	}
	return nil
}

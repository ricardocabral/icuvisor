// Package cli implements the standalone command-line view over icuvisor core tools.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	core "github.com/ricardocabral/icuvisor/pkg/icuvisor"
)

const cliContractVersion = "1"

// Options configures the standalone CLI process.
type Options struct {
	Version string
	Args    []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer

	LoadConfig func(context.Context, config.Options) (config.Config, error)
	NewClient  func(core.APIKeyClientOptions) (*core.Client, error)
	NewInvoker func(context.Context, core.InvokerOptions) (*core.Invoker, error)
}

type usageError struct{ message string }

func (e *usageError) Error() string { return e.message }

func usage(message string, args ...any) error {
	return &usageError{message: fmt.Sprintf(message, args...)}
}

// RunCLI runs the standalone CLI and returns a process exit code.
func RunCLI(ctx context.Context, opts Options) int {
	err := Run(ctx, opts)
	if err == nil {
		return 0
	}
	code := 1
	var invalidUsage *usageError
	if errors.As(err, &invalidUsage) {
		code = 2
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	_ = json.NewEncoder(stderr).Encode(map[string]any{"error": err.Error(), "exit_code": code})
	return code
}

// Run runs the standalone CLI.
func Run(ctx context.Context, opts Options) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	version := normalizedVersion(opts.Version)
	if len(opts.Args) == 0 || isHelp(opts.Args) {
		return writeHelp(stdout)
	}
	if opts.Args[0] == "version" {
		if len(opts.Args) != 1 {
			return usage("version does not accept arguments")
		}
		_, err := fmt.Fprintln(stdout, version)
		return err
	}

	switch opts.Args[0] {
	case "tools":
		return runTools(ctx, opts, stdout, opts.Args[1:])
	case "doctor":
		return runDoctor(ctx, opts, stdout, opts.Args[1:])
	case "capabilities":
		return runCapabilities(ctx, opts, stdout, opts.Args[1:])
	default:
		return usage("unknown command %q; use tools, doctor, capabilities, version, or help", opts.Args[0])
	}
}

func runTools(ctx context.Context, opts Options, stdout io.Writer, args []string) error {
	if len(args) == 0 {
		return usage("tools requires a subcommand: list, describe, or call")
	}
	switch args[0] {
	case "list":
		invoker, err := newInvoker(ctx, opts, args[1:])
		if err != nil {
			return err
		}
		entries := make([]toolListEntry, 0, len(invoker.Tools()))
		for _, tool := range invoker.Tools() {
			entries = append(entries, toolListEntry{Name: tool.Name, Summary: shortSummary(tool.Description), Toolset: tool.Toolset, Safety: tool.Requirement})
		}
		return writeJSON(stdout, toolList{Header: invoker.Info(), Tools: entries})
	case "describe":
		toolName, runtimeArgs, err := splitToolName(args[1:], "tools describe")
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
		toolName, runtimeArgs, err := splitToolName(args[1:], "tools call")
		if err != nil {
			return err
		}
		request, configOpts, err := parseCallArgs(runtimeArgs, stdinOrDefault(opts.Stdin))
		if err != nil {
			return err
		}
		invoker, err := loadInvoker(ctx, opts, configOpts)
		if err != nil {
			return err
		}
		result, err := invoker.Invoke(ctx, toolName, request)
		if err != nil {
			return err
		}
		if result.IsError {
			return fmt.Errorf("calling %s: %s", toolName, resultErrorMessage(result))
		}
		return writeJSON(stdout, result.StructuredContent)
	default:
		return usage("unknown tools subcommand %q; use list, describe, or call", args[0])
	}
}

type toolList struct {
	Header core.InvokerInfo `json:"header"`
	Tools  []toolListEntry  `json:"tools"`
}

type toolListEntry struct {
	Name    string           `json:"name"`
	Summary string           `json:"summary"`
	Toolset core.Toolset     `json:"toolset"`
	Safety  core.Requirement `json:"safety"`
}

func runCapabilities(ctx context.Context, opts Options, stdout io.Writer, args []string) error {
	invoker, err := newInvoker(ctx, opts, args)
	if err != nil {
		return err
	}
	return writeJSON(stdout, map[string]any{
		"cli_contract_version": cliContractVersion,
		"icuvisor_version":     invoker.Info().Version,
		"catalog_hash":         invoker.Info().CatalogHash,
		"active_gates": map[string]any{
			"toolset":     invoker.Info().Toolset,
			"delete_mode": invoker.Info().DeleteMode,
		},
		"surfaces": map[string]any{
			"tools":     true,
			"resources": false,
			"prompts":   false,
		},
		"coach_mode": map[string]any{"supported": false},
	})
}

func runDoctor(ctx context.Context, opts Options, stdout io.Writer, args []string) error {
	invoker, err := newInvoker(ctx, opts, args)
	if err != nil {
		return err
	}
	reachability := "ok"
	ready := true
	result, callErr := invoker.Invoke(ctx, "get_athlete_profile", json.RawMessage(`{}`))
	if callErr != nil || result.IsError {
		reachability = "failed"
		ready = false
	}
	return writeJSON(stdout, map[string]any{
		"ready": ready,
		"checks": map[string]string{
			"config":         "ok",
			"authentication": "configured",
			"reachability":   reachability,
		},
	})
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
		loader = loadConfigQuietly
	}
	cfg, err := loader(ctx, configOpts)
	if err != nil {
		return nil, errors.New("loading configuration failed; check icuvisor setup and local config")
	}
	if cfg.CoachModeEnabled() {
		return nil, errors.New("standalone CLI does not support coach mode; use the MCP server")
	}
	coreConfig := coreConfigFromInternal(cfg)
	newClient := opts.NewClient
	if newClient == nil {
		newClient = core.NewAPIKeyClient
	}
	client, err := newClient(core.APIKeyClientOptions{Config: coreConfig, Version: opts.Version})
	if err != nil {
		return nil, errors.New("creating intervals client failed; check local configuration")
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
		Version:    normalizedVersion(opts.Version),
		DeleteMode: core.DeleteMode(cfg.DeleteMode.String()),
		Toolset:    core.Toolset(cfg.Toolset.String()),
	})
	if err != nil {
		return nil, errors.New("creating direct tool invoker failed; check local configuration")
	}
	return invoker, nil
}

func loadConfigQuietly(ctx context.Context, opts config.Options) (config.Config, error) {
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer slog.SetDefault(previous)
	return config.Load(ctx, opts)
}

func coreConfigFromInternal(cfg config.Config) core.Config {
	return core.Config{APIKey: cfg.APIKey, AthleteID: cfg.AthleteID, Timezone: cfg.Timezone, APIBaseURL: cfg.APIBaseURL, HTTPTimeout: cfg.HTTPTimeout, DeleteMode: core.DeleteMode(cfg.DeleteMode.String()), Toolset: core.Toolset(cfg.Toolset.String()), DebugMetadata: cfg.DebugMetadata}
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
		return "", nil, usage("%s requires a tool name", command)
	}
	return args[0], args[1:], nil
}

func parseCallArgs(args []string, stdin io.Reader) (json.RawMessage, config.Options, error) {
	var request json.RawMessage
	argumentSource := ""
	filtered := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--args":
			if argumentSource != "" {
				return nil, config.Options{}, usage("use only one of --args or --args-file")
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
				return nil, config.Options{}, usage("use only one of --args or --args-file")
			}
			path, next, err := flagValue(args, index, "--args-file")
			if err != nil {
				return nil, config.Options{}, err
			}
			var body []byte
			if path == "-" {
				body, err = io.ReadAll(stdin)
			} else {
				body, err = os.ReadFile(path)
			}
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
		return nil, config.Options{}, usage("tool arguments must be valid JSON")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(request, &object); err != nil || object == nil {
		return nil, config.Options{}, usage("tool arguments must be a JSON object")
	}
	configOpts, err := parseRuntimeArgs(filtered)
	return request, configOpts, err
}

func parseRuntimeArgs(args []string) (config.Options, error) {
	var opts config.Options
	for index := 0; index < len(args); index++ {
		switch args[index] {
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
			return config.Options{}, usage("unknown flag %q", args[index])
		}
	}
	return opts, nil
}

func flagValue(args []string, index int, name string) (string, int, error) {
	next := index + 1
	if next >= len(args) || strings.TrimSpace(args[next]) == "" || strings.HasPrefix(args[next], "--") {
		return "", index, usage("missing value for %s", name)
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
	if len(args) > 0 && args[0] == "help" {
		return true
	}
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func normalizedVersion(version string) string {
	if version = strings.TrimSpace(version); version != "" {
		return version
	}
	return "dev"
}

func stdinOrDefault(r io.Reader) io.Reader {
	if r != nil {
		return r
	}
	return os.Stdin
}

func shortSummary(description string) string {
	description = strings.TrimSpace(description)
	if end := strings.Index(description, ". "); end >= 0 {
		description = description[:end+1]
	}
	const max = 180
	if len(description) > max {
		description = strings.TrimSpace(description[:max-3]) + "..."
	}
	return description
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
  icuvisor-cli tools list [runtime flags]
  icuvisor-cli tools describe <tool> [runtime flags]
  icuvisor-cli tools call <tool> [--args <json> | --args-file <path> | --args-file -] [runtime flags]
  icuvisor-cli doctor [runtime flags]
  icuvisor-cli capabilities [runtime flags]
  icuvisor-cli version

Commands:
  tools list       Print a compact tool catalog and active-policy header.
  tools describe   Print one canonical MCP tool descriptor.
  tools call       Invoke one tool and print its structured result.
  doctor           Check redacted config, authentication, and reachability readiness.
  capabilities     Print the CLI contract, supported surfaces, and active gates.
  version          Print the icuvisor-cli version.

Runtime flags:
  --config <path>    JSON config file path. Can also be set with ICUVISOR_CONFIG.
  --env-file <path>  Env-file path to load before process env. Can also be set with ICUVISOR_ENV_FILE.

Argument flags:
  --args <json>       JSON object passed to the tool. Defaults to {}.
  --args-file <path>  File containing the JSON object passed to the tool.
  --args-file -       Read the JSON object from stdin.

Process contract:
  Success writes JSON to stdout only. Failures leave stdout empty and write one JSON error to stderr.
  Exit codes: 0 success, 2 usage error, 1 runtime error.
  Output is deterministic, non-interactive, and contains no ANSI formatting.
  The standalone CLI uses the configured local athlete only; coach mode remains MCP-only.
  API keys are loaded from configuration or the OS keychain and are never flags or tool arguments.
`)
	if err != nil {
		return fmt.Errorf("writing help: %w", err)
	}
	return nil
}

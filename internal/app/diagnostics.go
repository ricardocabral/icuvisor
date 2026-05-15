package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/diagnostics"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	mcpserver "github.com/ricardocabral/icuvisor/internal/mcp"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

const diagnosticsRecentToolCallCount = 10

func runDiagnosticsCommand(ctx context.Context, opts Options, args []string) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	if helpRequested(args) {
		return writeDiagnosticsHelp(stdout)
	}
	configOpts, err := parseDefaultArgs(args)
	if err != nil {
		return err
	}
	loader := opts.LoadConfig
	if loader == nil {
		loader = config.Load
		if configOpts.CredentialStore == nil {
			configOpts.CredentialStore = credstore.OSKeychain()
		}
	}
	cfg, err := loader(ctx, configOpts)
	if err != nil {
		_, writeErr := fmt.Fprintf(stdout, "icuvisor diagnostics\nversion: %s\nconfig_source: error\n", normalizedVersion(opts.Version))
		if writeErr != nil {
			return fmt.Errorf("writing diagnostics: %w", writeErr)
		}
		return errors.New("loading config for diagnostics")
	}

	catalogHash, err := diagnosticsCatalogHash(ctx, cfg, normalizedVersion(opts.Version))
	if err != nil {
		return errors.New("computing diagnostics catalog hash")
	}
	recentPath := opts.DiagnosticsRecentToolCallsPath
	if strings.TrimSpace(recentPath) == "" {
		recentPath, _ = diagnostics.DefaultRecentToolCallsPath()
	}
	recent, recentErr := diagnostics.ReadRecentToolCalls(ctx, recentPath, diagnosticsRecentToolCallCount)
	if recentErr != nil {
		recent = nil
	}

	return writeDiagnostics(stdout, diagnosticsSnapshot{
		Version:        normalizedVersion(opts.Version),
		CatalogHash:    catalogHash,
		ConfigSource:   diagnosticsConfigSource(cfg.APIKeySource),
		Transport:      cfg.Transport.String(),
		DeleteMode:     safety.ParseMode(cfg.DeleteMode.String()).String(),
		Toolset:        safety.ParseToolset(cfg.Toolset.String()).String(),
		CoachMode:      diagnosticsCoachMode(cfg),
		Runtime:        runtime.GOOS + "/" + runtime.GOARCH + " " + runtime.Version(),
		RecentToolCall: recent,
		RecentReadable: recentErr == nil,
	})
}

type diagnosticsSnapshot struct {
	Version        string
	CatalogHash    string
	ConfigSource   string
	Transport      string
	DeleteMode     string
	Toolset        string
	CoachMode      string
	Runtime        string
	RecentToolCall []diagnostics.RecentToolCall
	RecentReadable bool
}

func diagnosticsCatalogHash(ctx context.Context, cfg config.Config, version string) (string, error) {
	client, err := intervals.NewClient(intervals.Options{
		Config:     cfg,
		Version:    version,
		HTTPClient: &http.Client{Transport: diagnosticsNoNetworkRoundTripper{}},
	})
	if err != nil {
		return "", err
	}
	capability := safety.NewCapability(cfg.DeleteMode)
	registry := tools.NewRegistryWithOptions(client, tools.RegistryOptions{
		Version:          version,
		TimezoneFallback: cfg.Timezone,
		DebugMetadata:    cfg.DebugMetadata,
		Capability:       capability,
		Toolset:          cfg.Toolset,
		CoachModeEnabled: cfg.CoachModeEnabled(),
		CoachConfig:      cfg.Coach,
	})
	return mcpserver.ComputeToolCatalogHash(ctx, mcpserver.CatalogHashOptions{
		Config:     cfg,
		Registry:   registry,
		Capability: capability,
		Toolset:    cfg.Toolset,
	})
}

type diagnosticsNoNetworkRoundTripper struct{}

func (diagnosticsNoNetworkRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("diagnostics must not perform intervals.icu HTTP requests")
}

func writeDiagnostics(w io.Writer, snapshot diagnosticsSnapshot) error {
	if _, err := fmt.Fprintf(w, "icuvisor diagnostics\nversion: %s\ncatalog_hash: %s\nconfig_source: %s\ntransport: %s\n%s: %s\n%s: %s\n%s: %s\nruntime: %s\nrecent_tool_calls:\n",
		snapshot.Version,
		snapshot.CatalogHash,
		snapshot.ConfigSource,
		snapshot.Transport,
		safety.EnvDeleteMode,
		snapshot.DeleteMode,
		safety.EnvToolset,
		snapshot.Toolset,
		config.EnvCoachMode,
		snapshot.CoachMode,
		snapshot.Runtime,
	); err != nil {
		return fmt.Errorf("writing diagnostics: %w", err)
	}
	if !snapshot.RecentReadable {
		_, err := fmt.Fprintln(w, "  unavailable")
		if err != nil {
			return fmt.Errorf("writing diagnostics: %w", err)
		}
		return nil
	}
	if len(snapshot.RecentToolCall) == 0 {
		_, err := fmt.Fprintln(w, "  none")
		if err != nil {
			return fmt.Errorf("writing diagnostics: %w", err)
		}
		return nil
	}
	for _, call := range snapshot.RecentToolCall {
		if _, err := fmt.Fprintf(w, "  - %s %s\n", call.Timestamp.UTC().Format(time.RFC3339), call.Name); err != nil {
			return fmt.Errorf("writing diagnostics: %w", err)
		}
	}
	return nil
}

func defaultRecentToolCallRecorder() (diagnostics.RecentToolCallRecorder, error) {
	path, err := diagnostics.DefaultRecentToolCallsPath()
	if err != nil {
		return nil, err
	}
	return diagnostics.NewRecentToolCallStore(path), nil
}

func diagnosticsConfigSource(source config.APIKeySource) string {
	source = config.APIKeySource(strings.TrimSpace(string(source)))
	switch source {
	case config.APIKeySourceEnv, config.APIKeySourceKeychain, config.APIKeySourceFile:
		return string(source)
	default:
		return "unset"
	}
}

func diagnosticsCoachMode(cfg config.Config) string {
	mode := strings.TrimSpace(string(cfg.EffectiveCoachMode()))
	if mode == "" {
		return "off"
	}
	return mode
}

func normalizedVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "dev"
	}
	return version
}

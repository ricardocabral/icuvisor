package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	diagnosticsdata "github.com/ricardocabral/icuvisor/internal/diagnostics"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestRunDiagnosticsPrintsRedactedMetadataAndBypassesServer(t *testing.T) {
	t.Parallel()

	secret := "sk-" + strings.Repeat("x", 40)
	rawAthleteID := "7777777"
	normalizedAthleteID := "i" + rawAthleteID
	recentPath := t.TempDir() + "/recent-tool-calls.jsonl"
	store := diagnosticsdata.NewRecentToolCallStore(recentPath)
	if err := store.RecordToolCall(context.Background(), "get_activities", time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("RecordToolCall() error = %v", err)
	}
	if err := store.RecordToolCall(context.Background(), "update_wellness", time.Date(2026, 5, 15, 11, 30, 0, 0, time.UTC)); err != nil {
		t.Fatalf("RecordToolCall() error = %v", err)
	}

	var stdout bytes.Buffer
	var gotConfig config.Options
	err := Run(context.Background(), Options{
		Version:                        "v0.5.0-test",
		Args:                           []string{"diagnostics", "--config", "/tmp/secret-config.json", "--transport", "http", "--http-bind", "127.0.0.1:9999"},
		Stdout:                         &stdout,
		DiagnosticsRecentToolCallsPath: recentPath,
		LoadConfig: func(_ context.Context, opts config.Options) (config.Config, error) {
			gotConfig = opts
			return diagnosticsTestConfig(secret, rawAthleteID, config.APIKeySourceEnv, config.TransportHTTP, safety.ModeFull, safety.ToolsetFull), nil
		},
		StartServer: func(context.Context, ServerInfo) error {
			t.Fatal("diagnostics must not start the MCP server")
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if gotConfig.Path != "/tmp/secret-config.json" || gotConfig.Transport != "http" || gotConfig.HTTPBindAddress != "127.0.0.1:9999" {
		t.Fatalf("config options = %#v, want diagnostics flags passed through", gotConfig)
	}
	out := stdout.String()
	for _, want := range []string{
		"icuvisor diagnostics",
		"version: v0.5.0-test",
		"catalog_hash: ",
		"config_source: env",
		"transport: http",
		"ICUVISOR_DELETE_MODE: full",
		"ICUVISOR_TOOLSET: full",
		"ICUVISOR_COACH_MODE: off",
		"runtime: ",
		"recent_tool_calls:",
		"2026-05-14T10:00:00Z get_activities",
		"2026-05-15T11:30:00Z update_wellness",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("diagnostics output missing %q:\n%s", want, out)
		}
	}
	assertDiagnosticsNoSecrets(t, out, secret, rawAthleteID, normalizedAthleteID)
}

func TestRunDiagnosticsLoadErrorIsSanitized(t *testing.T) {
	t.Parallel()

	secret := "sk-" + strings.Repeat("y", 40)
	rawAthleteID := "8888888"
	var stdout, stderr bytes.Buffer
	code := RunCLI(context.Background(), Options{
		Args:   []string{"diagnostics"},
		Stdout: &stdout,
		Stderr: &stderr,
		LoadConfig: func(context.Context, config.Options) (config.Config, error) {
			return config.Config{}, fmt.Errorf("bad config with %s and i%s", secret, rawAthleteID)
		},
	})
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	combined := stdout.String() + stderr.String()
	for _, want := range []string{"config_source: error", "icuvisor: loading config for diagnostics"} {
		if !strings.Contains(combined, want) {
			t.Fatalf("diagnostics error output missing %q:\n%s", want, combined)
		}
	}
	assertDiagnosticsNoSecrets(t, combined, secret, rawAthleteID, "i"+rawAthleteID)
}

func TestRunDiagnosticsHelpAndFlagErrors(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := Run(context.Background(), Options{
		Args:   []string{"diagnostics", "--help"},
		Stdout: &stdout,
		LoadConfig: func(context.Context, config.Options) (config.Config, error) {
			t.Fatal("diagnostics help must not load config")
			return config.Config{}, nil
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	for _, want := range []string{"Usage:\n  icuvisor diagnostics [flags]", "Diagnostics does not start the MCP server"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("diagnostics help missing %q:\n%s", want, stdout.String())
		}
	}
	if err := Run(context.Background(), Options{Args: []string{"diagnostics", "--bogus"}}); err == nil {
		t.Fatal("Run() error = nil, want usage error")
	} else {
		var usageErr UsageError
		if !errors.As(err, &usageErr) {
			t.Fatalf("Run() error = %T, want UsageError", err)
		}
	}
}

func TestDiagnosticsCatalogHashChangesWithCatalogMode(t *testing.T) {
	t.Parallel()

	safeCore := diagnosticsTestConfig("secret-key", "9999999", config.APIKeySourceKeychain, config.TransportStdio, safety.ModeSafe, safety.ToolsetCore)
	fullCatalog := diagnosticsTestConfig("secret-key", "9999999", config.APIKeySourceKeychain, config.TransportStdio, safety.ModeFull, safety.ToolsetFull)
	safeHash, err := diagnosticsCatalogHash(context.Background(), safeCore, "test")
	if err != nil {
		t.Fatalf("diagnosticsCatalogHash(safe) error = %v", err)
	}
	fullHash, err := diagnosticsCatalogHash(context.Background(), fullCatalog, "test")
	if err != nil {
		t.Fatalf("diagnosticsCatalogHash(full) error = %v", err)
	}
	if safeHash == fullHash {
		t.Fatalf("catalog hash did not change across safe/core and full/full modes: %s", safeHash)
	}
}

func diagnosticsTestConfig(apiKey, athleteID string, source config.APIKeySource, transport config.Transport, deleteMode safety.Mode, toolset safety.Toolset) config.Config {
	return config.Config{
		APIKey:          apiKey,
		APIKeySource:    source,
		AthleteID:       athleteID,
		Timezone:        "UTC",
		APIBaseURL:      "http://127.0.0.1",
		HTTPTimeout:     time.Second,
		Transport:       transport,
		HTTPBindAddress: "127.0.0.1:9999",
		DeleteMode:      deleteMode,
		Toolset:         toolset,
	}
}

func assertDiagnosticsNoSecrets(t *testing.T, out string, secret string, athleteIDs ...string) {
	t.Helper()
	if strings.Contains(out, secret) {
		t.Fatalf("diagnostics output leaked API key %q:\n%s", secret, out)
	}
	for _, athleteID := range athleteIDs {
		if strings.Contains(out, athleteID) {
			t.Fatalf("diagnostics output leaked athlete ID %q:\n%s", athleteID, out)
		}
	}
	redacted := out
	if catalogHash := regexp.MustCompile(`catalog_hash: ([a-f0-9]{64})`).FindStringSubmatch(out); len(catalogHash) == 2 {
		redacted = strings.ReplaceAll(out, catalogHash[1], "<catalog-hash>")
	}
	tokenPattern := regexp.MustCompile(`(?i)(api[_-]?key\s*[:=]\s*\S+|bearer\s+[a-z0-9._-]+|sk-[a-z0-9_-]{16,})`)
	if match := tokenPattern.FindString(redacted); match != "" {
		t.Fatalf("diagnostics output contained token-shaped string %q:\n%s", match, out)
	}
}

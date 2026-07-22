package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestRunToolsListIsCompactAndCarriesHeader(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := Run(context.Background(), Options{Args: []string{"tools", "list"}, Stdout: &stdout, Version: "vtest", LoadConfig: testLoadConfig})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	var parsed struct {
		Header struct {
			Version     string `json:"version"`
			Toolset     string `json:"toolset"`
			DeleteMode  string `json:"delete_mode"`
			CatalogHash string `json:"catalog_hash"`
		} `json:"header"`
		Tools []map[string]any `json:"tools"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if parsed.Header.Version != "vtest" || parsed.Header.Toolset != "core" || parsed.Header.DeleteMode != "safe" || parsed.Header.CatalogHash == "" {
		t.Fatalf("header = %#v", parsed.Header)
	}
	if len(parsed.Tools) == 0 {
		t.Fatal("tools list is empty")
	}
	for _, entry := range parsed.Tools {
		if entry["name"] == "get_athlete_profile" {
			if entry["summary"] == nil || entry["toolset"] != "core" || entry["safety"] != "read" {
				t.Fatalf("profile entry = %#v", entry)
			}
			if entry["inputSchema"] != nil || entry["outputSchema"] != nil {
				t.Fatalf("compact entry leaked schemas: %#v", entry)
			}
			return
		}
	}
	t.Fatal("list output missing get_athlete_profile")
}

func TestRunToolsDescribeUsesCanonicalMCPFields(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := Run(context.Background(), Options{Args: []string{"tools", "describe", "get_athlete_profile"}, Stdout: &stdout, Version: "vtest", LoadConfig: testLoadConfig})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	for _, key := range []string{"name", "description", "inputSchema", "outputSchema", "annotations", "_meta"} {
		if _, ok := parsed[key]; !ok {
			t.Fatalf("describe output missing %s: %s", key, stdout.Bytes())
		}
	}
	if parsed["input_schema"] != nil || parsed["output_schema"] != nil {
		t.Fatalf("describe output uses legacy snake_case fields: %s", stdout.Bytes())
	}
	meta := parsed["_meta"].(map[string]any)["icuvisor"].(map[string]any)
	if meta["toolset"] != "core" || meta["safety"] != "read" {
		t.Fatalf("_meta.icuvisor = %#v", meta)
	}
}

func TestRunToolsCallWritesStructuredJSON(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := Run(context.Background(), Options{Args: []string{"tools", "call", "icuvisor_check_server_version"}, Stdout: &stdout, Version: "vtest", LoadConfig: testLoadConfig})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"description_server_version": "vtest"`)) {
		t.Fatalf("call output = %s", stdout.Bytes())
	}
	if bytes.Contains(stdout.Bytes(), []byte(`structuredContent`)) {
		t.Fatalf("call output contains MCP envelope: %s", stdout.Bytes())
	}
}

func TestParseCallArgsReadsStdin(t *testing.T) {
	t.Parallel()

	request, _, err := parseCallArgs([]string{"--args-file", "-"}, strings.NewReader(`{"include_full":true}`))
	if err != nil {
		t.Fatalf("parseCallArgs() error = %v", err)
	}
	if string(request) != `{"include_full":true}` {
		t.Fatalf("request = %s", request)
	}
}

func TestParseCallArgsRejectsMultipleArgumentSources(t *testing.T) {
	t.Parallel()

	_, _, err := parseCallArgs([]string{"--args", `{}`, "--args-file", "request.json"}, strings.NewReader(""))
	if err == nil || !strings.Contains(err.Error(), "only one") {
		t.Fatalf("parseCallArgs() error = %v, want mutually-exclusive argument-source error", err)
	}
}

func TestRunCLIUsageFailureWritesOnlyMachineError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := RunCLI(context.Background(), Options{Args: []string{"tools", "call", "get_athlete_profile", "--args", `[]`}, Stdout: &stdout, Stderr: &stderr, LoadConfig: testLoadConfig})
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if bytes.Count(stderr.Bytes(), []byte("\n")) != 1 {
		t.Fatalf("stderr = %q, want one line", stderr.String())
	}
	var failure struct {
		Error    string `json:"error"`
		ExitCode int    `json:"exit_code"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &failure); err != nil || failure.ExitCode != 2 || !strings.Contains(failure.Error, "JSON object") {
		t.Fatalf("failure = %#v, unmarshal error = %v", failure, err)
	}
}

func TestRunCapabilitiesReportsToolsOnlyContract(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	if err := Run(context.Background(), Options{Args: []string{"capabilities"}, Stdout: &stdout, Version: "vtest", LoadConfig: testLoadConfig}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if parsed["cli_contract_version"] != cliContractVersion || parsed["catalog_hash"] == "" {
		t.Fatalf("capabilities = %#v", parsed)
	}
	surfaces := parsed["surfaces"].(map[string]any)
	if surfaces["tools"] != true || surfaces["resources"] != false || surfaces["prompts"] != false {
		t.Fatalf("surfaces = %#v", surfaces)
	}
}

func TestRunRefusesEffectiveCoachMode(t *testing.T) {
	t.Parallel()

	loader := func(context.Context, config.Options) (config.Config, error) {
		cfg, _ := testLoadConfig(context.Background(), config.Options{})
		cfg.CoachMode = coach.ModeOn
		return cfg, nil
	}
	err := Run(context.Background(), Options{Args: []string{"tools", "list"}, LoadConfig: loader})
	if err == nil || !strings.Contains(err.Error(), "does not support coach mode") {
		t.Fatalf("Run() error = %v", err)
	}
}

func testLoadConfig(context.Context, config.Options) (config.Config, error) {
	return config.Config{APIKey: "test-api-key", AthleteID: "i12345", Timezone: "UTC", APIBaseURL: config.DefaultAPIBaseURL, HTTPTimeout: time.Second, DeleteMode: safety.ModeSafe, Toolset: safety.ToolsetCore}, nil
}

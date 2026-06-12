package icuvisor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/response"
)

func TestBearerClientFacadeUsesBearerAuth(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/0"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if _, _, ok := r.BasicAuth(); ok {
			t.Fatal("BasicAuth() present, want bearer auth only")
		}
		if got, want := r.Header.Get("Authorization"), "Bearer public-bearer-token"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}
		if got, want := r.UserAgent(), "icuvisor/v-public"; got != want {
			t.Fatalf("User-Agent = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"0","name":"Hosted Bearer"}`))
	}))
	defer server.Close()

	client, err := NewBearerClient(BearerClientOptions{AccessToken: "public-bearer-token", APIBaseURL: server.URL, Version: "v-public", HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("NewBearerClient() error = %v", err)
	}
	profile, err := client.inner.GetAthleteProfile(context.Background())
	if err != nil {
		t.Fatalf("GetAthleteProfile() error = %v", err)
	}
	if profile.Name != "Hosted Bearer" {
		t.Fatalf("profile name = %q, want Hosted Bearer", profile.Name)
	}
}

func TestCoreRegistryFacadeFiltersAndAddsDiagnostics(t *testing.T) {
	t.Parallel()

	client := newFacadeTestClient(t)
	registry := NewCoreRegistry(client, RegistryOptions{
		Version: "v-public",
		Toolset: ToolsetFull,
		ToolFilter: func(tool ToolInfo) bool {
			return tool.Name != "get_fitness"
		},
		ExtraTools: []Tool{{
			Name:         "hosted_setup_status",
			Description:  "Report hosted connection setup status without athlete data.",
			InputSchema:  map[string]any{"type": "object", "additionalProperties": false},
			OutputSchema: map[string]any{"type": "object", "additionalProperties": true},
			Requirement:  RequirementRead,
			Toolset:      ToolsetCore,
			Handler: func(context.Context, ToolRequest) (ToolResult, error) {
				return TextResult(map[string]any{"status": "ok"}), nil
			},
		}},
	})

	catalog, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: facadeTestConfig(), Registry: registry, Mode: DeleteModeFull, Toolset: ToolsetFull})
	if err != nil {
		t.Fatalf("CollectToolCatalog() error = %v", err)
	}
	if hasTool(catalog, "get_fitness") {
		t.Fatal("filtered catalog contains get_fitness")
	}
	if !hasTool(catalog, "hosted_setup_status") {
		t.Fatal("catalog missing hosted_setup_status extra tool")
	}
}

func TestExtraToolSafetyMetadataFailsClosed(t *testing.T) {
	t.Parallel()

	invalidRequirement := facadeExtraTool("Invalid requirement.", ToolsetCore)
	invalidRequirement.Requirement = Requirement("deletee")
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public", Toolset: ToolsetFull, ExtraTools: []Tool{invalidRequirement}})
	if _, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: facadeTestConfig(), Registry: registry, Mode: DeleteModeFull, Toolset: ToolsetFull}); err == nil {
		t.Fatal("CollectToolCatalog() error = nil, want invalid requirement error")
	}

	invalidToolset := facadeExtraTool("Invalid toolset.", Toolset("coree"))
	registry = NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public", Toolset: ToolsetFull, ExtraTools: []Tool{invalidToolset}})
	if _, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: facadeTestConfig(), Registry: registry, Mode: DeleteModeFull, Toolset: ToolsetFull}); err == nil {
		t.Fatal("CollectToolCatalog() error = nil, want invalid toolset error")
	}
}

func TestExtraToolOmittedToolsetDefaultsToFull(t *testing.T) {
	t.Parallel()

	extra := facadeExtraTool("Omitted toolset should be full.", "")
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public", Toolset: ToolsetCore, ExtraTools: []Tool{extra}})
	coreCatalog, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: facadeTestConfig(), Registry: registry, Mode: DeleteModeSafe, Toolset: ToolsetCore})
	if err != nil {
		t.Fatalf("CollectToolCatalog(core) error = %v", err)
	}
	if hasTool(coreCatalog, "hosted_setup_status") {
		t.Fatal("core catalog includes extra tool with omitted toolset; want full-only default")
	}

	registry = NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public", Toolset: ToolsetFull, ExtraTools: []Tool{extra}})
	fullCatalog, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: facadeTestConfig(), Registry: registry, Mode: DeleteModeSafe, Toolset: ToolsetFull})
	if err != nil {
		t.Fatalf("CollectToolCatalog(full) error = %v", err)
	}
	if !hasTool(fullCatalog, "hosted_setup_status") {
		t.Fatal("full catalog missing extra tool with omitted toolset")
	}
}

func TestExtraToolsParticipateInCheckServerVersionFingerprint(t *testing.T) {
	t.Parallel()

	baseDescription := "Report hosted connection setup status without athlete data."
	changedDescription := "Report hosted connection setup status and reconnect guidance without athlete data."
	baseFingerprint := checkServerVersionFingerprintForExtra(t, baseDescription)
	changedFingerprint := checkServerVersionFingerprintForExtra(t, changedDescription)
	if baseFingerprint == "" || changedFingerprint == "" {
		t.Fatalf("fingerprints = %q/%q, want non-empty", baseFingerprint, changedFingerprint)
	}
	if baseFingerprint == changedFingerprint {
		t.Fatalf("extra tool description change did not change check-server-version fingerprint: %s", baseFingerprint)
	}
}

func TestServerFacadeCatalogHashAndSkipRuntimeMetadata(t *testing.T) {
	t.Parallel()

	response.SetRuntimeCatalogMetadata("before-version", "before-hash")
	t.Cleanup(func() { response.SetRuntimeCatalogMetadata("", "") })

	client := newFacadeTestClient(t)
	registry := NewCoreRegistry(client, RegistryOptions{Version: "v-public", Toolset: ToolsetCore})
	cfg := facadeTestConfig()
	wantHash, err := ComputeToolCatalogHash(context.Background(), CatalogOptions{Config: cfg, Registry: registry, Mode: DeleteModeSafe, Toolset: ToolsetCore})
	if err != nil {
		t.Fatalf("ComputeToolCatalogHash() error = %v", err)
	}
	server, err := NewServer(context.Background(), ServerOptions{Config: cfg, Version: "v-public", Registry: registry, ResourceRegistry: NewResourceRegistry(client, ResourceRegistryOptions{Version: "v-public"}), PromptRegistry: NewPromptRegistry(), DeleteMode: DeleteModeSafe, Toolset: ToolsetCore, SkipRuntimeCatalogMetadata: true})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if got := server.CatalogHash(); got != wantHash {
		t.Fatalf("server catalog hash = %q, want %q", got, wantHash)
	}
	if runtime := response.RuntimeCatalogMetadata(); runtime.CatalogHash != "before-hash" || runtime.Version != "before-version" {
		t.Fatalf("runtime metadata = %+v, want unchanged before-version/before-hash", runtime)
	}
}

func TestConfigValidationNormalizesAthleteIDForServerRouting(t *testing.T) {
	t.Parallel()

	cfg := facadeTestConfig()
	cfg.AthleteID = " I12345 "
	internalCfg, err := cfg.toInternalValidated()
	if err != nil {
		t.Fatalf("toInternalValidated() error = %v", err)
	}
	if internalCfg.AthleteID != "i12345" {
		t.Fatalf("AthleteID = %q, want i12345", internalCfg.AthleteID)
	}
	cfg.AthleteID = "bad athlete"
	if _, err := cfg.toInternalValidated(); err == nil {
		t.Fatal("toInternalValidated() error = nil, want invalid athlete ID error")
	}
}

func TestStreamableHTTPHandlerFacadeMapsFactoryError(t *testing.T) {
	t.Parallel()

	handler := NewStreamableHTTPHandler(func(*http.Request) (*Server, error) {
		return nil, errors.New("internal token refresh failed for sensitive-token")
	}, StreamableHTTPHandlerOptions{Stateless: true, FactoryErrorMessage: "hosted MCP authorization failed; reconnect Intervals in settings"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, StreamableHTTPPath, strings.NewReader(`{}`)))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "hosted MCP authorization failed; reconnect Intervals in settings") {
		t.Fatalf("body = %q, want public factory error message", body)
	}
	if strings.Contains(body, "sensitive-token") {
		t.Fatalf("body leaked internal error: %q", body)
	}
}

func checkServerVersionFingerprintForExtra(t *testing.T, description string) string {
	t.Helper()
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public", Toolset: ToolsetFull, ExtraTools: []Tool{facadeExtraTool(description, ToolsetCore)}})
	catalog, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: facadeTestConfig(), Registry: registry, Mode: DeleteModeSafe, Toolset: ToolsetFull})
	if err != nil {
		t.Fatalf("CollectToolCatalog() error = %v", err)
	}
	for _, tool := range catalog {
		if tool.Name != "icuvisor_check_server_version" {
			continue
		}
		marker := "description_catalog_fingerprint="
		start := strings.Index(tool.Description, marker)
		if start < 0 {
			t.Fatalf("check-server-version description missing fingerprint: %q", tool.Description)
		}
		start += len(marker)
		end := strings.Index(tool.Description[start:], ";")
		if end < 0 {
			return tool.Description[start:]
		}
		return tool.Description[start : start+end]
	}
	t.Fatal("catalog missing icuvisor_check_server_version")
	return ""
}

func facadeExtraTool(description string, toolset Toolset) Tool {
	return Tool{
		Name:         "hosted_setup_status",
		Description:  description,
		InputSchema:  map[string]any{"type": "object", "additionalProperties": false},
		OutputSchema: map[string]any{"type": "object", "additionalProperties": true},
		Requirement:  RequirementRead,
		Toolset:      toolset,
		Handler: func(context.Context, ToolRequest) (ToolResult, error) {
			return TextResult(map[string]any{"status": "ok"}), nil
		},
	}
}

func TestStreamableHTTPHandlerFacadeDefaultsToPublicFactoryError(t *testing.T) {
	t.Parallel()

	handler := NewStreamableHTTPHandler(func(*http.Request) (*Server, error) {
		return nil, errors.New("internal token refresh failed for sensitive-token")
	}, StreamableHTTPHandlerOptions{Stateless: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, StreamableHTTPPath, strings.NewReader(`{}`)))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "MCP authorization failed") {
		t.Fatalf("body = %q, want default public factory error message", body)
	}
	if strings.Contains(body, "sensitive-token") || strings.Contains(body, "token refresh") {
		t.Fatalf("body leaked internal error: %q", body)
	}
}

func newFacadeTestClient(t *testing.T) *Client {
	t.Helper()
	client, err := NewAPIKeyClient(APIKeyClientOptions{Config: facadeTestConfig(), Version: "v-public"})
	if err != nil {
		t.Fatalf("NewAPIKeyClient() error = %v", err)
	}
	return client
}

func facadeTestConfig() Config {
	return Config{APIKey: "x", AthleteID: "i12345", Timezone: "UTC", APIBaseURL: "https://intervals.icu/api/v1", HTTPTimeout: time.Second, DeleteMode: DeleteModeSafe, Toolset: ToolsetCore}
}

func hasTool(catalog []ToolInfo, name string) bool {
	for _, tool := range catalog {
		if tool.Name == name {
			return true
		}
	}
	return false
}

package icuvisor

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	internalmcp "github.com/ricardocabral/icuvisor/internal/mcp"
	internalprompts "github.com/ricardocabral/icuvisor/internal/prompts"
	internalresources "github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
	internaltools "github.com/ricardocabral/icuvisor/internal/tools"
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

func TestBearerClientFacadePassesConfiguredAthleteID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345"; got != want {
			http.Error(w, "wrong athlete", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i12345","name":"Hosted Bearer"}`))
	}))
	defer server.Close()

	client, err := NewBearerClient(BearerClientOptions{
		AccessToken: "public-bearer-token",
		AthleteID:   "i12345",
		APIBaseURL:  server.URL,
		HTTPClient:  server.Client(),
	})
	if err != nil {
		t.Fatalf("NewBearerClient() error = %v", err)
	}
	profile, err := client.inner.GetAthleteProfile(context.Background())
	if err != nil {
		t.Fatalf("GetAthleteProfile() error = %v", err)
	}
	if profile.ID != "i12345" {
		t.Fatalf("profile ID = %q, want i12345", profile.ID)
	}
}

func TestAPIKeyClientFacadeMatchesInternalBasicAuth(t *testing.T) {
	t.Parallel()

	requests := make(chan *http.Request, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i12345","name":"Fixture Athlete"}`))
	}))
	defer server.Close()

	cfg := facadeTestConfig()
	cfg.APIBaseURL = server.URL
	publicClient, err := NewAPIKeyClient(APIKeyClientOptions{Config: cfg, Version: "v-public", HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("NewAPIKeyClient() error = %v", err)
	}
	if _, err := publicClient.inner.GetAthleteProfile(context.Background()); err != nil {
		t.Fatalf("public GetAthleteProfile() error = %v", err)
	}

	internalClient, err := intervals.NewClient(intervals.Options{Config: config.Config{APIKey: cfg.APIKey, AthleteID: cfg.AthleteID, Timezone: cfg.Timezone, APIBaseURL: server.URL, HTTPTimeout: cfg.HTTPTimeout}, Version: "v-public", HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("internal NewClient() error = %v", err)
	}
	if _, err := internalClient.GetAthleteProfile(context.Background()); err != nil {
		t.Fatalf("internal GetAthleteProfile() error = %v", err)
	}

	for _, label := range []string{"public", "internal"} {
		req := <-requests
		username, password, ok := req.BasicAuth()
		if !ok || username == "" || password != "x" {
			t.Fatalf("%s BasicAuth() = %q/%q/%v, want API-key basic auth", label, username, password, ok)
		}
		if auth := req.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			t.Fatalf("%s Authorization = %q, want no bearer auth", label, auth)
		}
		if got, want := req.UserAgent(), "icuvisor/v-public"; got != want {
			t.Fatalf("%s User-Agent = %q, want %q", label, got, want)
		}
	}
}

func TestPublicCoreCatalogMatchesInternalForPolicyMatrix(t *testing.T) {
	t.Parallel()

	client := newFacadeTestClient(t)
	cfg := facadeTestConfig()
	modes := []DeleteMode{DeleteModeSafe, DeleteModeFull, DeleteModeNone}
	toolsets := []Toolset{ToolsetCompact, ToolsetCore, ToolsetFull}
	for _, mode := range modes {
		for _, toolset := range toolsets {
			t.Run(string(mode)+"/"+string(toolset), func(t *testing.T) {
				t.Parallel()

				publicRegistry := NewCoreRegistry(client, RegistryOptions{Version: "v-public", TimezoneFallback: cfg.Timezone, DebugMetadata: true, DeleteMode: mode, Toolset: toolset})
				publicCatalog, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: cfg, Registry: publicRegistry, Mode: mode, Toolset: toolset})
				if err != nil {
					t.Fatalf("public CollectToolCatalog() error = %v", err)
				}
				publicHash, err := ComputeToolCatalogHash(context.Background(), CatalogOptions{Config: cfg, Registry: publicRegistry, Mode: mode, Toolset: toolset})
				if err != nil {
					t.Fatalf("public ComputeToolCatalogHash() error = %v", err)
				}

				internalCfg, err := cfg.toInternalValidated()
				if err != nil {
					t.Fatalf("toInternalValidated() error = %v", err)
				}
				internalMode := mode.toInternal()
				internalToolset := toolset.toInternal()
				internalRegistry := internaltools.NewRegistryWithOptions(client.inner, internaltools.RegistryOptions{Version: "v-public", TimezoneFallback: cfg.Timezone, DebugMetadata: true, Capability: safety.NewCapability(internalMode), Toolset: internalToolset})
				internalCatalog, err := internalmcp.CollectToolCatalog(context.Background(), internalmcp.CatalogHashOptions{Config: internalCfg, Registry: internalRegistry, Capability: safety.NewCapability(internalMode), Toolset: internalToolset})
				if err != nil {
					t.Fatalf("internal CollectToolCatalog() error = %v", err)
				}
				internalHash, err := internalmcp.ComputeToolCatalogHash(context.Background(), internalmcp.CatalogHashOptions{Config: internalCfg, Registry: internalRegistry, Capability: safety.NewCapability(internalMode), Toolset: internalToolset})
				if err != nil {
					t.Fatalf("internal ComputeToolCatalogHash() error = %v", err)
				}

				if got, want := publicToolNames(publicCatalog), internalToolNames(internalCatalog); strings.Join(got, ",") != strings.Join(want, ",") {
					t.Fatalf("public tool names = %v, want internal %v", got, want)
				}
				if publicHash != internalHash {
					t.Fatalf("public catalog hash = %q, want internal %q", publicHash, internalHash)
				}
			})
		}
	}
}

func TestPublicResourcesAndPromptsMatchInternalDefaults(t *testing.T) {
	t.Parallel()

	client := newFacadeTestClient(t)
	cfg := facadeTestConfig()
	publicResources := &collectingResourceRegistrar{}
	if err := NewResourceRegistry(client, ResourceRegistryOptions{Version: "v-public", TimezoneFallback: cfg.Timezone, DeleteMode: DeleteModeSafe, Toolset: ToolsetCore}).inner.Register(context.Background(), publicResources); err != nil {
		t.Fatalf("public resource Register() error = %v", err)
	}
	internalResources := &collectingResourceRegistrar{}
	if err := internalresources.NewRegistryWithOptions(client.inner, internalresources.ResourceOptions{Version: "v-public", TimezoneFallback: cfg.Timezone, DeleteMode: safety.ModeSafe, Toolset: safety.ToolsetCore}).Register(context.Background(), internalResources); err != nil {
		t.Fatalf("internal resource Register() error = %v", err)
	}
	if got, want := resourceURIs(publicResources.resources), resourceURIs(internalResources.resources); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("public resource URIs = %v, want internal %v", got, want)
	}

	publicPrompts := &collectingPromptRegistrar{}
	if err := NewPromptRegistry().inner.Register(context.Background(), publicPrompts); err != nil {
		t.Fatalf("public prompt Register() error = %v", err)
	}
	internalPrompts := &collectingPromptRegistrar{}
	if err := internalprompts.NewRegistry().Register(context.Background(), internalPrompts); err != nil {
		t.Fatalf("internal prompt Register() error = %v", err)
	}
	if got, want := promptNames(publicPrompts.prompts), promptNames(internalPrompts.prompts); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("public prompt names = %v, want internal %v", got, want)
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

func TestServerFacadeAcceptsCompactToolset(t *testing.T) {
	t.Parallel()

	client := newFacadeTestClient(t)
	cfg := facadeTestConfig()
	cfg.Toolset = ToolsetCompact
	registry := NewCoreRegistry(client, RegistryOptions{Version: "v-public"})
	catalog, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: cfg, Registry: registry})
	if err != nil {
		t.Fatalf("CollectToolCatalog() error = %v", err)
	}
	if !hasTool(catalog, "get_activity_streams") {
		t.Fatal("compact catalog missing read-only full-tier stream tool")
	}
	if hasTool(catalog, "add_or_update_event") {
		t.Fatal("compact catalog contains hidden write tool add_or_update_event")
	}
	wantHash, err := ComputeToolCatalogHash(context.Background(), CatalogOptions{Config: cfg, Registry: registry})
	if err != nil {
		t.Fatalf("ComputeToolCatalogHash() error = %v", err)
	}
	server, err := NewServer(context.Background(), ServerOptions{Config: cfg, Version: "v-public", Registry: registry, ResourceRegistry: NewResourceRegistry(client, ResourceRegistryOptions{Version: "v-public"}), PromptRegistry: NewPromptRegistry(), SkipRuntimeCatalogMetadata: true})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if got := server.CatalogHash(); got != wantHash {
		t.Fatalf("server catalog hash = %q, want compact hash %q", got, wantHash)
	}
}

func TestFacadeUsesConfigDeleteModeAndToolsetDefaults(t *testing.T) {
	t.Parallel()

	cfg := facadeTestConfig()
	cfg.DeleteMode = DeleteModeFull
	cfg.Toolset = ToolsetFull
	extra := facadeExtraTool("Full default extra.", "")
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public", ExtraTools: []Tool{extra}})
	catalog, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: cfg, Registry: registry})
	if err != nil {
		t.Fatalf("CollectToolCatalog() error = %v", err)
	}
	if !hasTool(catalog, "delete_event") {
		t.Fatal("catalog missing delete_event with Config.DeleteModeFull")
	}
	if !hasTool(catalog, "hosted_setup_status") {
		t.Fatal("catalog missing full-only extra tool with Config.ToolsetFull")
	}
	checkVersion := findToolInfo(t, catalog, "icuvisor_check_server_version")
	if !strings.Contains(checkVersion.Description, "description_toolset=full") || !strings.Contains(checkVersion.Description, "description_delete_mode=full") {
		t.Fatalf("check-server-version description = %q, want Config policy defaults", checkVersion.Description)
	}

	resourceRegistry := NewResourceRegistry(newFacadeTestClient(t), ResourceRegistryOptions{Version: "v-public"})
	server, err := NewServer(context.Background(), ServerOptions{Config: cfg, Version: "v-public", Registry: registry, ResourceRegistry: resourceRegistry, DeleteMode: "", Toolset: "", SkipRuntimeCatalogMetadata: true})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if server.CatalogHash() == "" {
		t.Fatal("server catalog hash is empty")
	}

	profileResource := facadeAthleteProfileResourceWithBoundaryConfig(t, cfg)
	result, err := profileResource.Handler(context.Background(), internalresources.Request{URI: internalresources.AthleteProfileURI})
	if err != nil {
		t.Fatalf("athlete profile resource handler error = %v", err)
	}
	var shaped map[string]any
	if err := json.Unmarshal([]byte(result.Text), &shaped); err != nil {
		t.Fatalf("json.Unmarshal(resource) error = %v", err)
	}
	meta := shaped["_meta"].(map[string]any)
	if meta["delete_mode"] != "full" || meta["toolset"] != "full" {
		t.Fatalf("resource _meta = %#v, want Config policy defaults", meta)
	}
}

func TestFacadeConfigDeleteModeNoneDisablesWriteTools(t *testing.T) {
	t.Parallel()

	cfg := facadeTestConfig()
	cfg.DeleteMode = DeleteModeNone
	cfg.Toolset = ToolsetFull
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public"})
	catalog, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: cfg, Registry: registry})
	if err != nil {
		t.Fatalf("CollectToolCatalog() error = %v", err)
	}
	if hasTool(catalog, "update_wellness") {
		t.Fatal("catalog includes write tool update_wellness with Config.DeleteModeNone")
	}
}

func TestSkipRuntimeCatalogMetadataInjectsServerHashIntoHandlers(t *testing.T) {
	response.SetRuntimeCatalogMetadata("global-version", "global-hash")
	t.Cleanup(func() { response.SetRuntimeCatalogMetadata("dev", "dev-catalog-hash") })

	cfg := facadeTestConfig()
	cfg.DeleteMode = DeleteModeFull
	cfg.Toolset = ToolsetFull
	client := newHTTPFacadeTestClient(t, cfg)
	registry := NewCoreRegistry(client, RegistryOptions{Version: "v-public"})
	resourceRegistry := NewResourceRegistry(client, ResourceRegistryOptions{Version: "v-public"})
	server, err := NewServer(context.Background(), ServerOptions{Config: cfg, Version: "v-public", Registry: registry, ResourceRegistry: resourceRegistry, SkipRuntimeCatalogMetadata: true})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	internalCfg, err := cfg.toInternalValidated()
	if err != nil {
		t.Fatalf("toInternalValidated() error = %v", err)
	}
	catalog, err := internalmcp.CollectToolCatalog(context.Background(), internalmcp.CatalogHashOptions{Config: internalCfg, Registry: registryForConfigWithCatalogHash(registry, cfg, server.CatalogHash()), Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	if err != nil {
		t.Fatalf("internal CollectToolCatalog() error = %v", err)
	}
	var checkTool internaltools.Tool
	for _, tool := range catalog {
		if tool.Name == "icuvisor_check_server_version" {
			checkTool = tool
			break
		}
	}
	if checkTool.Handler == nil {
		t.Fatal("catalog missing icuvisor_check_server_version handler")
	}
	checkResult, err := checkTool.Handler(context.Background(), internaltools.Request{Name: checkTool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("check-server-version handler error = %v", err)
	}
	var checkPayload map[string]any
	if err := json.Unmarshal([]byte(checkResult.Content[0].Text), &checkPayload); err != nil {
		t.Fatalf("json.Unmarshal(check) error = %v", err)
	}
	if checkPayload["catalog_hash"] != server.CatalogHash() {
		t.Fatalf("check catalog_hash = %#v, want server hash %q", checkPayload["catalog_hash"], server.CatalogHash())
	}

	profileResource := facadeAthleteProfileResourceWithBoundaryConfigAndHash(t, cfg, server.CatalogHash())
	resourceResult, err := profileResource.Handler(context.Background(), internalresources.Request{URI: internalresources.AthleteProfileURI})
	if err != nil {
		t.Fatalf("athlete profile resource handler error = %v", err)
	}
	var shaped map[string]any
	if err := json.Unmarshal([]byte(resourceResult.Text), &shaped); err != nil {
		t.Fatalf("json.Unmarshal(resource) error = %v", err)
	}
	meta := shaped["_meta"].(map[string]any)
	if meta["catalog_hash"] != server.CatalogHash() {
		t.Fatalf("resource catalog_hash = %#v, want server hash %q", meta["catalog_hash"], server.CatalogHash())
	}
	if runtime := response.RuntimeCatalogMetadata(); runtime.Version != "global-version" || runtime.CatalogHash != "global-hash" {
		t.Fatalf("global runtime metadata = %+v, want unchanged", runtime)
	}
}

func TestInvalidPublicPolicyValuesFailClosed(t *testing.T) {
	t.Parallel()

	cfg := facadeTestConfig()
	cfg.DeleteMode = DeleteMode("nonee")
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public"})
	if _, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: cfg, Registry: registry}); err == nil {
		t.Fatal("CollectToolCatalog() error = nil, want invalid delete mode error")
	}

	cfg = facadeTestConfig()
	cfg.Toolset = Toolset("fulll")
	if _, err := NewServer(context.Background(), ServerOptions{Config: cfg, Registry: registry, SkipRuntimeCatalogMetadata: true}); err == nil {
		t.Fatal("NewServer() error = nil, want invalid toolset error")
	}

	registry = NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public", DeleteMode: DeleteMode("safee")})
	if _, err := CollectToolCatalog(context.Background(), CatalogOptions{Config: facadeTestConfig(), Registry: registry}); err == nil {
		t.Fatal("CollectToolCatalog() error = nil, want invalid registry delete mode error")
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

func TestDirectInvokerUsesMCPRegistrationPolicyAndHandlers(t *testing.T) {
	t.Parallel()

	client := newFacadeTestClient(t)
	registry := NewCoreRegistry(client, RegistryOptions{
		Version: "v-public",
		ExtraTools: []Tool{{
			Name:         "direct_echo",
			Description:  "Returns its arguments for direct-invocation coverage.",
			InputSchema:  map[string]any{"type": "object"},
			OutputSchema: map[string]any{"type": "object"},
			Requirement:  RequirementRead,
			Toolset:      ToolsetCore,
			Handler: func(_ context.Context, req ToolRequest) (ToolResult, error) {
				return TextResult(map[string]string{"name": req.Name, "arguments": string(req.Arguments)}), nil
			},
		}},
	})
	cfg := facadeTestConfig()
	invoker, err := NewInvoker(context.Background(), InvokerOptions{Config: cfg, Registry: registry, DeleteMode: DeleteModeSafe, Toolset: ToolsetCore})
	if err != nil {
		t.Fatalf("NewInvoker() error = %v", err)
	}
	if !hasTool(invoker.Tools(), "direct_echo") {
		t.Fatal("direct invoker catalog missing extra core tool")
	}
	result, err := invoker.Invoke(context.Background(), "direct_echo", json.RawMessage(`{"message":"hello"}`))
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if result.IsError {
		t.Fatal("Invoke() returned an error result")
	}
	payload, ok := result.StructuredContent.(map[string]string)
	if !ok || payload["name"] != "direct_echo" || payload["arguments"] != `{"message":"hello"}` {
		t.Fatalf("Invoke() result = %#v, want direct handler result", result.StructuredContent)
	}
	if _, err := invoker.Invoke(context.Background(), "delete_event", json.RawMessage(`{}`)); err == nil {
		t.Fatal("safe direct invoker exposed delete_event")
	}
}

func TestDirectInvokerSanitizesHandlerErrorsAndPanics(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		handler Handler
		want    string
	}{
		{
			name: "raw error",
			handler: func(context.Context, ToolRequest) (ToolResult, error) {
				return ToolResult{}, errors.New("secret upstream detail")
			},
			want: "tool failed; try again or check icuvisor logs",
		},
		{
			name: "public error",
			handler: func(context.Context, ToolRequest) (ToolResult, error) {
				return ToolResult{}, NewUserError("use a supported value", errors.New("secret cause"))
			},
			want: "use a supported value",
		},
		{
			name: "panic",
			handler: func(context.Context, ToolRequest) (ToolResult, error) {
				panic("secret panic value")
			},
			want: "tool failed; try again or check icuvisor logs",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tool := facadeExtraTool("Failure parity tool.", ToolsetCore)
			tool.Handler = tc.handler
			registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{Version: "v-public", ExtraTools: []Tool{tool}})
			invoker, err := NewInvoker(context.Background(), InvokerOptions{Config: facadeTestConfig(), Registry: registry})
			if err != nil {
				t.Fatalf("NewInvoker() error = %v", err)
			}
			_, err = invoker.Invoke(context.Background(), tool.Name, json.RawMessage(`{}`))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Invoke() error = %v, want %q", err, tc.want)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatalf("Invoke() leaked internal detail: %v", err)
			}
		})
	}
}

func TestDirectInvokerRejectsMissingRegistry(t *testing.T) {
	t.Parallel()

	if _, err := NewInvoker(context.Background(), InvokerOptions{Config: facadeTestConfig()}); err == nil || !strings.Contains(err.Error(), "missing tool registry") {
		t.Fatalf("NewInvoker() error = %v, want missing-registry error", err)
	}
}

func TestDirectInvokerRoutesConfiguredAthleteAndRejectsArgumentOverride(t *testing.T) {
	t.Parallel()

	invoker, err := NewInvoker(context.Background(), InvokerOptions{
		Config:   facadeTestConfig(),
		Registry: Registry{inner: directTargetTestRegistry{}},
	})
	if err != nil {
		t.Fatalf("NewInvoker() error = %v", err)
	}
	result, err := invoker.Invoke(context.Background(), "get_athlete_profile", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	payload, ok := result.StructuredContent.(map[string]string)
	if !ok || payload["target_athlete_id"] != "i12345" {
		t.Fatalf("Invoke() result = %#v, want configured athlete target", result.StructuredContent)
	}
	if _, err := invoker.Invoke(context.Background(), "get_athlete_profile", json.RawMessage(`{"athlete_id":"i99999"}`)); err == nil || !strings.Contains(err.Error(), "athlete_id is only supported") {
		t.Fatalf("Invoke() athlete_id error = %v, want local-mode rejection", err)
	}
}

type directTargetTestRegistry struct{}

func (directTargetTestRegistry) Register(_ context.Context, registrar internaltools.Registrar) error {
	return registrar.AddTool(internaltools.Tool{
		Name:         "get_athlete_profile",
		Description:  "Direct target routing test tool.",
		InputSchema:  map[string]any{"type": "object"},
		OutputSchema: map[string]any{"type": "object"},
		Toolset:      safety.ToolsetCore,
		Handler: func(ctx context.Context, _ internaltools.Request) (internaltools.Result, error) {
			targetAthleteID, _ := intervals.TargetAthleteIDFromContext(ctx)
			return internaltools.TextResult(map[string]string{"target_athlete_id": targetAthleteID}), nil
		},
	})
}

func newHTTPFacadeTestClient(t *testing.T, cfg Config) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i12345","name":"Resource Athlete","timezone":"UTC","sportSettings":[]}`))
	}))
	t.Cleanup(server.Close)
	clientCfg := cfg
	clientCfg.APIBaseURL = server.URL
	client, err := NewAPIKeyClient(APIKeyClientOptions{Config: clientCfg, Version: "v-public", HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("NewAPIKeyClient() error = %v", err)
	}
	return client
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

func findToolInfo(t *testing.T, catalog []ToolInfo, name string) ToolInfo {
	t.Helper()
	for _, tool := range catalog {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("catalog missing %s", name)
	return ToolInfo{}
}

func facadeAthleteProfileResourceWithBoundaryConfig(t *testing.T, cfg Config) internalresources.Resource {
	t.Helper()
	return facadeAthleteProfileResourceWithBoundaryConfigAndHash(t, cfg, "")
}

func facadeAthleteProfileResourceWithBoundaryConfigAndHash(t *testing.T, cfg Config, catalogHash string) internalresources.Resource {
	t.Helper()
	client := newHTTPFacadeTestClient(t, cfg)
	registry := NewResourceRegistry(client, ResourceRegistryOptions{Version: "v-public"})
	inner := resourceRegistryForConfigWithCatalogHash(registry, cfg, catalogHash)
	registrar := &collectingResourceRegistrar{}
	if err := inner.Register(context.Background(), registrar); err != nil {
		t.Fatalf("resource registry Register() error = %v", err)
	}
	for _, resource := range registrar.resources {
		if resource.URI == internalresources.AthleteProfileURI {
			return resource
		}
	}
	t.Fatal("resource registry missing athlete profile")
	return internalresources.Resource{}
}

type collectingResourceRegistrar struct {
	resources []internalresources.Resource
}

func (r *collectingResourceRegistrar) AddResource(resource internalresources.Resource) error {
	r.resources = append(r.resources, resource)
	return nil
}

func hasTool(catalog []ToolInfo, name string) bool {
	for _, tool := range catalog {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func publicToolNames(catalog []ToolInfo) []string {
	names := make([]string, 0, len(catalog))
	for _, tool := range catalog {
		names = append(names, tool.Name)
	}
	sort.Strings(names)
	return names
}

func internalToolNames(catalog []internaltools.Tool) []string {
	names := make([]string, 0, len(catalog))
	for _, tool := range catalog {
		names = append(names, tool.Name)
	}
	sort.Strings(names)
	return names
}

func resourceURIs(resources []internalresources.Resource) []string {
	uris := make([]string, 0, len(resources))
	for _, resource := range resources {
		uris = append(uris, resource.URI)
	}
	sort.Strings(uris)
	return uris
}

type collectingPromptRegistrar struct {
	prompts []internalprompts.Prompt
}

func (r *collectingPromptRegistrar) AddPrompt(prompt internalprompts.Prompt) error {
	r.prompts = append(r.prompts, prompt)
	return nil
}

func promptNames(prompts []internalprompts.Prompt) []string {
	names := make([]string, 0, len(prompts))
	for _, prompt := range prompts {
		names = append(names, prompt.Name)
	}
	sort.Strings(names)
	return names
}

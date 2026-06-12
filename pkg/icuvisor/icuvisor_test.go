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

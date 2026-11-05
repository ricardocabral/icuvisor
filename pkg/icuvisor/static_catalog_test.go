package icuvisor

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestStaticCatalogHTTP(t *testing.T) {
	for _, static := range []bool{false, true} {
		name := "default"
		if static {
			name = "static"
		}
		t.Run(name, func(t *testing.T) {
			cfg := facadeTestConfig()
			client := newFacadeTestClient(t)
			registry := NewCoreRegistry(client, RegistryOptions{ExtraTools: []Tool{facadeExtraTool("Report setup status.", ToolsetCore)}})
			handler := NewStreamableHTTPHandler(func(r *http.Request) (*Server, error) {
				return NewServer(r.Context(), ServerOptions{
					Config: cfg, Registry: registry,
					ResourceRegistry: NewResourceRegistry(client, ResourceRegistryOptions{DisableAthleteProfile: true}),
					PromptRegistry:   NewPromptRegistry(), SkipRuntimeCatalogMetadata: true,
					StaticCatalog: static,
				})
			}, StreamableHTTPHandlerOptions{Stateless: true, JSONResponse: true})
			httpServer := httptest.NewServer(handler)
			defer httpServer.Close()
			post := func(payload string) []byte {
				t.Helper()
				return catalogHTTPPost(t, httpServer, payload)
			}
			for _, payload := range []string{
				`{"jsonrpc":"2.0","id":1,"method":"server/discover"}`,
				`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","clientInfo":{"name":"catalog-test","version":"test"},"capabilities":{}}}`,
			} {
				body := catalogHTTPResult(t, post(payload))
				var result struct {
					Capabilities sdkmcp.ServerCapabilities `json:"capabilities"`
				}
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatal(err)
				}
				caps := result.Capabilities
				if caps.Tools == nil || caps.Prompts == nil || caps.Resources == nil {
					t.Fatalf("missing configured catalog capabilities: %+v", caps)
				}
				if caps.Tools.ListChanged != !static || caps.Prompts.ListChanged != !static || caps.Resources.ListChanged != !static {
					t.Fatalf("listChanged = %v/%v/%v, want %v", caps.Tools.ListChanged, caps.Prompts.ListChanged, caps.Resources.ListChanged, !static)
				}
				if caps.Resources.Subscribe {
					t.Fatal("resource subscriptions unexpectedly advertised")
				}
				var wire struct {
					Capabilities map[string]json.RawMessage `json:"capabilities"`
				}
				if err := json.Unmarshal(body, &wire); err != nil {
					t.Fatal(err)
				}
				if _, ok := wire.Capabilities["logging"]; !ok {
					t.Fatal("default logging capability was lost")
				}
			}

			tools := catalogHTTPResult(t, post(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`))
			if !strings.Contains(string(tools), `"name":"hosted_setup_status"`) {
				t.Fatalf("tools/list missing extra tool: %s", tools)
			}
			called := catalogHTTPResult(t, post(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"hosted_setup_status","arguments":{}}}`))
			if !strings.Contains(string(called), `"status":"ok"`) {
				t.Fatalf("tools/call missing success: %s", called)
			}
			var prompts struct {
				Prompts []struct {
					Name string `json:"name"`
				} `json:"prompts"`
			}
			if err := json.Unmarshal(catalogHTTPResult(t, post(`{"jsonrpc":"2.0","id":4,"method":"prompts/list"}`)), &prompts); err != nil || len(prompts.Prompts) == 0 {
				t.Fatalf("prompts/list = %+v, error %v", prompts, err)
			}
			promptParams, err := json.Marshal(map[string]any{"name": prompts.Prompts[0].Name, "arguments": map[string]string{}})
			if err != nil {
				t.Fatal(err)
			}
			catalogHTTPResult(t, post(`{"jsonrpc":"2.0","id":5,"method":"prompts/get","params":`+string(promptParams)+`}`))
			var resources struct {
				Resources []struct {
					URI string `json:"uri"`
				} `json:"resources"`
			}
			if err := json.Unmarshal(catalogHTTPResult(t, post(`{"jsonrpc":"2.0","id":6,"method":"resources/list"}`)), &resources); err != nil || len(resources.Resources) == 0 {
				t.Fatalf("resources/list = %+v, error %v", resources, err)
			}
			resourceParams, err := json.Marshal(map[string]string{"uri": resources.Resources[0].URI})
			if err != nil {
				t.Fatal(err)
			}
			catalogHTTPResult(t, post(`{"jsonrpc":"2.0","id":7,"method":"resources/read","params":`+string(resourceParams)+`}`))

			if static {
				for _, notifications := range []string{`{}`, `{"toolsListChanged":true}`, `{"promptsListChanged":true,"resourcesListChanged":true}`, `{"toolsListChanged":true,"promptsListChanged":true,"resourcesListChanged":true,"resourceSubscriptions":["icuvisor://unsupported-sensitive-uri"]}`} {
					body := post(`{"jsonrpc":"2.0","id":"syntheti","method":"subscriptions/listen","params":{"notifications":` + notifications + `}}`)
					if !strings.Contains(string(body), `"method":"notifications/subscriptions/acknowledged"`) || !strings.Contains(string(body), `"notifications":{}`) || !strings.Contains(string(body), `"result":`) {
						t.Fatalf("subscription did not acknowledge empty notifications and finish: %s", body)
					}
					if strings.Contains(string(body), "unsupported-sensitive-uri") {
						t.Fatalf("subscription echoed unsupported URI: %s", body)
					}
				}
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
				defer cancel()
				request := catalogHTTPRequest(t, ctx, httpServer, `{"jsonrpc":"2.0","id":"syntheti","method":"subscriptions/listen","params":{"notifications":{"toolsListChanged":true,"promptsListChanged":true,"resourcesListChanged":true}}}`)
				response, err := httpServer.Client().Do(request)
				if err != nil {
					t.Fatal(err)
				}
				defer response.Body.Close()
				body, err := io.ReadAll(response.Body)
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("default subscription ended before cancellation: error %v, body %s", err, body)
				}
				if response.StatusCode != http.StatusOK || !strings.Contains(string(body), `"toolsListChanged":true`) || !strings.Contains(string(body), `"promptsListChanged":true`) || !strings.Contains(string(body), `"resourcesListChanged":true`) || strings.Contains(string(body), `"result":`) {
					t.Fatalf("default subscription did not remain active after acknowledgement: status %d, body %s", response.StatusCode, body)
				}
			}
		})
	}
}

func TestStaticCatalogOmitsUnconfiguredCapabilities(t *testing.T) {
	handler := NewStreamableHTTPHandler(func(r *http.Request) (*Server, error) {
		return NewServer(r.Context(), ServerOptions{Config: facadeTestConfig(), StaticCatalog: true, SkipRuntimeCatalogMetadata: true})
	}, StreamableHTTPHandlerOptions{Stateless: true, JSONResponse: true})
	server := httptest.NewServer(handler)
	defer server.Close()
	var result struct {
		Capabilities map[string]json.RawMessage `json:"capabilities"`
	}
	body := catalogHTTPResult(t, catalogHTTPPost(t, server, `{"jsonrpc":"2.0","id":1,"method":"server/discover"}`))
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tools", "prompts", "resources"} {
		if _, ok := result.Capabilities[name]; ok {
			t.Fatalf("unconfigured %s capability advertised: %s", name, body)
		}
	}
}

func TestStaticCatalogSDKClientDoesNotReconnect(t *testing.T) {
	for _, stale := range []bool{false, true} {
		name := "new_discovery"
		if stale {
			name = "stale_capabilities"
		}
		t.Run(name, func(t *testing.T) {
			registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{})
			handler := NewStreamableHTTPHandler(func(r *http.Request) (*Server, error) {
				return NewServer(r.Context(), ServerOptions{Config: facadeTestConfig(), Registry: registry, StaticCatalog: true, SkipRuntimeCatalogMetadata: true})
			}, StreamableHTTPHandlerOptions{Stateless: true, JSONResponse: true})
			var listens, gets atomic.Int64
			finished := make(chan struct{}, 10)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					gets.Add(1)
				}
				switch r.Header.Get("MCP-Method") {
				case "subscriptions/listen":
					listens.Add(1)
					handler.ServeHTTP(w, r)
					select {
					case finished <- struct{}{}:
					default:
					}
				case "server/discover":
					if !stale {
						handler.ServeHTTP(w, r)
						return
					}
					// Simulate discovery cached before a rollout to a static server.
					recorder := httptest.NewRecorder()
					handler.ServeHTTP(recorder, r)
					var envelope struct {
						JSONRPC string                `json:"jsonrpc"`
						ID      json.RawMessage       `json:"id"`
						Result  sdkmcp.DiscoverResult `json:"result"`
					}
					if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
						t.Errorf("decoding discovery: %v", err)
						return
					}
					envelope.Result.Capabilities.Tools = &sdkmcp.ToolCapabilities{ListChanged: true}
					envelope.Result.Capabilities.Prompts = &sdkmcp.PromptCapabilities{ListChanged: true}
					envelope.Result.Capabilities.Resources = &sdkmcp.ResourceCapabilities{ListChanged: true}
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(envelope); err != nil {
						t.Errorf("encoding stale discovery: %v", err)
					}
				default:
					handler.ServeHTTP(w, r)
				}
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "static-catalog-test", Version: "test"}, &sdkmcp.ClientOptions{
				ToolListChangedHandler:     func(context.Context, *sdkmcp.ToolListChangedRequest) {},
				PromptListChangedHandler:   func(context.Context, *sdkmcp.PromptListChangedRequest) {},
				ResourceListChangedHandler: func(context.Context, *sdkmcp.ResourceListChangedRequest) {},
			})
			session, err := client.Connect(ctx, &sdkmcp.StreamableClientTransport{Endpoint: server.URL + StreamableHTTPPath, HTTPClient: server.Client()}, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			wantListens := int64(0)
			if stale {
				wantListens = 1
				select {
				case <-finished:
				case <-ctx.Done():
					t.Fatal("stale subscription did not finish")
				}
				// Observe beyond the SDK's one-second initial SSE reconnect delay.
				select {
				case <-time.After(1100 * time.Millisecond):
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
			}
			if _, err := session.ListTools(ctx, nil); err != nil {
				t.Fatalf("session unusable after subscription completion: %v", err)
			}
			if got := listens.Load(); got != wantListens || gets.Load() != 0 {
				t.Fatalf("subscription POSTs = %d, want %d; reconnect GETs = %d, want 0", got, wantListens, gets.Load())
			}
		})
	}
}

func catalogHTTPPost(t *testing.T, server *httptest.Server, payload string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	response, err := server.Client().Do(catalogHTTPRequest(t, ctx, server, payload))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("HTTP response did not finish: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("HTTP status = %d, body %s", response.StatusCode, body)
	}
	return body
}

func catalogHTTPRequest(t *testing.T, ctx context.Context, server *httptest.Server, payload string) *http.Request {
	t.Helper()
	var requestBody map[string]any
	if err := json.Unmarshal([]byte(payload), &requestBody); err != nil {
		t.Fatal(err)
	}
	protocolVersion := "2026-07-28"
	if requestBody["method"] == "initialize" {
		protocolVersion = "2025-11-25"
	}
	params, _ := requestBody["params"].(map[string]any)
	if params == nil {
		params = make(map[string]any)
		requestBody["params"] = params
	}
	params["_meta"] = map[string]any{
		"io.modelcontextprotocol/protocolVersion":    protocolVersion,
		"io.modelcontextprotocol/clientInfo":         map[string]string{"name": "catalog-test", "version": "test"},
		"io.modelcontextprotocol/clientCapabilities": map[string]any{},
	}
	encoded, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}
	payload = string(encoded)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+StreamableHTTPPath, strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("MCP-Method", requestBody["method"].(string))
	if name, ok := params["name"].(string); ok {
		req.Header.Set("MCP-Name", name)
	}
	if uri, ok := params["uri"].(string); ok {
		req.Header.Set("MCP-Name", uri)
	}
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", protocolVersion)
	return req
}

func catalogHTTPResult(t *testing.T, body []byte) json.RawMessage {
	t.Helper()
	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("invalid JSON-RPC response: %v: %s", err, body)
	}
	if len(envelope.Error) > 0 || len(envelope.Result) == 0 {
		t.Fatalf("expected JSON-RPC result: %s", body)
	}
	return envelope.Result
}

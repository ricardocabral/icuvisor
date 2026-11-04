package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
)

func TestHTTPTransportConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		fields       string
		dotenv       string
		env          map[string]string
		flags        []string
		stateless    bool
		jsonResponse bool
	}{
		{name: "defaults preserve sessions and SSE"},
		{name: "JSON enables both", fields: `,"http_stateless":true,"http_json_response":true`, stateless: true, jsonResponse: true},
		{name: "coach with JSON responses", fields: `,"http_json_response":true,"coach":{"athletes":[{"id":"i123","allowed_tools":["*"]}]}`, env: map[string]string{config.EnvCoachMode: "on"}, jsonResponse: true},
		{name: "JSON response only", fields: `,"http_json_response":true`, jsonResponse: true},
		{name: "stateless only", fields: `,"http_stateless":true`, stateless: true},
		{name: "dotenv enables both", dotenv: "ICUVISOR_HTTP_STATELESS=true\nICUVISOR_HTTP_JSON_RESPONSE=true\n", stateless: true, jsonResponse: true},
		{name: "JSON false overrides dotenv true", fields: `,"http_stateless":false,"http_json_response":false`, dotenv: "ICUVISOR_HTTP_STATELESS=true\nICUVISOR_HTTP_JSON_RESPONSE=true\n"},
		{name: "env false overrides JSON true", fields: `,"http_stateless":true,"http_json_response":true`, env: map[string]string{"ICUVISOR_HTTP_STATELESS": "false", "ICUVISOR_HTTP_JSON_RESPONSE": "false"}},
		{name: "env enables both", env: map[string]string{"ICUVISOR_HTTP_STATELESS": "true", "ICUVISOR_HTTP_JSON_RESPONSE": "true"}, stateless: true, jsonResponse: true},
		{name: "CLI false overrides env true", env: map[string]string{"ICUVISOR_HTTP_STATELESS": "true", "ICUVISOR_HTTP_JSON_RESPONSE": "true"}, flags: []string{"--http-stateless=false", "--http-json-response", "false"}},
		{name: "CLI true overrides env false", env: map[string]string{"ICUVISOR_HTTP_STATELESS": "false", "ICUVISOR_HTTP_JSON_RESPONSE": "false"}, flags: []string{"--http-stateless", "true", "--http-json-response=true"}, stateless: true, jsonResponse: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.json")
			if err := os.WriteFile(path, []byte(`{"athlete_id":"i123","transport":"http"`+tc.fields+`}`), 0600); err != nil {
				t.Fatal(err)
			}
			dotenv := filepath.Join(dir, ".env")
			if err := os.WriteFile(dotenv, []byte(tc.dotenv), 0600); err != nil {
				t.Fatal(err)
			}
			env := map[string]string{config.EnvAPIKey: "dummy-test-key"}
			for key, value := range tc.env {
				env[key] = value
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			ready := make(chan struct{})
			done := make(chan error, 1)
			go func() {
				done <- Run(ctx, Options{
					Args: append([]string{"--config", path, "--env-file", dotenv}, tc.flags...),
					LoadConfig: func(ctx context.Context, opts config.Options) (config.Config, error) {
						opts.Env = env
						return config.Load(ctx, opts)
					},
					StartServer: func(ctx context.Context, info ServerInfo) error {
						if info.Config.CoachModeEnabled() {
							server, cleanup, err := wireServer(ctx, info, deps{})
							if err != nil {
								return err
							}
							defer cleanup()
							close(ready)
							return server.ServeStreamableHTTP(ctx, listener)
						}
						server, err := wireCoreServer(ctx, info)
						if err != nil {
							return err
						}
						close(ready)
						return server.ServeStreamableHTTP(ctx, listener)
					},
				})
			}()
			select {
			case err := <-done:
				t.Fatalf("start server: %v", err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-ready:
			}
			defer func() {
				cancel()
				if err := <-done; !errors.Is(err, context.Canceled) {
					t.Errorf("stop server: %v", err)
				}
			}()
			endpoint := "http://" + listener.Addr().String() + "/mcp"
			if tc.stateless {
				httpConfigurationRequest(t, ctx, endpoint, `{"jsonrpc":"2.0","id":0,"method":"tools/list","params":{}}`, "")
			}
			headers, result := httpConfigurationRequest(t, ctx, endpoint, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"config-test","version":"1"}}}`, "")
			if result["serverInfo"] == nil {
				t.Fatalf("initialize result: %s", result)
			}
			sessionID := headers.Get("Mcp-Session-Id")
			if (sessionID == "") != tc.stateless {
				t.Fatalf("session ID = %q, stateless = %t", sessionID, tc.stateless)
			}
			wantContentType := "text/event-stream"
			if tc.jsonResponse {
				wantContentType = "application/json"
			}
			if got := headers.Get("Content-Type"); !strings.HasPrefix(got, wantContentType) {
				t.Fatalf("Content-Type = %q, want %q", got, wantContentType)
			}
			_, listed := httpConfigurationRequest(t, ctx, endpoint, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`, sessionID)
			var tools []struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(listed["tools"], &tools); err != nil || len(tools) == 0 {
				t.Fatalf("tools/list: %s, error %v", listed, err)
			}
			if tc.stateless {
				// Stateless handlers must ignore stale client session headers.
				httpConfigurationRequest(t, ctx, endpoint, `{"jsonrpc":"2.0","id":3,"method":"ping"}`, "stale-session")
				for _, method := range []string{http.MethodGet, http.MethodDelete} {
					req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
					if err != nil {
						t.Fatal(err)
					}
					req.Header.Set("Accept", "text/event-stream")
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						t.Fatal(err)
					}
					resp.Body.Close()
					if resp.StatusCode != http.StatusMethodNotAllowed {
						t.Errorf("%s status = %d, want 405", method, resp.StatusCode)
					}
				}
			}
		})
	}
}

func httpConfigurationRequest(t *testing.T, ctx context.Context, endpoint, body, sessionID string) (http.Header, map[string]json.RawMessage) {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Mcp-Protocol-Version", "2025-11-25")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST status = %d: %s", resp.StatusCode, data)
	}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "data: ") {
				data = []byte(strings.TrimPrefix(line, "data: "))
				break
			}
		}
	}
	var envelope struct {
		JSONRPC string                     `json:"jsonrpc"`
		Result  map[string]json.RawMessage `json:"result"`
		Error   json.RawMessage            `json:"error"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.JSONRPC != "2.0" || envelope.Error != nil || envelope.Result == nil {
		t.Fatalf("invalid response: %s", data)
	}
	return resp.Header, envelope.Result
}

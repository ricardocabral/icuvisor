package config

import (
	"context"
	"strings"
	"testing"
)

func TestLoadTransportAndHTTPBindSelection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	dotEnvPath := dir + "/.env"
	writeFile(t, configPath, `{
		"api_key": "json-key",
		"athlete_id": "i111",
		"transport": "http",
		"http_bind": "127.0.0.1:9000"
	}`)
	writeFile(t, dotEnvPath, strings.Join([]string{
		"ICUVISOR_TRANSPORT=stdio",
		"ICUVISOR_HTTP_BIND=127.0.0.1:9001",
	}, "\n"))

	cfg, err := Load(context.Background(), Options{Path: configPath, DotEnvPath: dotEnvPath, Env: map[string]string{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Transport != TransportHTTP || cfg.HTTPBindAddress != "127.0.0.1:9000" {
		t.Fatalf("JSON transport/bind = %q %q, want http 127.0.0.1:9000", cfg.Transport, cfg.HTTPBindAddress)
	}

	cfg, err = Load(context.Background(), Options{
		Path:       configPath,
		DotEnvPath: dotEnvPath,
		Env: map[string]string{
			EnvTransport: "stdio",
			EnvHTTPBind:  "127.0.0.1:9002",
		},
	})
	if err != nil {
		t.Fatalf("Load() with env override error = %v", err)
	}
	if cfg.Transport != TransportStdio || cfg.HTTPBindAddress != "127.0.0.1:9002" {
		t.Fatalf("env transport/bind = %q %q, want stdio 127.0.0.1:9002", cfg.Transport, cfg.HTTPBindAddress)
	}

	cfg, err = Load(context.Background(), Options{
		Path:            configPath,
		DotEnvPath:      dotEnvPath,
		Env:             map[string]string{EnvTransport: "stdio", EnvHTTPBind: "127.0.0.1:9002"},
		Transport:       "http",
		HTTPBindAddress: "192.168.1.20:9003",
	})
	if err != nil {
		t.Fatalf("Load() with CLI override error = %v", err)
	}
	if cfg.Transport != TransportHTTP || cfg.HTTPBindAddress != "192.168.1.20:9003" {
		t.Fatalf("CLI transport/bind = %q %q, want http 192.168.1.20:9003", cfg.Transport, cfg.HTTPBindAddress)
	}
	if HTTPBindAddressIsLoopback(cfg.HTTPBindAddress) {
		t.Fatalf("HTTPBindAddressIsLoopback(%q) = true, want false", cfg.HTTPBindAddress)
	}

	cfg, err = Load(context.Background(), Options{
		DotEnvPath: dir + "/missing.env",
		Env: map[string]string{
			EnvAPIKey:    "env-key",
			EnvAthleteID: "i333",
			EnvTransport: "http",
		},
	})
	if err != nil {
		t.Fatalf("Load() HTTP default bind error = %v", err)
	}
	if cfg.Transport != TransportHTTP {
		t.Fatalf("Transport = %q, want http", cfg.Transport)
	}
	if cfg.HTTPBindAddress != DefaultHTTPBindAddress {
		t.Fatalf("HTTPBindAddress = %q, want default %q", cfg.HTTPBindAddress, DefaultHTTPBindAddress)
	}
	if !HTTPBindAddressIsLoopback(cfg.HTTPBindAddress) {
		t.Fatalf("HTTP-mode default bind %q is not loopback", cfg.HTTPBindAddress)
	}
}

func TestValidateHTTPBindAddress(t *testing.T) {
	t.Parallel()

	valid := []string{"127.0.0.1:8765", "192.168.1.20:8765", "[::1]:8765", "127.0.0.1 : 8765"}
	for _, value := range valid {
		if err := ValidateHTTPBindAddress(value); err != nil {
			t.Fatalf("ValidateHTTPBindAddress(%q) error = %v", value, err)
		}
	}
	normalized, err := NormalizeHTTPBindAddress("127.0.0.1 : 8765")
	if err != nil {
		t.Fatalf("NormalizeHTTPBindAddress() error = %v", err)
	}
	if normalized != "127.0.0.1:8765" {
		t.Fatalf("NormalizeHTTPBindAddress() = %q, want 127.0.0.1:8765", normalized)
	}
	if !HTTPBindAddressIsLoopback("127.0.0.1:8765") {
		t.Fatal("127.0.0.1:8765 should be loopback")
	}
	if !HTTPBindAddressIsLoopback("[::1]:8765") {
		t.Fatal("[::1]:8765 should be loopback")
	}

	invalid := []string{"", ":8765", "127.0.0.1", "127.0.0.1:", "127.0.0.1:http", "127.0.0.1:0", "127.0.0.1:65536", "http://127.0.0.1:8765", "localhost:8765"}
	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			if err := ValidateHTTPBindAddress(value); err == nil {
				t.Fatalf("ValidateHTTPBindAddress(%q) error = nil, want error", value)
			}
		})
	}
}

func TestHTTPOptionsValidation(t *testing.T) {
	tests := []struct {
		name   string
		fields string
		env    map[string]string
		want   string
	}{
		{name: "invalid stateless", env: map[string]string{"ICUVISOR_HTTP_STATELESS": "yes"}, want: "http_stateless"},
		{name: "invalid JSON response", env: map[string]string{"ICUVISOR_HTTP_JSON_RESPONSE": "yes"}, want: "http_json_response"},
		{name: "JSON requires boolean", fields: `,"http_stateless":"true"`, want: "invalid config JSON"},
		{name: "coach on rejects stateless", fields: `,"http_stateless":true,"coach":{"athletes":[{"id":"i123","allowed_tools":["*"]}]}`, env: map[string]string{EnvCoachMode: "on"}, want: "coach mode requires stateful HTTP"},
		{name: "coach auto rejects stateless", fields: `,"http_stateless":true,"coach":{"athletes":[{"id":"i123","allowed_tools":["*"]}]}`, env: map[string]string{EnvCoachMode: "auto"}, want: "coach mode requires stateful HTTP"},
		{name: "stdio coach ignores stateless HTTP", fields: `,"http_stateless":true,"coach":{"athletes":[{"id":"i123","allowed_tools":["*"]}]}`, env: map[string]string{EnvCoachMode: "on", EnvTransport: "stdio"}},
		{name: "coach auto without roster allows stateless", fields: `,"http_stateless":true`, env: map[string]string{EnvCoachMode: "auto"}},
		{name: "coach off allows stateless", fields: `,"http_stateless":true,"coach":{"athletes":[{"id":"i123","allowed_tools":["*"]}]}`, env: map[string]string{EnvCoachMode: "off"}},
		{name: "coach allows JSON responses", fields: `,"http_json_response":true,"coach":{"athletes":[{"id":"i123","allowed_tools":["*"]}]}`, env: map[string]string{EnvCoachMode: "on"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := dir + "/config.json"
			writeFile(t, path, `{"athlete_id":"i123","transport":"http"`+tc.fields+`}`)
			env := map[string]string{EnvAPIKey: "test-key"}
			for k, v := range tc.env {
				env[k] = v
			}
			_, err := Load(context.Background(), Options{Path: path, DotEnvPath: dir + "/absent.env", Env: env})
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Load error = %v, want %q", err, tc.want)
			}
		})
	}
}

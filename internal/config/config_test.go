package config

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestNormalizeAthleteIDForDisplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "numeric", input: "12345", want: "i12345"},
		{name: "prefixed", input: "i12345", want: "i12345"},
		{name: "invalid", input: " athlete ", want: "athlete"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeAthleteIDForDisplay(tc.input); got != tc.want {
				t.Fatalf("NormalizeAthleteIDForDisplay(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeAthleteID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "digits", input: "12345", want: "i12345"},
		{name: "prefixed", input: "i12345", want: "i12345"},
		{name: "uppercase prefix", input: "I12345", want: "i12345"},
		{name: "trim spaces", input: " 12345 ", want: "i12345"},
		{name: "empty", input: "", wantErr: true},
		{name: "prefix only", input: "i", wantErr: true},
		{name: "letters", input: "abc", wantErr: true},
		{name: "mixed", input: "i12x", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeAthleteID(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("NormalizeAthleteID() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeAthleteID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeAthleteID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoadPrecedenceAndDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	dotEnvPath := dir + "/.env"
	writeFile(t, configPath, `{
		"api_key": "json-key",
		"athlete_id": "111",
		"timezone": "America/Sao_Paulo",
		"api_base_url": "https://json.example.test/api",
		"http_timeout": "10s"
	}`)
	writeFile(t, dotEnvPath, strings.Join([]string{
		"INTERVALS_ICU_API_KEY=dotenv-key",
		"INTERVALS_ICU_ATHLETE_ID=222",
		"ICUVISOR_TIMEZONE=Europe/Lisbon",
		"ICUVISOR_API_BASE_URL=https://dotenv.example.test/api",
		"ICUVISOR_HTTP_TIMEOUT=20s",
		"ICUVISOR_TOOLSET=full",
		"IGNORED=value",
	}, "\n"))

	cfg, err := Load(context.Background(), Options{
		Path:       configPath,
		DotEnvPath: dotEnvPath,
		Env: map[string]string{
			EnvAPIKey:            "env-key",
			EnvAthleteID:         "333",
			EnvHTTPTimeout:       "45s",
			safety.EnvToolset:    "core",
			safety.EnvDeleteMode: "full",
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "env-key" {
		t.Fatalf("APIKey = %q, want env-key", cfg.APIKey)
	}
	if cfg.AthleteID != "i333" {
		t.Fatalf("AthleteID = %q, want i333", cfg.AthleteID)
	}
	if cfg.Timezone != "America/Sao_Paulo" {
		t.Fatalf("Timezone = %q, want America/Sao_Paulo", cfg.Timezone)
	}
	if cfg.APIBaseURL != "https://json.example.test/api" {
		t.Fatalf("APIBaseURL = %q, want JSON value", cfg.APIBaseURL)
	}
	if cfg.HTTPTimeout != 45*time.Second {
		t.Fatalf("HTTPTimeout = %s, want 45s", cfg.HTTPTimeout)
	}
	if cfg.Toolset != safety.ToolsetCore {
		t.Fatalf("Toolset = %q, want core", cfg.Toolset)
	}
	if cfg.DeleteMode != safety.ModeFull {
		t.Fatalf("DeleteMode = %q, want full", cfg.DeleteMode)
	}
	if cfg.Transport != TransportStdio {
		t.Fatalf("Transport = %q, want stdio", cfg.Transport)
	}
	if cfg.HTTPBindAddress != DefaultHTTPBindAddress {
		t.Fatalf("HTTPBindAddress = %q, want %q", cfg.HTTPBindAddress, DefaultHTTPBindAddress)
	}
}

func TestLoadDotEnvFillsAbsentValues(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dotEnvPath := dir + "/.env"
	writeFile(t, dotEnvPath, strings.Join([]string{
		"INTERVALS_ICU_API_KEY=dotenv-key",
		"INTERVALS_ICU_ATHLETE_ID=i444",
	}, "\n"))

	cfg, err := Load(context.Background(), Options{DotEnvPath: dotEnvPath, Env: map[string]string{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "dotenv-key" || cfg.AthleteID != "i444" {
		t.Fatalf("Load() = api key %q athlete %q, want .env values", cfg.APIKey, cfg.AthleteID)
	}
	if cfg.Timezone != DefaultTimezone {
		t.Fatalf("Timezone = %q, want %q", cfg.Timezone, DefaultTimezone)
	}
	if cfg.APIBaseURL != DefaultAPIBaseURL {
		t.Fatalf("APIBaseURL = %q, want %q", cfg.APIBaseURL, DefaultAPIBaseURL)
	}
	if cfg.HTTPTimeout != DefaultHTTPTimeout {
		t.Fatalf("HTTPTimeout = %s, want %s", cfg.HTTPTimeout, DefaultHTTPTimeout)
	}
	if cfg.Toolset != safety.ToolsetCore {
		t.Fatalf("Toolset = %q, want default core", cfg.Toolset)
	}
	if cfg.Transport != TransportStdio {
		t.Fatalf("Transport = %q, want default stdio", cfg.Transport)
	}
	if cfg.HTTPBindAddress != DefaultHTTPBindAddress {
		t.Fatalf("HTTPBindAddress = %q, want %q", cfg.HTTPBindAddress, DefaultHTTPBindAddress)
	}
	if !HTTPBindAddressIsLoopback(cfg.HTTPBindAddress) {
		t.Fatalf("default HTTP bind %q is not loopback", cfg.HTTPBindAddress)
	}
}

func TestLoadToolsetFromDotEnvAndInvalidEnvFallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dotEnvPath := dir + "/.env"
	writeFile(t, dotEnvPath, strings.Join([]string{
		"INTERVALS_ICU_API_KEY=dotenv-key",
		"INTERVALS_ICU_ATHLETE_ID=i444",
		"ICUVISOR_TOOLSET=full",
	}, "\n"))

	cfg, err := Load(context.Background(), Options{DotEnvPath: dotEnvPath, Env: map[string]string{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Toolset != safety.ToolsetFull {
		t.Fatalf("Toolset = %q, want full from .env", cfg.Toolset)
	}

	cfg, err = Load(context.Background(), Options{DotEnvPath: dotEnvPath, Env: map[string]string{safety.EnvToolset: "unexpected"}})
	if err != nil {
		t.Fatalf("Load() with invalid env toolset error = %v", err)
	}
	if cfg.Toolset != safety.ToolsetCore {
		t.Fatalf("Toolset = %q, want invalid env fallback core", cfg.Toolset)
	}
}

func TestLoadUsesConfigPathFromEnv(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	writeFile(t, configPath, `{"api_key":"json-key","athlete_id":"555"}`)

	cfg, err := Load(context.Background(), Options{
		DotEnvPath: dir + "/missing.env",
		Env: map[string]string{
			EnvConfigPath: configPath,
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "json-key" || cfg.AthleteID != "i555" {
		t.Fatalf("Load() = api key %q athlete %q, want JSON values", cfg.APIKey, cfg.AthleteID)
	}
}

func TestLoadConfigFileErrorsAreActionableAndRedacted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	testCredential := strings.Repeat("x", 12)

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		_, err := Load(context.Background(), Options{
			Path:       dir + "/missing.json",
			DotEnvPath: dir + "/missing.env",
			Env: map[string]string{
				EnvAPIKey:    testCredential,
				EnvAthleteID: "123",
			},
		})
		if err == nil {
			t.Fatal("Load() error = nil, want error")
		}
		msg := err.Error()
		for _, want := range []string{"config file", "not found", "--config", EnvConfigPath} {
			if !strings.Contains(msg, want) {
				t.Fatalf("error %q does not contain %q", msg, want)
			}
		}
		if strings.Contains(msg, testCredential) {
			t.Fatalf("error leaked API key: %q", msg)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		path := dir + "/invalid.json"
		writeFile(t, path, `{"api_key":"`+testCredential+`","athlete_id":"123","extra":true}`)

		_, err := Load(context.Background(), Options{Path: path, DotEnvPath: dir + "/missing.env", Env: map[string]string{}})
		if err == nil {
			t.Fatal("Load() error = nil, want error")
		}
		msg := err.Error()
		for _, want := range []string{"invalid config JSON", "expected fields", "api_key", "athlete_id"} {
			if !strings.Contains(msg, want) {
				t.Fatalf("error %q does not contain %q", msg, want)
			}
		}
		if strings.Contains(msg, testCredential) {
			t.Fatalf("error leaked API key: %q", msg)
		}
	})
}

func TestLoadValidationErrorsAreActionableAndRedacted(t *testing.T) {
	t.Parallel()

	testCredential := strings.Repeat("x", 12)
	withCredential := func(values map[string]string) map[string]string {
		values[EnvAPIKey] = testCredential
		return values
	}

	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{name: "missing API key", env: map[string]string{EnvAthleteID: "123"}, wantErr: "missing intervals.icu API key"},
		{name: "missing athlete ID", env: withCredential(map[string]string{}), wantErr: "missing athlete ID"},
		{name: "invalid athlete ID", env: withCredential(map[string]string{EnvAthleteID: "abc"}), wantErr: "invalid athlete ID"},
		{name: "invalid timezone", env: withCredential(map[string]string{EnvAthleteID: "123", EnvTimezone: "Mars/Base"}), wantErr: "invalid timezone"},
		{name: "invalid timeout", env: withCredential(map[string]string{EnvAthleteID: "123", EnvHTTPTimeout: "0s"}), wantErr: "invalid HTTP timeout"},
		{name: "invalid base URL", env: withCredential(map[string]string{EnvAthleteID: "123", EnvAPIBaseURL: "ftp://example.test"}), wantErr: "invalid API base URL"},
		{name: "invalid transport", env: withCredential(map[string]string{EnvAthleteID: "123", EnvTransport: "websocket"}), wantErr: "invalid MCP transport"},
		{name: "invalid bind", env: withCredential(map[string]string{EnvAthleteID: "123", EnvHTTPBind: ":8765"}), wantErr: "invalid HTTP bind address"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Load(context.Background(), Options{DotEnvPath: t.TempDir() + "/missing.env", Env: tc.env})
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.wantErr) {
				t.Fatalf("error = %q, want to contain %q", msg, tc.wantErr)
			}
			if strings.Contains(msg, testCredential) {
				t.Fatalf("error leaked API key: %q", msg)
			}
		})
	}
}

func TestLoadTransportAndHTTPBindSelection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	dotEnvPath := dir + "/.env"
	writeFile(t, configPath, `{
		"api_key": "json-key",
		"athlete_id": "111",
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

func TestConfigStringRedactsSecret(t *testing.T) {
	t.Parallel()

	testCredential := strings.Repeat("x", 12)
	cfg := Config{
		APIKey:      testCredential,
		AthleteID:   "i12345",
		Timezone:    "UTC",
		APIBaseURL:  DefaultAPIBaseURL,
		HTTPTimeout: DefaultHTTPTimeout,
	}
	got := cfg.String()
	if strings.Contains(got, testCredential) || strings.Contains(got, "i12345") {
		t.Fatalf("Config.String() leaked sensitive data: %q", got)
	}
	for _, want := range []string{"api_key=<redacted>", "athlete_id=<set>", "UTC", "toolset=core"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Config.String() = %q, want %q", got, want)
		}
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

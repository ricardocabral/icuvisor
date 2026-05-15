package config

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/credstore"
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

func TestLoadDebugMetadataFromEnv(t *testing.T) {
	t.Parallel()

	cfg, err := Load(context.Background(), Options{Env: map[string]string{
		EnvAPIKey:        "env-key",
		EnvAthleteID:     "12345",
		EnvDebugMetadata: " TRUE ",
	}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.DebugMetadata {
		t.Fatal("DebugMetadata = false, want true")
	}
}

func TestParseDebugMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "true", in: "true", want: true},
		{name: "mixed case", in: " TRUE ", want: true},
		{name: "false", in: "false", want: false},
		{name: "invalid", in: "yes", want: false},
		{name: "empty", in: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ParseDebugMetadata(tt.in); got != tt.want {
				t.Fatalf("ParseDebugMetadata(%q) = %t, want %t", tt.in, got, tt.want)
			}
		})
	}
}

func TestLoadDotEnvExplicitMissingErrorsActionable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := Load(context.Background(), Options{
		DotEnvPath:     dir + "/missing.env",
		DotEnvExplicit: true,
		Env:            map[string]string{},
	})
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	msg := err.Error()
	for _, want := range []string{"env file", "not found", "--env-file", EnvDotEnvPath} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not contain %q", msg, want)
		}
	}
}

func TestLoadDotEnvEnvVarOverridesDefaultPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	customPath := dir + "/custom.env"
	writeFile(t, customPath, strings.Join([]string{
		"INTERVALS_ICU_API_KEY=custom-key",
		"INTERVALS_ICU_ATHLETE_ID=i777",
	}, "\n"))

	cfg, err := Load(context.Background(), Options{
		Env: map[string]string{EnvDotEnvPath: customPath},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "custom-key" || cfg.AthleteID != "i777" {
		t.Fatalf("Load() = api key %q athlete %q, want custom env-file values", cfg.APIKey, cfg.AthleteID)
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

func TestLoadCoachModeFromEnvAndDotEnv(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dotEnvPath := dir + "/.env"
	writeFile(t, dotEnvPath, strings.Join([]string{
		"INTERVALS_ICU_API_KEY=dotenv-key",
		"INTERVALS_ICU_ATHLETE_ID=i444",
		"ICUVISOR_COACH_MODE=auto",
	}, "\n"))

	cfg, err := Load(context.Background(), Options{DotEnvPath: dotEnvPath, Env: map[string]string{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CoachMode != coach.ModeAuto || cfg.CoachModeEnabled() {
		t.Fatalf("CoachMode = %q enabled=%t, want auto disabled with empty roster", cfg.CoachMode, cfg.CoachModeEnabled())
	}

	cfg, err = Load(context.Background(), Options{DotEnvPath: dotEnvPath, Env: map[string]string{
		EnvAPIKey:     "env-key",
		EnvAthleteID:  "12345",
		EnvCoachMode:  " ON ",
		EnvConfigPath: "",
	}})
	if err == nil {
		t.Fatal("Load() with coach mode on and empty roster error = nil, want error")
	}
	if !strings.Contains(err.Error(), "coach mode is on") {
		t.Fatalf("error = %q, want coach mode roster error", err)
	}

	_, err = Load(context.Background(), Options{Env: map[string]string{
		EnvAPIKey:     "env-key",
		EnvAthleteID:  "12345",
		EnvCoachMode:  "maybe",
		EnvConfigPath: "",
	}})
	if err == nil || !strings.Contains(err.Error(), "invalid coach mode") {
		t.Fatalf("Load() invalid coach mode error = %v, want invalid coach mode", err)
	}
}

func TestLoadCoachConfigSchemaAndValidation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	writeFile(t, configPath, `{
		"api_key": "json-key",
		"athlete_id": "111",
		"coach": {
			"athletes": [
				{"id": "222", "label": " Jane ", "allowed_tools": ["get_*", "get_*"], "denied_tools": ["delete_event"]},
				{"id": "i333", "allowed_tools": ["*"], "denied_tools": []}
			],
			"default_athlete_id": "333"
		}
	}`)

	cfg, err := Load(context.Background(), Options{Path: configPath, Env: map[string]string{EnvCoachMode: "auto"}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CoachMode != coach.ModeAuto || !cfg.CoachModeEnabled() {
		t.Fatalf("CoachMode = %q enabled=%t, want auto enabled", cfg.CoachMode, cfg.CoachModeEnabled())
	}
	if cfg.Coach.DefaultAthleteID != "i333" {
		t.Fatalf("DefaultAthleteID = %q, want i333", cfg.Coach.DefaultAthleteID)
	}
	if len(cfg.Coach.Athletes) != 2 || cfg.Coach.Athletes[0].ID != "i222" || cfg.Coach.Athletes[0].Label != "Jane" {
		t.Fatalf("Coach athletes = %#v, want normalized roster", cfg.Coach.Athletes)
	}
	if got := cfg.Coach.Athletes[0].AllowedTools; len(got) != 1 || got[0] != "get_*" {
		t.Fatalf("AllowedTools = %#v, want deduped get_*", got)
	}
}

func TestLoadCoachModeAllowsCoachOnlyConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode string
	}{
		{name: "on", mode: "on"},
		{name: "auto", mode: "auto"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := dir + "/config.json"
			writeFile(t, path, `{
				"api_key": "json-key",
				"coach": {
					"athletes": [{"id": "222", "allowed_tools": ["*"]}],
					"default_athlete_id": "222"
				}
			}`)
			cfg, err := Load(context.Background(), Options{Path: path, DotEnvPath: dir + "/missing.env", Env: map[string]string{EnvCoachMode: tc.mode}})
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if !cfg.CoachModeEnabled() || cfg.AthleteID != "i222" {
				t.Fatalf("Coach enabled=%t AthleteID=%q, want enabled with default i222", cfg.CoachModeEnabled(), cfg.AthleteID)
			}
		})
	}
}

func TestLoadCoachConfigValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		env     map[string]string
		wantErr string
	}{
		{name: "unknown json field", json: `{"api_key":"k","athlete_id":"1","coach":{"athletes":[],"typo":true}}`, wantErr: "unknown field"},
		{name: "duplicate normalized athlete", json: `{"api_key":"k","athlete_id":"1","coach":{"athletes":[{"id":"2"},{"id":"i2"}]}}`, wantErr: "duplicate coach athlete id"},
		{name: "default outside roster", json: `{"api_key":"k","athlete_id":"1","coach":{"athletes":[{"id":"2"}],"default_athlete_id":"3"}}`, wantErr: "coach.default_athlete_id"},
		{name: "unknown exact tool", json: `{"api_key":"k","athlete_id":"1","coach":{"athletes":[{"id":"2","allowed_tools":["get_athlete_profiel"]}]}}`, wantErr: "unknown athlete-scoped tool"},
		{name: "unknown wildcard", json: `{"api_key":"k","athlete_id":"1","coach":{"athletes":[{"id":"2","allowed_tools":["bogus_*"]}]}}`, wantErr: "matches no athlete-scoped tools"},
		{name: "off still validates stanza", json: `{"api_key":"k","athlete_id":"1","coach":{"athletes":[{"id":"2","denied_tools":["select_athlete"]}]}}`, env: map[string]string{EnvCoachMode: "off"}, wantErr: "unknown athlete-scoped tool"},
		{name: "on multiple athletes needs default", json: `{"api_key":"k","athlete_id":"1","coach":{"athletes":[{"id":"2"},{"id":"3"}]}}`, env: map[string]string{EnvCoachMode: "on"}, wantErr: "default_athlete_id is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := dir + "/config.json"
			writeFile(t, path, tc.json)
			env := tc.env
			if env == nil {
				env = map[string]string{}
			}
			_, err := Load(context.Background(), Options{Path: path, DotEnvPath: dir + "/missing.env", Env: env})
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %q, want %q", err, tc.wantErr)
			}
		})
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

	cfg, err = Load(context.Background(), Options{
		DotEnvPath: dir + "/missing.env",
		Env: map[string]string{
			EnvAPIKey:    "env-key",
			EnvAthleteID: "333",
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

func TestWriteStoresOnlyNonSecretFieldsAndRoundTrips(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/icuvisor/config.json"
	plainValue := strings.Join([]string{"must", "not", "write"}, "-")
	err := Write(context.Background(), path, Config{APIKey: plainValue, AthleteID: "12345", Timezone: "Europe/Madrid", APIBaseURL: DefaultAPIBaseURL}, WriteOptions{})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written config: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "api_key") || strings.Contains(content, plainValue) {
		t.Fatalf("written config leaked API key: %s", content)
	}
	if !strings.Contains(content, `"athlete_id": "i12345"`) || !strings.Contains(content, `"timezone": "Europe/Madrid"`) {
		t.Fatalf("written config missing normalized fields: %s", content)
	}
	if strings.Contains(content, "api_base_url") {
		t.Fatalf("default api_base_url should be omitted: %s", content)
	}
	loaded, err := Load(context.Background(), Options{Path: path, Env: map[string]string{}, CredentialStore: &fakeCredentialStore{value: "keychain-key"}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.APIKey != "keychain-key" || loaded.AthleteID != "i12345" || loaded.Timezone != "Europe/Madrid" || loaded.APIBaseURL != DefaultAPIBaseURL {
		t.Fatalf("loaded config = %+v", loaded)
	}
}

func TestWriteRefusesClobberWithoutAllowOverwrite(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/config.json"
	if err := Write(context.Background(), path, Config{AthleteID: "12345", Timezone: "UTC"}, WriteOptions{}); err != nil {
		t.Fatalf("initial Write() error = %v", err)
	}
	if err := Write(context.Background(), path, Config{AthleteID: "67890", Timezone: "UTC"}, WriteOptions{}); err == nil {
		t.Fatal("second Write() error = nil, want clobber refusal")
	}
	if err := Write(context.Background(), path, Config{AthleteID: "67890", Timezone: "UTC"}, WriteOptions{AllowOverwrite: true}); err != nil {
		t.Fatalf("overwrite Write() error = %v", err)
	}
	loaded, err := Load(context.Background(), Options{Path: path, Env: map[string]string{EnvAPIKey: "env-key"}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.AthleteID != "i67890" {
		t.Fatalf("AthleteID = %q, want i67890", loaded.AthleteID)
	}
}

type fakeCredentialStore struct {
	value string
	err   error
	calls int
}

func (f *fakeCredentialStore) Get(ctx context.Context, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.value, nil
}

func (f *fakeCredentialStore) Set(context.Context, string, string) error {
	return nil
}

func (f *fakeCredentialStore) Delete(context.Context, string) error {
	return nil
}

func TestLoadAPIKeyPrecedenceWithCredentialStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		json      string
		dotEnv    string
		env       map[string]string
		store     *fakeCredentialStore
		wantKey   string
		wantSrc   APIKeySource
		wantCalls int
		wantErr   string
	}{
		{
			name:      "process env wins and skips keychain",
			json:      `{"api_key":"json-key","athlete_id":"111"}`,
			dotEnv:    EnvAPIKey + `=dotenv-key\n` + EnvAthleteID + `=222`,
			env:       map[string]string{EnvAPIKey: "env-key", EnvAthleteID: "333"},
			store:     &fakeCredentialStore{err: errors.New("should not be called")},
			wantKey:   "env-key",
			wantSrc:   APIKeySourceEnv,
			wantCalls: 0,
		},
		{
			name:      "keychain beats plaintext files",
			json:      `{"api_key":"json-key","athlete_id":"111"}`,
			dotEnv:    EnvAPIKey + `=dotenv-key`,
			env:       map[string]string{},
			store:     &fakeCredentialStore{value: "keychain-key"},
			wantKey:   "keychain-key",
			wantSrc:   APIKeySourceKeychain,
			wantCalls: 1,
		},
		{
			name:      "not found falls through to file",
			json:      `{"api_key":"json-key","athlete_id":"111"}`,
			dotEnv:    EnvAPIKey + `=dotenv-key`,
			env:       map[string]string{},
			store:     &fakeCredentialStore{err: credstore.ErrNotFound},
			wantKey:   "json-key",
			wantSrc:   APIKeySourceFile,
			wantCalls: 1,
		},
		{
			name:      "dotenv supplies legacy file key when json omits it",
			json:      `{"athlete_id":"111"}`,
			dotEnv:    EnvAPIKey + `=dotenv-key`,
			env:       map[string]string{},
			store:     &fakeCredentialStore{err: credstore.ErrNotFound},
			wantKey:   "dotenv-key",
			wantSrc:   APIKeySourceFile,
			wantCalls: 1,
		},
		{
			name:      "unexpected keychain error fails load",
			json:      `{"api_key":"json-key","athlete_id":"111"}`,
			dotEnv:    "",
			env:       map[string]string{},
			store:     &fakeCredentialStore{err: errors.New("keychain unavailable in an unexpected way")},
			wantCalls: 1,
			wantErr:   "read intervals.icu API key from OS keychain",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			configPath := dir + "/config.json"
			dotEnvPath := dir + "/.env"
			writeFile(t, configPath, tc.json)
			writeFile(t, dotEnvPath, tc.dotEnv)

			cfg, err := Load(context.Background(), Options{Path: configPath, DotEnvPath: dotEnvPath, Env: tc.env, CredentialStore: tc.store})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("Load() error = %v, want containing %q", err, tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("Load() error = %v", err)
			} else {
				if cfg.APIKey != tc.wantKey || cfg.APIKeySource != tc.wantSrc {
					t.Fatalf("Load() api key/source = %q/%q, want %q/%q", cfg.APIKey, cfg.APIKeySource, tc.wantKey, tc.wantSrc)
				}
			}
			if tc.store != nil && tc.store.calls != tc.wantCalls {
				t.Fatalf("credential store calls = %d, want %d", tc.store.calls, tc.wantCalls)
			}
		})
	}
}

func TestLoadWarnsForLegacyFileAPIKeyWithoutLeakingValue(t *testing.T) {
	credential := strings.Repeat("w", 12)
	dir := t.TempDir()
	configPath := dir + "/config.json"
	writeFile(t, configPath, `{"api_key":"`+credential+`","athlete_id":"123"}`)

	var logs strings.Builder
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelWarn})))

	cfg, err := Load(context.Background(), Options{Path: configPath, DotEnvPath: dir + "/missing.env", Env: map[string]string{}, CredentialStore: credstore.NoopStore{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKeySource != APIKeySourceFile {
		t.Fatalf("APIKeySource = %q, want file", cfg.APIKeySource)
	}
	gotLogs := logs.String()
	if !strings.Contains(gotLogs, "api_key found in plaintext config") {
		t.Fatalf("logs = %q, want legacy warning", gotLogs)
	}
	if strings.Contains(gotLogs, credential) {
		t.Fatalf("logs leaked credential: %q", gotLogs)
	}
}

func TestConfigStringRedactsSecret(t *testing.T) {
	t.Parallel()

	testCredential := strings.Repeat("x", 12)
	cfg := Config{
		APIKey:       testCredential,
		APIKeySource: APIKeySourceKeychain,
		AthleteID:    "i12345",
		Timezone:     "UTC",
		APIBaseURL:   DefaultAPIBaseURL,
		HTTPTimeout:  DefaultHTTPTimeout,
		CoachMode:    coach.ModeAuto,
		Coach: coach.Config{Athletes: []coach.Athlete{
			{ID: "i222", Label: "Jane"},
		}},
	}
	got := cfg.String()
	if strings.Contains(got, testCredential) || strings.Contains(got, "i12345") || strings.Contains(got, "i222") || strings.Contains(got, "Jane") {
		t.Fatalf("Config.String() leaked sensitive data: %q", got)
	}
	for _, want := range []string{"api_key=<redacted>", "api_key_source=keychain", "athlete_id=<set>", "UTC", "toolset=core", "coach_mode=auto", "coach_enabled=true", "coach_athletes=1"} {
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

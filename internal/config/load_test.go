package config

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestLoadPrecedenceAndDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	dotEnvPath := dir + "/.env"
	writeFile(t, configPath, `{
		"api_key": "json-key",
		"athlete_id": "i111",
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
			EnvAthleteID:         "i333",
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
		EnvAthleteID:     "i12345",
		EnvDebugMetadata: " TRUE ",
	}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.DebugMetadata {
		t.Fatal("DebugMetadata = false, want true")
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
	if cfg.DeleteMode != safety.ModeSafe {
		t.Fatalf("DeleteMode = %q, want default %q", cfg.DeleteMode, safety.ModeSafe)
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

func TestLoadDeleteModeFromDotEnvAndProcessEnvPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		dotEnvMode  string // value written into .env; "" omits the line entirely
		processMode string // value of ICUVISOR_DELETE_MODE in process env; "" omits the key
		want        safety.Mode
	}{
		{name: "dotenv full honored when process env is silent", dotEnvMode: "full", processMode: "", want: safety.ModeFull},
		{name: "process env safe overrides dotenv full", dotEnvMode: "full", processMode: "safe", want: safety.ModeSafe},
		{name: "process env none overrides dotenv full", dotEnvMode: "full", processMode: "none", want: safety.ModeNone},
		{name: "invalid dotenv value never unlocks deletes", dotEnvMode: "bogus", processMode: "", want: safety.ModeSafe},
		{name: "absent from every source defaults to safe", dotEnvMode: "", processMode: "", want: safety.ModeSafe},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			dotEnvPath := dir + "/.env"
			lines := []string{
				"INTERVALS_ICU_API_KEY=dotenv-key",
				"INTERVALS_ICU_ATHLETE_ID=i444",
			}
			if tc.dotEnvMode != "" {
				lines = append(lines, safety.EnvDeleteMode+"="+tc.dotEnvMode)
			}
			writeFile(t, dotEnvPath, strings.Join(lines, "\n"))

			env := map[string]string{}
			if tc.processMode != "" {
				env[safety.EnvDeleteMode] = tc.processMode
			}

			cfg, err := Load(context.Background(), Options{DotEnvPath: dotEnvPath, Env: env})
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.DeleteMode != tc.want {
				t.Fatalf("DeleteMode = %q, want %q", cfg.DeleteMode, tc.want)
			}
		})
	}
}

func TestLoadUsesConfigPathFromEnv(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	writeFile(t, configPath, `{"api_key":"json-key","athlete_id":"i555"}`)

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

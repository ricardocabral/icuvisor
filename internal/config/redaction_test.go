package config

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestConfigLogValueStructuredAndRedacted(t *testing.T) {
	secret := "secret-xxxxx"
	cfg := Config{
		APIKey:           secret,
		AthleteID:        "i12345",
		APIBaseURL:       "https://example.invalid/api",
		HTTPBindAddress:  "127.0.0.1:9876",
		HTTPStateless:    true,
		HTTPJSONResponse: true,
		DeleteMode:       safety.ModeFull,
		Toolset:          safety.ToolsetFull,
		Coach: coach.Config{Athletes: []coach.Athlete{
			{ID: "i222", Label: "Jane"},
			{ID: "i333", Label: "John"},
		}},
	}

	if got := cfg.LogValue().Kind(); got != slog.KindGroup {
		t.Fatalf("Config.LogValue().Kind() = %s, want Group", got)
	}

	var buf bytes.Buffer
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	slog.Default().Info("config", "cfg", cfg)

	line := buf.String()
	for _, leak := range []string{"api_key", secret, "i12345", "i222", "i333", "Jane", "John"} {
		if strings.Contains(line, leak) {
			t.Fatalf("structured config log leaked %q in %s", leak, line)
		}
	}

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("decode slog JSON %q: %v", line, err)
	}
	cfgRecord, ok := record["cfg"].(map[string]any)
	if !ok {
		t.Fatalf("slog record cfg = %#v, want object", record["cfg"])
	}
	want := map[string]any{
		"api_base_url":         "https://example.invalid/api",
		"default_athlete_id":   "<set>",
		"http_bind":            "127.0.0.1:9876",
		"http_stateless":       true,
		"http_json_response":   true,
		"coach_athletes_count": float64(2),
		"delete_mode":          "full",
		"toolset":              "full",
	}
	if len(cfgRecord) != len(want) {
		t.Fatalf("structured config attrs = %#v, want exactly %#v", cfgRecord, want)
	}
	for key, wantValue := range want {
		if cfgRecord[key] != wantValue {
			t.Fatalf("structured config attr %s = %#v, want %#v in %#v", key, cfgRecord[key], wantValue, cfgRecord)
		}
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
	for _, want := range []string{"api_key=<redacted>", "api_key_source=keychain", "athlete_id=<set>", "UTC", "toolset=core", "coach_mode=auto", "coach_enabled=true", "coach_athletes=1", "http_stateless=false", "http_json_response=false"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Config.String() = %q, want %q", got, want)
		}
	}
}

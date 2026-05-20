package config

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/credstore"
)

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
			json:      `{"api_key":"json-key","athlete_id":"i111"}`,
			dotEnv:    EnvAPIKey + `=dotenv-key\n` + EnvAthleteID + `=222`,
			env:       map[string]string{EnvAPIKey: "env-key", EnvAthleteID: "i333"},
			store:     &fakeCredentialStore{err: errors.New("should not be called")},
			wantKey:   "env-key",
			wantSrc:   APIKeySourceEnv,
			wantCalls: 0,
		},
		{
			name:      "keychain beats plaintext files",
			json:      `{"api_key":"json-key","athlete_id":"i111"}`,
			dotEnv:    EnvAPIKey + `=dotenv-key`,
			env:       map[string]string{},
			store:     &fakeCredentialStore{value: "keychain-key"},
			wantKey:   "keychain-key",
			wantSrc:   APIKeySourceKeychain,
			wantCalls: 1,
		},
		{
			name:      "not found falls through to file",
			json:      `{"api_key":"json-key","athlete_id":"i111"}`,
			dotEnv:    EnvAPIKey + `=dotenv-key`,
			env:       map[string]string{},
			store:     &fakeCredentialStore{err: credstore.ErrNotFound},
			wantKey:   "json-key",
			wantSrc:   APIKeySourceFile,
			wantCalls: 1,
		},
		{
			name:      "dotenv supplies legacy file key when json omits it",
			json:      `{"athlete_id":"i111"}`,
			dotEnv:    EnvAPIKey + `=dotenv-key`,
			env:       map[string]string{},
			store:     &fakeCredentialStore{err: credstore.ErrNotFound},
			wantKey:   "dotenv-key",
			wantSrc:   APIKeySourceFile,
			wantCalls: 1,
		},
		{
			name:      "unexpected keychain error fails load",
			json:      `{"api_key":"json-key","athlete_id":"i111"}`,
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
	writeFile(t, configPath, `{"api_key":"`+credential+`","athlete_id":"i123"}`)

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
func TestLoadAcceptsSupportedCredentialReference(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	writeFile(t, configPath, `{
		"credential_ref": {"type":"keychain", "service":"icuvisor", "account":"intervals-icu-api-key"},
		"athlete_id": "i12345"
	}`)

	cfg, err := Load(context.Background(), Options{Path: configPath, Env: map[string]string{}, CredentialStore: &fakeCredentialStore{value: "keychain-key"}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CredentialRef != DefaultCredentialReference() {
		t.Fatalf("CredentialRef = %#v, want %#v", cfg.CredentialRef, DefaultCredentialReference())
	}
	if cfg.APIKey != "keychain-key" || cfg.APIKeySource != APIKeySourceKeychain {
		t.Fatalf("API key/source = %q/%q, want keychain", cfg.APIKey, cfg.APIKeySource)
	}
}

func TestLoadRejectsUnsupportedCredentialReference(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	writeFile(t, configPath, `{
		"credential_ref": {"type":"keychain", "service":"icuvisor", "account":"other-account"},
		"athlete_id": "i12345"
	}`)

	_, err := Load(context.Background(), Options{Path: configPath, Env: map[string]string{EnvAPIKey: "env-key"}})
	if err == nil || !strings.Contains(err.Error(), "unsupported credential_ref") || !strings.Contains(err.Error(), "intervals-icu-api-key") {
		t.Fatalf("Load() error = %v, want supported credential_ref guidance", err)
	}
}

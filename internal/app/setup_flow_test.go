package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
)

func TestRunSetupOfflineSkipsVerifyAndReadsAthleteIDTimezone(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{lines: []string{"i12345", "Europe/Madrid"}, secrets: []string{"api-key"}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Offline:         true,
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ConfigWriter:    noOpSetupConfigWriter,
		ProfileFetcher: func(context.Context, string, string) (SetupProfile, error) {
			t.Fatal("offline setup must not fetch profile")
			return SetupProfile{}, nil
		},
		TimezoneDetector: func() string {
			t.Fatal("offline setup must not autodetect timezone")
			return "UTC"
		},
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if got := prompter.linePrompts; len(got) != 2 || !strings.Contains(got[0], "Athlete ID") || !strings.Contains(got[1], "Timezone") {
		t.Fatalf("line prompts = %v, want athlete ID and timezone", got)
	}
	if !strings.Contains(stdout.String(), "Offline setup skips") || !strings.Contains(stdout.String(), "athlete id i12345") {
		t.Fatalf("stdout = %q, want offline and normalized athlete", stdout.String())
	}
}

func TestRunSetupWritesConfigAndVerifiesKeychainRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	store := &fakeSetupStore{getErr: credstore.ErrNotFound}
	secret := "super-secret-token"
	prompter := &fakeSetupPrompter{confirms: []bool{true}, lines: []string{"i12345"}, secrets: []string{secret}}
	fetchCalls := 0
	var gotAthleteIDs []string
	var stdout bytes.Buffer
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      configPath,
		Stdout:          &stdout,
		CredentialStore: store,
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ProfileFetcher: func(_ context.Context, _ string, athleteID string) (SetupProfile, error) {
			fetchCalls++
			gotAthleteIDs = append(gotAthleteIDs, athleteID)
			return SetupProfile{AthleteID: "i12345", DisplayName: "Jane Doe", FTP: 245}, nil
		},
		TimezoneDetector: func() string { return "Europe/Madrid" },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if fetchCalls != 2 {
		t.Fatalf("profile fetch calls = %d, want pre-write and final test", fetchCalls)
	}
	if len(store.sets) != 1 || store.sets[0] != secret {
		t.Fatalf("store sets = %v, want %q", store.sets, secret)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(data), "api_key") || strings.Contains(string(data), secret) {
		t.Fatalf("config leaked API key: %s", data)
	}
	for _, want := range []string{`"credential_ref"`, `"type": "keychain"`, `"service": "icuvisor"`, `"account": "intervals-icu-api-key"`} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("config missing credential reference %q: %s", want, data)
		}
	}
	cfg, err := config.Load(context.Background(), config.Options{Path: configPath, Env: map[string]string{}, CredentialStore: store})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != secret || cfg.AthleteID != "i12345" || cfg.Timezone != "Europe/Madrid" || cfg.APIBaseURL != config.DefaultAPIBaseURL {
		t.Fatalf("loaded config = %+v", cfg)
	}
	if got, want := prompter.secretPrompts, []string{"Paste your intervals.icu API key (from https://intervals.icu/settings):"}; !slices.Equal(got, want) {
		t.Fatalf("secret prompts = %v, want %v", got, want)
	}
	if got, want := prompter.confirmPrompts, []string{"Detected timezone: Europe/Madrid. Use this? [Y/n]"}; !slices.Equal(got, want) {
		t.Fatalf("confirm prompts = %v, want %v", got, want)
	}
	if got := prompter.linePrompts; len(got) != 1 || !strings.Contains(got[0], "Athlete ID") {
		t.Fatalf("line prompts = %v, want athlete ID prompt", got)
	}
	if got, want := gotAthleteIDs, []string{"i12345", "i12345"}; !slices.Equal(got, want) {
		t.Fatalf("fetcher athlete IDs = %v, want %v", got, want)
	}
	wantStdout := strings.Join([]string{
		"Welcome to icuvisor.",
		"This setup stores your intervals.icu API key in the OS keychain and writes non-secret settings to your icuvisor config file.",
		"Checking intervals.icu… connected as \"Jane Doe\" (athlete i12345, FTP 245 W).",
		"Saved. Your key is in the OS keychain; athlete id i12345 + timezone Europe/Madrid are in " + configPath + ".",
		"Test connection OK: Jane Doe, FTP 245 W.",
		"Next: point Claude Desktop at icuvisor — see https://icuvisor.app/connect/claude-desktop/",
	}, "\n") + "\n"
	if stdout.String() != wantStdout {
		t.Fatalf("stdout = %q, want %q", stdout.String(), wantStdout)
	}
}

func TestRunSetupKeychainWriteFailuresDoNotClaimSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		store *fakeSetupStore
		want  string
	}{
		{name: "set failure", store: &fakeSetupStore{getErr: credstore.ErrNotFound, setErr: errors.New("keychain unavailable")}, want: "store intervals.icu API key"},
		{name: "get failure", store: &fakeSetupStore{getErr: credstore.ErrNotFound, getErrAfterSet: errors.New("keychain read failed")}, want: "verify intervals.icu API key"},
		{name: "mismatch", store: &fakeSetupStore{getErr: credstore.ErrNotFound, mismatchAfterSet: true}, want: "stored API key verification failed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			configPath := t.TempDir() + "/config.json"
			err := RunSetup(context.Background(), SetupOptions{
				ConfigPath:      configPath,
				Stdout:          &stdout,
				CredentialStore: tc.store,
				Prompter:        &fakeSetupPrompter{confirms: []bool{true}, lines: []string{"i12345"}, secrets: []string{"api-key"}},
				ConfigExists:    func(string) (bool, error) { return false, nil },
				ProfileFetcher: func(context.Context, string, string) (SetupProfile, error) {
					return SetupProfile{AthleteID: "i12345", DisplayName: "Jane Doe"}, nil
				},
				TimezoneDetector: func() string { return "UTC" },
			})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("RunSetup() error = %v, want %q", err, tc.want)
			}
			if strings.Contains(stdout.String(), "Test connection OK") {
				t.Fatalf("stdout claimed success after failure: %q", stdout.String())
			}
			if _, statErr := os.Stat(configPath); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("config file exists after keychain failure: stat error = %v", statErr)
			}
		})
	}
}

func TestDetectLocalTimezoneUsesIANAZoneWhenLocalNameIsLocal(t *testing.T) {
	t.Parallel()

	got := detectLocalTimezoneWith("Local", "", func(path string) (string, error) {
		if path != "/etc/localtime" {
			t.Fatalf("readlink path = %q, want /etc/localtime", path)
		}
		return "/var/db/timezone/zoneinfo/America/Sao_Paulo", nil
	})
	if got != "America/Sao_Paulo" {
		t.Fatalf("timezone = %q, want America/Sao_Paulo", got)
	}
}

func TestDetectLocalTimezonePrefersValidTZEnvironment(t *testing.T) {
	t.Parallel()

	got := detectLocalTimezoneWith("Local", ":Europe/Madrid", func(string) (string, error) {
		t.Fatal("readlink must not be called when TZ is valid")
		return "", nil
	})
	if got != "Europe/Madrid" {
		t.Fatalf("timezone = %q, want Europe/Madrid", got)
	}
}

func TestRunSetupStillPromptsForExistingKeyWithForce(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{false}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Force:           true,
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{secret: "stored"},
		Prompter:        prompter,
		ConfigExists: func(string) (bool, error) {
			t.Fatal("config existence must not be checked after key overwrite denial")
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if got := prompter.confirmPrompts; len(got) != 1 || !strings.Contains(got[0], "API key is already stored") {
		t.Fatalf("confirm prompts = %v, want existing-key prompt", got)
	}
	if len(prompter.secretPrompts) != 0 {
		t.Fatalf("ReadSecret prompts = %v, want none", prompter.secretPrompts)
	}
}

func TestRunSetupNetworkErrorMentionsOfflineOverride(t *testing.T) {
	t.Parallel()

	prompter := &fakeSetupPrompter{lines: []string{"i12345"}, secrets: []string{"api-key"}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ProfileFetcher: func(context.Context, string, string) (SetupProfile, error) {
			return SetupProfile{}, errors.New("dial tcp timeout")
		},
	})
	if err == nil {
		t.Fatal("RunSetup() error = nil, want network error")
	}
	if !strings.Contains(err.Error(), "Nothing was written") || !strings.Contains(err.Error(), "--offline") {
		t.Fatalf("error = %q, want offline guidance", err.Error())
	}
}

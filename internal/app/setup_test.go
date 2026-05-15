package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeSetupStore struct {
	secret           string
	getErr           error
	setErr           error
	getErrAfterSet   error
	mismatchAfterSet bool
	sets             []string
}

func (s *fakeSetupStore) Get(ctx context.Context, account string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if account != credstore.IntervalsAPIKeyAccount {
		return "", errors.New("unexpected account")
	}
	if s.getErr != nil {
		return "", s.getErr
	}
	return s.secret, nil
}

func (s *fakeSetupStore) Set(ctx context.Context, account, secret string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if account != credstore.IntervalsAPIKeyAccount {
		return errors.New("unexpected account")
	}
	if s.setErr != nil {
		return s.setErr
	}
	s.sets = append(s.sets, secret)
	if s.mismatchAfterSet {
		s.secret = strings.Repeat("x", len(secret)+1)
	} else {
		s.secret = secret
	}
	s.getErr = s.getErrAfterSet
	return nil
}

func (s *fakeSetupStore) Delete(ctx context.Context, account string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if account != credstore.IntervalsAPIKeyAccount {
		return errors.New("unexpected account")
	}
	return nil
}

type fakeSetupPrompter struct {
	confirms       []bool
	confirmPrompts []string
	lines          []string
	linePrompts    []string
	secrets        []string
	secretPrompts  []string
}

func (p *fakeSetupPrompter) Confirm(ctx context.Context, prompt string, _ bool) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	p.confirmPrompts = append(p.confirmPrompts, prompt)
	if len(p.confirms) == 0 {
		return false, errors.New("unexpected confirm prompt")
	}
	answer := p.confirms[0]
	p.confirms = p.confirms[1:]
	return answer, nil
}

func (p *fakeSetupPrompter) ReadLine(ctx context.Context, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	p.linePrompts = append(p.linePrompts, prompt)
	if len(p.lines) == 0 {
		return "", errors.New("unexpected line prompt")
	}
	line := p.lines[0]
	p.lines = p.lines[1:]
	return line, nil
}

func noOpSetupConfigWriter(context.Context, string, config.Config, config.WriteOptions) error {
	return nil
}

func (p *fakeSetupPrompter) ReadSecret(ctx context.Context, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	p.secretPrompts = append(p.secretPrompts, prompt)
	if len(p.secrets) == 0 {
		return "", errors.New("unexpected secret prompt")
	}
	secret := p.secrets[0]
	p.secrets = p.secrets[1:]
	return secret, nil
}

func TestRunSetupExistingKeyDefaultNoCancelsBeforeReadingSecret(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{false}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
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
	if !strings.Contains(stdout.String(), "Setup canceled; nothing changed.") {
		t.Fatalf("stdout %q missing cancellation", stdout.String())
	}
	if len(prompter.secretPrompts) != 0 {
		t.Fatalf("ReadSecret prompts = %v, want none", prompter.secretPrompts)
	}
	if got := prompter.confirmPrompts; len(got) != 1 || !strings.Contains(got[0], "API key is already stored") {
		t.Fatalf("confirm prompts = %v, want existing-key prompt", got)
	}
}

func TestRunSetupExistingConfigDefaultNoCancelsBeforeReadingSecret(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{false}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(path string) (bool, error) { return path == "/tmp/icuvisor.json", nil },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Setup canceled; nothing changed.") {
		t.Fatalf("stdout %q missing cancellation", stdout.String())
	}
	if len(prompter.secretPrompts) != 0 {
		t.Fatalf("ReadSecret prompts = %v, want none", prompter.secretPrompts)
	}
	if got := prompter.confirmPrompts; len(got) != 1 || !strings.Contains(got[0], "config file already exists") {
		t.Fatalf("confirm prompts = %v, want existing-config prompt", got)
	}
}

func TestRunSetupForceSkipsOnlyConfigPrompt(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{true}, secrets: []string{"api-key"}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Force:           true,
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(path string) (bool, error) { return path == "/tmp/icuvisor.json", nil },
		ConfigWriter: func(_ context.Context, _ string, _ config.Config, opts config.WriteOptions) error {
			if !opts.AllowOverwrite {
				t.Fatal("--force must allow config overwrite")
			}
			return nil
		},
		ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
			return SetupProfile{AthleteID: "12345", DisplayName: "Jane Doe", FTP: 245}, nil
		},
		TimezoneDetector: func() string { return "UTC" },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if len(prompter.confirmPrompts) != 1 || !strings.Contains(prompter.confirmPrompts[0], "Detected timezone") {
		t.Fatalf("confirm prompts = %v, want only timezone prompt", prompter.confirmPrompts)
	}
	if got := prompter.secretPrompts; len(got) != 1 || !strings.Contains(got[0], "intervals.icu API key") {
		t.Fatalf("secret prompts = %v, want API-key prompt", got)
	}
}

func TestRunSetupFetchesProfileNormalizesIDAndConfirmsTimezone(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{true}, secrets: []string{"api-key"}}
	var gotKey string
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ConfigWriter:    noOpSetupConfigWriter,
		ProfileFetcher: func(_ context.Context, apiKey string) (SetupProfile, error) {
			gotKey = apiKey
			return SetupProfile{AthleteID: "12345", DisplayName: "Jane Doe", FTP: 245}, nil
		},
		TimezoneDetector: func() string { return "Europe/Madrid" },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if gotKey != "api-key" {
		t.Fatalf("profile fetcher key = %q, want api-key", gotKey)
	}
	for _, want := range []string{"connected as \"Jane Doe\"", "athlete i12345", "FTP 245 W", "timezone Europe/Madrid"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout %q missing %q", stdout.String(), want)
		}
	}
}

func TestRunSetupAllowsTimezoneOverride(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{false}, lines: []string{"America/Sao_Paulo"}, secrets: []string{"api-key"}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ConfigWriter:    noOpSetupConfigWriter,
		ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
			return SetupProfile{AthleteID: "i12345", DisplayName: "Jane Doe"}, nil
		},
		TimezoneDetector: func() string { return "Europe/Madrid" },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if got := prompter.linePrompts; len(got) != 1 || !strings.Contains(got[0], "Timezone") {
		t.Fatalf("line prompts = %v, want timezone override prompt", got)
	}
	if !strings.Contains(stdout.String(), "timezone America/Sao_Paulo") {
		t.Fatalf("stdout %q missing timezone override", stdout.String())
	}
}

func TestRunSetupUnauthorizedKeyReturnsFixURL(t *testing.T) {
	t.Parallel()

	prompter := &fakeSetupPrompter{secrets: []string{"bad-key"}}
	store := &fakeSetupStore{getErr: credstore.ErrNotFound}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		CredentialStore: store,
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
			return SetupProfile{}, errors.Join(errors.New("wrapped"), intervals.ErrUnauthorized)
		},
	})
	if err == nil {
		t.Fatal("RunSetup() error = nil, want unauthorized error")
	}
	if !strings.Contains(err.Error(), "API key not accepted") || !strings.Contains(err.Error(), "https://intervals.icu/settings") {
		t.Fatalf("error = %q, want fix URL", err.Error())
	}
	if len(store.sets) != 0 {
		t.Fatalf("Set calls = %v, want none", store.sets)
	}
	if len(prompter.linePrompts) != 0 || len(prompter.confirmPrompts) != 0 {
		t.Fatalf("prompts after unauthorized = confirms %v lines %v, want none", prompter.confirmPrompts, prompter.linePrompts)
	}
}

func TestRunSetupNetworkErrorMentionsOfflineOverride(t *testing.T) {
	t.Parallel()

	prompter := &fakeSetupPrompter{secrets: []string{"api-key"}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
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

func TestRunSetupOfflineSkipsVerifyAndReadsAthleteIDTimezone(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{lines: []string{"12345", "Europe/Madrid"}, secrets: []string{"api-key"}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Offline:         true,
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ConfigWriter:    noOpSetupConfigWriter,
		ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
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
	prompter := &fakeSetupPrompter{confirms: []bool{true}, secrets: []string{"api-key"}}
	fetchCalls := 0
	var stdout bytes.Buffer
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      configPath,
		Stdout:          &stdout,
		CredentialStore: store,
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
			fetchCalls++
			return SetupProfile{AthleteID: "12345", DisplayName: "Jane Doe", FTP: 245}, nil
		},
		TimezoneDetector: func() string { return "Europe/Madrid" },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if fetchCalls != 2 {
		t.Fatalf("profile fetch calls = %d, want pre-write and final test", fetchCalls)
	}
	if len(store.sets) != 1 || store.sets[0] != "api-key" {
		t.Fatalf("store sets = %v, want api-key", store.sets)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(data), "api_key") || strings.Contains(string(data), "api-key") {
		t.Fatalf("config leaked API key: %s", data)
	}
	cfg, err := config.Load(context.Background(), config.Options{Path: configPath, Env: map[string]string{}, CredentialStore: store})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "api-key" || cfg.AthleteID != "i12345" || cfg.Timezone != "Europe/Madrid" || cfg.APIBaseURL != config.DefaultAPIBaseURL {
		t.Fatalf("loaded config = %+v", cfg)
	}
	for _, want := range []string{"Saved. Your key is in the OS keychain", "Test connection OK: Jane Doe, FTP 245 W", "docs/clients/claude-desktop.md"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout %q missing %q", stdout.String(), want)
		}
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
			err := RunSetup(context.Background(), SetupOptions{
				ConfigPath:      t.TempDir() + "/config.json",
				Stdout:          &stdout,
				CredentialStore: tc.store,
				Prompter:        &fakeSetupPrompter{confirms: []bool{true}, secrets: []string{"api-key"}},
				ConfigExists:    func(string) (bool, error) { return false, nil },
				ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
					return SetupProfile{AthleteID: "12345", DisplayName: "Jane Doe"}, nil
				},
				TimezoneDetector: func() string { return "UTC" },
			})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("RunSetup() error = %v, want %q", err, tc.want)
			}
			if strings.Contains(stdout.String(), "Test connection OK") {
				t.Fatalf("stdout claimed success after failure: %q", stdout.String())
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

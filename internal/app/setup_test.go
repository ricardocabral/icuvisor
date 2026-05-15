package app

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/credstore"
)

type fakeSetupStore struct {
	secret string
	getErr error
	sets   []string
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
	s.sets = append(s.sets, secret)
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
	prompter := &fakeSetupPrompter{secrets: []string{"api-key"}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Force:           true,
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(path string) (bool, error) { return path == "/tmp/icuvisor.json", nil },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if len(prompter.confirmPrompts) != 0 {
		t.Fatalf("confirm prompts = %v, want none", prompter.confirmPrompts)
	}
	if got := prompter.secretPrompts; len(got) != 1 || !strings.Contains(got[0], "intervals.icu API key") {
		t.Fatalf("secret prompts = %v, want API-key prompt", got)
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

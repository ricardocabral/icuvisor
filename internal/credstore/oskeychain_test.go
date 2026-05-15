package credstore

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	keyring "github.com/zalando/go-keyring"
)

type fakeKeyringBackend struct {
	getSecret string
	getErr    error
	setErr    error
	deleteErr error

	gotService string
	gotAccount string
	gotSecret  string
}

func (f *fakeKeyringBackend) Get(service, account string) (string, error) {
	f.gotService = service
	f.gotAccount = account
	if f.getErr != nil {
		return "", f.getErr
	}
	return f.getSecret, nil
}

func (f *fakeKeyringBackend) Set(service, account, secret string) error {
	f.gotService = service
	f.gotAccount = account
	f.gotSecret = secret
	return f.setErr
}

func (f *fakeKeyringBackend) Delete(service, account string) error {
	f.gotService = service
	f.gotAccount = account
	return f.deleteErr
}

func TestOSKeychainReturnsStore(t *testing.T) {
	t.Parallel()

	if OSKeychain() == nil {
		t.Fatal("OSKeychain() = nil, want Store")
	}
}

func TestKeyringStoreGetSuccess(t *testing.T) {
	t.Parallel()

	backend := &fakeKeyringBackend{getSecret: "secret-value"}
	store := keyringStore{backend: backend}
	got, err := store.Get(context.Background(), IntervalsAPIKeyAccount)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != "secret-value" {
		t.Fatalf("Get() = %q, want secret-value", got)
	}
	if backend.gotService != ServiceName || backend.gotAccount != IntervalsAPIKeyAccount {
		t.Fatalf("backend called with service=%q account=%q", backend.gotService, backend.gotAccount)
	}
}

func TestKeyringStoreMapsNotFound(t *testing.T) {
	t.Parallel()

	store := keyringStore{backend: &fakeKeyringBackend{getErr: keyring.ErrNotFound}}
	secret, err := store.Get(context.Background(), IntervalsAPIKeyAccount)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
	if secret != "" {
		t.Fatalf("Get() secret = %q, want empty", secret)
	}

	store = keyringStore{backend: &fakeKeyringBackend{deleteErr: keyring.ErrNotFound}}
	if err := store.Delete(context.Background(), IntervalsAPIKeyAccount); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestKeyringStoreWrapsUnexpectedErrors(t *testing.T) {
	t.Parallel()

	backendErr := errors.New("permission denied by keychain")
	store := keyringStore{backend: &fakeKeyringBackend{getErr: backendErr}}
	_, err := store.Get(context.Background(), IntervalsAPIKeyAccount)
	if !errors.Is(err, backendErr) {
		t.Fatalf("Get() error = %v, want wrapped backend error", err)
	}
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want not ErrNotFound", err)
	}

	store = keyringStore{backend: &fakeKeyringBackend{setErr: backendErr}}
	if err := store.Set(context.Background(), IntervalsAPIKeyAccount, "secret-value"); !errors.Is(err, backendErr) {
		t.Fatalf("Set() error = %v, want wrapped backend error", err)
	}

	store = keyringStore{backend: &fakeKeyringBackend{deleteErr: backendErr}}
	if err := store.Delete(context.Background(), IntervalsAPIKeyAccount); !errors.Is(err, backendErr) {
		t.Fatalf("Delete() error = %v, want wrapped backend error", err)
	}
}

func TestKeyringStoreHonorsContextBeforeBackend(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	backend := &fakeKeyringBackend{}
	store := keyringStore{backend: backend}

	if _, err := store.Get(ctx, IntervalsAPIKeyAccount); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() error = %v, want context.Canceled", err)
	}
	if err := store.Set(ctx, IntervalsAPIKeyAccount, "secret-value"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Set() error = %v, want context.Canceled", err)
	}
	if err := store.Delete(ctx, IntervalsAPIKeyAccount); !errors.Is(err, context.Canceled) {
		t.Fatalf("Delete() error = %v, want context.Canceled", err)
	}
	if backend.gotAccount != "" {
		t.Fatalf("backend called despite canceled context: account=%q", backend.gotAccount)
	}
}

func TestKeyringStoreLogsDoNotContainSecret(t *testing.T) {
	t.Parallel()

	credential := strings.Join([]string{"sentinel", "redaction", "token"}, "-")
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	backend := &fakeKeyringBackend{}
	store := keyringStore{backend: backend, logger: logger}

	if err := store.Set(context.Background(), IntervalsAPIKeyAccount, credential); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	logs := buf.String()
	if strings.Contains(logs, credential) {
		t.Fatalf("logs leaked credential: %q", logs)
	}
	for _, want := range []string{"credential set started", "credential set succeeded", "service=icuvisor", "account=intervals-icu-api-key"} {
		if !strings.Contains(logs, want) {
			t.Fatalf("logs = %q, want %q", logs, want)
		}
	}
}

package credstore

import (
	"context"
	"errors"
	"testing"
)

func TestCanonicalNames(t *testing.T) {
	t.Parallel()

	if ServiceName != "icuvisor" {
		t.Fatalf("ServiceName = %q, want icuvisor", ServiceName)
	}
	if IntervalsAPIKeyAccount != "intervals-icu-api-key" {
		t.Fatalf("IntervalsAPIKeyAccount = %q, want intervals-icu-api-key", IntervalsAPIKeyAccount)
	}
}

func TestNoopStoreContract(t *testing.T) {
	t.Parallel()

	var store Store = NoopStore{}
	secret, err := store.Get(context.Background(), IntervalsAPIKeyAccount)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
	if secret != "" {
		t.Fatalf("Get() secret = %q, want empty", secret)
	}
	if err := store.Set(context.Background(), IntervalsAPIKeyAccount, "secret-value"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := store.Delete(context.Background(), IntervalsAPIKeyAccount); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestNoopStoreHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var store Store = NoopStore{}
	if _, err := store.Get(ctx, IntervalsAPIKeyAccount); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() error = %v, want context.Canceled", err)
	}
	if err := store.Set(ctx, IntervalsAPIKeyAccount, "secret-value"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Set() error = %v, want context.Canceled", err)
	}
	if err := store.Delete(ctx, IntervalsAPIKeyAccount); !errors.Is(err, context.Canceled) {
		t.Fatalf("Delete() error = %v, want context.Canceled", err)
	}
}

func TestErrNotFoundSupportsErrorsIs(t *testing.T) {
	t.Parallel()

	err := errors.Join(errors.New("credential lookup failed"), ErrNotFound)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("errors.Is(%v, ErrNotFound) = false, want true", err)
	}
}

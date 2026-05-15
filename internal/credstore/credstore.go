package credstore

import (
	"context"
	"errors"
)

const (
	// ServiceName is the OS-keychain service namespace used by icuvisor.
	ServiceName = "icuvisor"
	// IntervalsAPIKeyAccount is the single-host intervals.icu API-key account name.
	IntervalsAPIKeyAccount = "intervals-icu-api-key"
)

// ErrNotFound reports that a credential does not exist or that a read-only
// startup lookup cannot reach the OS keychain and should fall through.
var ErrNotFound = errors.New("credential not found")

// Store reads and writes secrets in a local credential store.
type Store interface {
	Get(ctx context.Context, account string) (string, error)
	Set(ctx context.Context, account, secret string) error
	Delete(ctx context.Context, account string) error
}

// NoopStore is a credential store for tests and explicit keychain opt-out paths.
type NoopStore struct{}

// Get always reports ErrNotFound without touching an OS keychain.
func (NoopStore) Get(ctx context.Context, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "", ErrNotFound
}

// Set discards the secret after honoring context cancellation.
func (NoopStore) Set(ctx context.Context, _, _ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

// Delete is a no-op after honoring context cancellation.
func (NoopStore) Delete(ctx context.Context, _ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

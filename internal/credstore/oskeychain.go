package credstore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	keyring "github.com/zalando/go-keyring"
)

type keyringBackend interface {
	Get(service, account string) (string, error)
	Set(service, account, secret string) error
	Delete(service, account string) error
}

type goKeyringBackend struct{}

func (goKeyringBackend) Get(service, account string) (string, error) {
	return keyring.Get(service, account)
}

func (goKeyringBackend) Set(service, account, secret string) error {
	return keyring.Set(service, account, secret)
}

func (goKeyringBackend) Delete(service, account string) error {
	return keyring.Delete(service, account)
}

type keyringStore struct {
	service string
	backend keyringBackend
	logger  *slog.Logger
}

// OSKeychain returns the platform-native credential store for this host.
func OSKeychain() Store {
	return keyringStore{service: ServiceName, backend: goKeyringBackend{}, logger: slog.Default()}
}

func (s keyringStore) Get(ctx context.Context, account string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	service := s.serviceName()
	s.log("credential get started", account)
	secret, err := s.backend.Get(service, account)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", ctxErr
	}
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) || isKeychainUnavailable(err) {
			s.log("credential get not found", account)
			return "", ErrNotFound
		}
		s.log("credential get failed", account)
		return "", fmt.Errorf("get credential %q from OS keychain: %w", account, err)
	}
	s.log("credential get succeeded", account)
	return secret, nil
}

func (s keyringStore) Set(ctx context.Context, account, secret string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	service := s.serviceName()
	s.log("credential set started", account)
	if err := s.backend.Set(service, account, secret); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			s.log("credential set not found", account)
			return ErrNotFound
		}
		s.log("credential set failed", account)
		return fmt.Errorf("set credential %q in OS keychain: %w", account, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.log("credential set succeeded", account)
	return nil
}

func (s keyringStore) Delete(ctx context.Context, account string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	service := s.serviceName()
	s.log("credential delete started", account)
	if err := s.backend.Delete(service, account); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			s.log("credential delete not found", account)
			return ErrNotFound
		}
		s.log("credential delete failed", account)
		return fmt.Errorf("delete credential %q from OS keychain: %w", account, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.log("credential delete succeeded", account)
	return nil
}

func (s keyringStore) serviceName() string {
	if s.service != "" {
		return s.service
	}
	return ServiceName
}

func (s keyringStore) log(message, account string) {
	logger := s.logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.Debug(message, "service", s.serviceName(), "account", account)
}

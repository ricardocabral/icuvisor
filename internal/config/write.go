package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type writeFileConfig struct {
	CredentialRef CredentialReference `json:"credential_ref"`
	AthleteID     string              `json:"athlete_id"`
	Timezone      string              `json:"timezone"`
	APIBaseURL    string              `json:"api_base_url,omitempty"`
}

// Write stores non-secret config fields as JSON.
func Write(ctx context.Context, path string, cfg Config, opts WriteOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		var err error
		trimmedPath, err = DefaultPath()
		if err != nil {
			return err
		}
	}
	athleteID, err := NormalizeAthleteID(cfg.AthleteID)
	if err != nil {
		return fmt.Errorf("normalizing athlete ID for config write: %w", err)
	}
	timezoneName := strings.TrimSpace(cfg.Timezone)
	if timezoneName == "" {
		timezoneName = DefaultTimezone
	}
	if _, err := time.LoadLocation(timezoneName); err != nil {
		return fmt.Errorf("validating timezone for config write: %w", err)
	}
	apiBaseURL := strings.TrimSpace(cfg.APIBaseURL)
	if apiBaseURL == DefaultAPIBaseURL {
		apiBaseURL = ""
	}
	payload, err := json.MarshalIndent(writeFileConfig{CredentialRef: DefaultCredentialReference(), AthleteID: athleteID, Timezone: timezoneName, APIBaseURL: apiBaseURL}, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config file: %w", err)
	}
	payload = append(payload, '\n')
	if err := ctx.Err(); err != nil {
		return err
	}
	return writeConfigFile(trimmedPath, payload, opts.AllowOverwrite)
}

func writeConfigFile(path string, payload []byte, allowOverwrite bool) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory %q: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".icuvisor-config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary config file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("set temporary config permissions: %w", err)
	}
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary config file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary config file: %w", err)
	}
	if allowOverwrite {
		if err := os.Rename(tmpPath, path); err != nil {
			return fmt.Errorf("replace config file %q: %w", path, err)
		}
		return nil
	}
	if err := os.Link(tmpPath, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("config file %q already exists; rerun setup with --force or approve overwrite prompt", path)
		}
		return fmt.Errorf("create config file %q: %w", path, err)
	}
	return nil
}

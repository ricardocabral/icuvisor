package app

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/response"
)

func TestRunVersionWritesInjectedVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := Run(context.Background(), Options{
		Version: "v1.2.3-test",
		Args:    []string{"version"},
		Stdout:  &stdout,
		LoadConfig: func(context.Context, config.Options) (config.Config, error) {
			t.Fatal("version command must not load config")
			return config.Config{}, nil
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got, want := stdout.String(), "v1.2.3-test\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunDefaultDelegatesToStarterWithVersionAndConfig(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("starter failed")
	wantConfig := config.Config{
		APIKey:      "secret",
		AthleteID:   "i12345",
		Timezone:    "UTC",
		APIBaseURL:  config.DefaultAPIBaseURL,
		HTTPTimeout: 30 * time.Second,
	}
	var gotInfo ServerInfo
	err := Run(context.Background(), Options{
		Version: "v9.8.7",
		LoadConfig: func(_ context.Context, opts config.Options) (config.Config, error) {
			if opts.Path != "" {
				t.Fatalf("config path = %q, want empty", opts.Path)
			}
			return wantConfig, nil
		},
		StartServer: func(_ context.Context, info ServerInfo) error {
			gotInfo = info
			return wantErr
		},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	if gotInfo.Version != "v9.8.7" {
		t.Fatalf("server version = %q, want %q", gotInfo.Version, "v9.8.7")
	}
	if gotInfo.Config.AthleteID != wantConfig.AthleteID {
		t.Fatalf("server athlete ID = %q, want %q", gotInfo.Config.AthleteID, wantConfig.AthleteID)
	}
}

func TestDefaultStartServerLogsStartupVersion(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo})))

	err := defaultStartServer(context.Background(), ServerInfo{Version: "v7.8.9"})
	if err == nil {
		t.Fatal("defaultStartServer() error = nil, want config/client error")
	}
	out := logs.String()
	for _, want := range []string{"server starting", "version=v7.8.9"} {
		if !strings.Contains(out, want) {
			t.Fatalf("startup log %q missing %q", out, want)
		}
	}
}

func TestRunCapturesDebugMetadataOnceForServerInfo(t *testing.T) {
	t.Setenv(response.EnvDebugMetadata, "true")
	wantConfig := config.Config{
		APIKey:      "secret",
		AthleteID:   "i12345",
		Timezone:    "UTC",
		APIBaseURL:  config.DefaultAPIBaseURL,
		HTTPTimeout: 30 * time.Second,
	}
	var gotInfo ServerInfo
	wantErr := errors.New("stop")
	err := Run(context.Background(), Options{
		LoadConfig: func(context.Context, config.Options) (config.Config, error) {
			t.Setenv(response.EnvDebugMetadata, "false")
			return wantConfig, nil
		},
		StartServer: func(_ context.Context, info ServerInfo) error {
			gotInfo = info
			return wantErr
		},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	if !gotInfo.DebugMetadata {
		t.Fatal("DebugMetadata = false, want startup-captured true")
	}
}

func TestRunDefaultPassesConfigPath(t *testing.T) {
	t.Parallel()

	var gotPath string
	err := Run(context.Background(), Options{
		Args: []string{"--config=/tmp/icuvisor.json"},
		LoadConfig: func(_ context.Context, opts config.Options) (config.Config, error) {
			gotPath = opts.Path
			return config.Config{}, errors.New("stop")
		},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want loader error")
	}
	if gotPath != "/tmp/icuvisor.json" {
		t.Fatalf("config path = %q, want /tmp/icuvisor.json", gotPath)
	}
}

func TestRunUnknownCommandReturnsActionableError(t *testing.T) {
	t.Parallel()

	err := Run(context.Background(), Options{Args: []string{"bogus"}})
	if err == nil {
		t.Fatal("Run() error = nil, want unknown command error")
	}
	msg := err.Error()
	for _, want := range []string{"unknown command", "bogus", "icuvisor version"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not contain %q", msg, want)
		}
	}
}

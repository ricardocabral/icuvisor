package app

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type safeAppLogBuffer struct {
	mu sync.Mutex
	bytes.Buffer
}

func (b *safeAppLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.Write(p)
}

func (b *safeAppLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.String()
}

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
		Toolset:     safety.ToolsetFull,
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
	if gotInfo.Toolset != safety.ToolsetFull {
		t.Fatalf("server toolset = %q, want full", gotInfo.Toolset)
	}
}

func TestDefaultStartServerLogsStartupVersion(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo})))

	err := defaultStartServer(context.Background(), ServerInfo{Version: "v7.8.9", Toolset: safety.ToolsetFull})
	if err == nil {
		t.Fatal("defaultStartServer() error = nil, want config/client error")
	}
	out := logs.String()
	for _, want := range []string{"server starting", "version=v7.8.9", "resolved toolset", "toolset=full"} {
		if !strings.Contains(out, want) {
			t.Fatalf("startup log %q missing %q", out, want)
		}
	}
	if got := strings.Count(out, "resolved toolset"); got != 1 {
		t.Fatalf("resolved toolset log count = %d, want 1 in %q", got, out)
	}
	for _, forbidden := range []string{"get_activity_streams", "delete_event", "advanced_capabilities"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("startup log leaked tool name %q: %q", forbidden, out)
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

func TestRunDefaultPassesConfigFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want config.Options
	}{
		{
			name: "inline config",
			args: []string{"--config=/tmp/icuvisor.json"},
			want: config.Options{Path: "/tmp/icuvisor.json"},
		},
		{
			name: "separate config transport and bind",
			args: []string{"--config", "/tmp/icuvisor.json", "--transport", "http", "--http-bind", "127.0.0.1:9999"},
			want: config.Options{Path: "/tmp/icuvisor.json", Transport: "http", HTTPBindAddress: "127.0.0.1:9999"},
		},
		{
			name: "inline transport and bind",
			args: []string{"--transport=http", "--http-bind=192.168.1.20:8765"},
			want: config.Options{Transport: "http", HTTPBindAddress: "192.168.1.20:8765"},
		},
		{
			name: "separate env file",
			args: []string{"--env-file", "/tmp/icuvisor.env"},
			want: config.Options{DotEnvPath: "/tmp/icuvisor.env", DotEnvExplicit: true},
		},
		{
			name: "inline env file",
			args: []string{"--env-file=/tmp/icuvisor.env"},
			want: config.Options{DotEnvPath: "/tmp/icuvisor.env", DotEnvExplicit: true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got config.Options
			err := Run(context.Background(), Options{
				Args: tc.args,
				LoadConfig: func(_ context.Context, opts config.Options) (config.Config, error) {
					got = opts
					return config.Config{}, errors.New("stop")
				},
			})
			if err == nil {
				t.Fatal("Run() error = nil, want loader error")
			}
			if got.Path != tc.want.Path || got.Transport != tc.want.Transport || got.HTTPBindAddress != tc.want.HTTPBindAddress || got.DotEnvPath != tc.want.DotEnvPath || got.DotEnvExplicit != tc.want.DotEnvExplicit {
				t.Fatalf("config options = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestRunFlagErrorsAreActionable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "unknown", args: []string{"bogus"}, want: []string{"unknown command", "bogus", "icuvisor version"}},
		{name: "missing config", args: []string{"--config"}, want: []string{"missing value", "--config"}},
		{name: "empty transport", args: []string{"--transport="}, want: []string{"missing value", "--transport"}},
		{name: "missing bind", args: []string{"--http-bind", "--transport"}, want: []string{"missing value", "--http-bind"}},
		{name: "missing env file", args: []string{"--env-file"}, want: []string{"missing value", "--env-file"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := Run(context.Background(), Options{Args: tc.args})
			if err == nil {
				t.Fatal("Run() error = nil, want error")
			}
			msg := err.Error()
			for _, want := range tc.want {
				if !strings.Contains(msg, want) {
					t.Fatalf("error %q does not contain %q", msg, want)
				}
			}
		})
	}
}

func TestDefaultStartServerDispatchesHTTPTransport(t *testing.T) {
	logs := &safeAppLogBuffer{}
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- defaultStartServer(ctx, ServerInfo{Version: "v7.8.9", Config: config.Config{
			APIKey:          "secret",
			AthleteID:       "i12345",
			Timezone:        "UTC",
			APIBaseURL:      config.DefaultAPIBaseURL,
			HTTPTimeout:     30 * time.Second,
			Transport:       config.TransportHTTP,
			HTTPBindAddress: "127.0.0.1:0",
		}})
	}()
	deadline := time.After(time.Second)
	for !strings.Contains(logs.String(), "transport=streamable_http") {
		select {
		case <-deadline:
			cancel()
			t.Fatalf("startup log %q missing streamable_http transport", logs.String())
		case <-time.After(10 * time.Millisecond):
		}
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("defaultStartServer() error = %v, want context.Canceled", err)
	}
}

func TestDefaultStartServerWarnsForHTTPNonLoopbackBind(t *testing.T) {
	logs := &safeAppLogBuffer{}
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- defaultStartServer(ctx, ServerInfo{Version: "v7.8.9", Config: config.Config{
			APIKey:          "secret",
			AthleteID:       "i12345",
			Timezone:        "UTC",
			APIBaseURL:      config.DefaultAPIBaseURL,
			HTTPTimeout:     30 * time.Second,
			Transport:       config.TransportHTTP,
			HTTPBindAddress: "0.0.0.0:0",
		}})
	}()
	deadline := time.After(time.Second)
	for !strings.Contains(logs.String(), "non-loopback bind") {
		select {
		case <-deadline:
			cancel()
			t.Fatalf("startup log %q missing non-loopback bind warning", logs.String())
		case <-time.After(10 * time.Millisecond):
		}
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("defaultStartServer() error = %v, want context.Canceled", err)
	}
	out := logs.String()
	for _, want := range []string{"level=WARN", "non-loopback bind", "transport=http", "http_bind=0.0.0.0:0"} {
		if !strings.Contains(out, want) {
			t.Fatalf("startup log %q missing %q", out, want)
		}
	}
	for _, forbidden := range []string{"secret", "i12345"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("startup log leaked sensitive value %q: %q", forbidden, out)
		}
	}
}

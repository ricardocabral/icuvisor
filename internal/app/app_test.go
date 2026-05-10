package app

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRunVersionWritesInjectedVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := Run(context.Background(), Options{
		Version: "v1.2.3-test",
		Args:    []string{"version"},
		Stdout:  &stdout,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got, want := stdout.String(), "v1.2.3-test\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunDefaultDelegatesToStarterWithVersion(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("starter failed")
	var gotInfo ServerInfo
	err := Run(context.Background(), Options{
		Version: "v9.8.7",
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

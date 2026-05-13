package safety

import (
	"sync"
	"testing"
)

func TestParseModeDefaultsToSafe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  Mode
	}{
		{name: "empty", value: "", want: ModeSafe},
		{name: "safe", value: "safe", want: ModeSafe},
		{name: "safe uppercase", value: " SAFE ", want: ModeSafe},
		{name: "full", value: "full", want: ModeFull},
		{name: "full mixed case", value: " FuLl ", want: ModeFull},
		{name: "none", value: "none", want: ModeNone},
		{name: "unknown", value: "delete-everything", want: ModeSafe},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ParseMode(tc.value); got != tc.want {
				t.Fatalf("ParseMode(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestCapabilityMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode      Mode
		wantMode  string
		wantWrite bool
		wantDel   bool
	}{
		{mode: ModeSafe, wantMode: "safe", wantWrite: true, wantDel: false},
		{mode: ModeFull, wantMode: "full", wantWrite: true, wantDel: true},
		{mode: ModeNone, wantMode: "none", wantWrite: false, wantDel: false},
	}
	for _, tc := range tests {
		t.Run(tc.wantMode, func(t *testing.T) {
			t.Parallel()

			capability := NewCapability(tc.mode)
			if got := capability.Mode(); got != tc.wantMode {
				t.Fatalf("Mode() = %q, want %q", got, tc.wantMode)
			}
			if got := capability.CanWrite(); got != tc.wantWrite {
				t.Fatalf("CanWrite() = %v, want %v", got, tc.wantWrite)
			}
			if got := capability.CanDelete(); got != tc.wantDel {
				t.Fatalf("CanDelete() = %v, want %v", got, tc.wantDel)
			}
		})
	}
}

func TestNewCapabilityFromEnvReadsOnce(t *testing.T) {
	t.Parallel()

	value := "full"
	capability := NewCapabilityFromEnv(func(key string) string {
		if key != EnvDeleteMode {
			t.Fatalf("getenv key = %q, want %q", key, EnvDeleteMode)
		}
		return value
	})
	value = "none"

	if !capability.CanWrite() || !capability.CanDelete() || capability.Mode() != "full" {
		t.Fatalf("capability changed after env mutation: mode=%q write=%v delete=%v", capability.Mode(), capability.CanWrite(), capability.CanDelete())
	}
}

func TestCapabilityConcurrentReads(t *testing.T) {
	t.Parallel()

	capability := NewCapability(ModeFull)
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if capability.Mode() != "full" || !capability.CanWrite() || !capability.CanDelete() {
					t.Errorf("unexpected capability result: mode=%q write=%v delete=%v", capability.Mode(), capability.CanWrite(), capability.CanDelete())
					return
				}
			}
		}()
	}
	wg.Wait()
}

package safety

import "testing"

func TestParseToolsetDefaultsToCore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  Toolset
	}{
		{name: "empty", value: "", want: ToolsetCore},
		{name: "core", value: "core", want: ToolsetCore},
		{name: "core uppercase", value: " CORE ", want: ToolsetCore},
		{name: "full", value: "full", want: ToolsetFull},
		{name: "full mixed case", value: " FuLl ", want: ToolsetFull},
		{name: "unknown", value: "everything", want: ToolsetCore},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ParseToolset(tc.value); got != tc.want {
				t.Fatalf("ParseToolset(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestToolsetStringDefaultsToCore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   Toolset
		want string
	}{
		{name: "core", in: ToolsetCore, want: "core"},
		{name: "full", in: ToolsetFull, want: "full"},
		{name: "invalid", in: Toolset("surprise"), want: "core"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.in.String(); got != tc.want {
				t.Fatalf("Toolset.String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNewToolsetFromEnvReadsOnce(t *testing.T) {
	t.Parallel()

	value := "full"
	toolset := NewToolsetFromEnv(func(key string) string {
		if key != EnvToolset {
			t.Fatalf("getenv key = %q, want %q", key, EnvToolset)
		}
		return value
	})
	value = "core"

	if toolset != ToolsetFull || toolset.String() != "full" {
		t.Fatalf("toolset changed after env mutation: %q", toolset)
	}
}

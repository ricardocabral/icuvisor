package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestRunListAndDescribeUseDirectView(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		args  []string
		check func(*testing.T, []byte)
	}{
		{
			name: "list",
			args: []string{"list"},
			check: func(t *testing.T, output []byte) {
				t.Helper()
				if !bytes.Contains(output, []byte(`"name": "get_athlete_profile"`)) {
					t.Fatalf("list output missing core tool: %s", output)
				}
			},
		},
		{
			name: "describe",
			args: []string{"describe", "get_athlete_profile"},
			check: func(t *testing.T, output []byte) {
				t.Helper()
				var parsed struct {
					Name        string `json:"name"`
					InputSchema any    `json:"input_schema"`
				}
				if err := json.Unmarshal(output, &parsed); err != nil {
					t.Fatalf("unmarshal describe output: %v", err)
				}
				if parsed.Name != "get_athlete_profile" || parsed.InputSchema == nil {
					t.Fatalf("describe output = %#v", parsed)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			err := Run(context.Background(), Options{Args: tc.args, Stdout: &stdout, Version: "vtest", LoadConfig: testLoadConfig})
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			tc.check(t, stdout.Bytes())
		})
	}
}

func TestRunCallWritesStructuredJSON(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := Run(context.Background(), Options{
		Args:       []string{"call", "icuvisor_check_server_version"},
		Stdout:     &stdout,
		Version:    "vtest",
		LoadConfig: testLoadConfig,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"description_server_version": "vtest"`)) {
		t.Fatalf("call output = %s", stdout.Bytes())
	}
}

func TestParseCallArgsRejectsMultipleArgumentSources(t *testing.T) {
	t.Parallel()

	_, _, err := parseCallArgs([]string{"--args", `{}`, "--args-file", "request.json"})
	if err == nil || !strings.Contains(err.Error(), "only one") {
		t.Fatalf("parseCallArgs() error = %v, want mutually-exclusive argument-source error", err)
	}
}

func TestRunCallRejectsNonObjectArgumentsWithoutWritingStdout(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := Run(context.Background(), Options{Args: []string{"call", "get_athlete_profile", "--args", `[]`}, Stdout: &stdout, LoadConfig: testLoadConfig})
	if err == nil || !strings.Contains(err.Error(), "JSON object") {
		t.Fatalf("Run() error = %v, want JSON-object error", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func testLoadConfig(context.Context, config.Options) (config.Config, error) {
	return config.Config{
		APIKey:      "test-api-key",
		AthleteID:   "i12345",
		Timezone:    "UTC",
		APIBaseURL:  config.DefaultAPIBaseURL,
		HTTPTimeout: time.Second,
		DeleteMode:  safety.ModeSafe,
		Toolset:     safety.ToolsetCore,
	}, nil
}

package toolchecks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSchemaSnapshotsIncludesWriteTools(t *testing.T) {
	generated, err := GenerateSchemaSnapshots(t.Context())
	if err != nil {
		t.Fatalf("GenerateSchemaSnapshots() error = %v", err)
	}
	for _, name := range []string{"add_or_update_event", "link_activity_to_event", "add_activity_message"} {
		if _, ok := generated[name]; !ok {
			t.Fatalf("generated snapshots missing %s; write tools must be represented in schema catalog", name)
		}
	}
}

func TestGenerateSchemaSnapshotsUsesCallerContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := GenerateSchemaSnapshots(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GenerateSchemaSnapshots() error = %v, want context.Canceled", err)
	}
}

func TestCheckSnapshotFreshness(t *testing.T) {
	currentDir := t.TempDir()
	generated := map[string]Snapshot{
		"get_example": testSnapshot(t, "get_example", map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"activity_id": map[string]any{"type": "string", "description": "Activity ID."}}, "required": []string{"activity_id"}}),
	}
	writeTestSnapshot(t, currentDir, generated["get_example"])
	report, err := CheckSnapshotFreshness(currentDir, generated)
	if err != nil {
		t.Fatalf("CheckSnapshotFreshness() error = %v", err)
	}
	if !report.OK() {
		t.Fatalf("CheckSnapshotFreshness() failures = %#v, want pass", report.Failures)
	}
}

func TestCheckSchemaStability(t *testing.T) {
	baseSchema := map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"activity_id": map[string]any{"type": "string", "description": "Activity ID."}}, "required": []string{"activity_id"}}
	tests := []struct {
		name      string
		baseline  map[string]map[string]any
		generated map[string]map[string]any
		wantOK    bool
		wantKind  string
	}{
		{
			name:      "clean diff passes",
			baseline:  map[string]map[string]any{"get_example": baseSchema},
			generated: map[string]map[string]any{"get_example": baseSchema},
			wantOK:    true,
		},
		{
			name:     "added optional argument passes",
			baseline: map[string]map[string]any{"get_example": baseSchema},
			generated: map[string]map[string]any{"get_example": {"type": "object", "additionalProperties": false, "properties": map[string]any{
				"activity_id":  map[string]any{"type": "string", "description": "Activity ID."},
				"include_full": map[string]any{"type": "boolean", "default": false, "description": "Include raw payload."},
			}, "required": []string{"activity_id"}}},
			wantOK: true,
		},
		{
			name:     "description-only argument guidance change passes",
			baseline: map[string]map[string]any{"get_example": baseSchema},
			generated: map[string]map[string]any{"get_example": {"type": "object", "additionalProperties": false, "properties": map[string]any{
				"activity_id": map[string]any{"type": "string", "description": "Intervals.icu activity ID."},
			}, "required": []string{"activity_id"}}},
			wantOK: true,
		},
		{
			name:     "argument validation change fails",
			baseline: map[string]map[string]any{"get_example": baseSchema},
			generated: map[string]map[string]any{"get_example": {"type": "object", "additionalProperties": false, "properties": map[string]any{
				"activity_id": map[string]any{"type": "integer", "description": "Activity ID."},
			}, "required": []string{"activity_id"}}},
			wantOK:   false,
			wantKind: "property-changed",
		},
		{
			name:      "removed argument fails",
			baseline:  map[string]map[string]any{"get_example": baseSchema},
			generated: map[string]map[string]any{"get_example": {"type": "object", "additionalProperties": false, "properties": map[string]any{}, "required": []string{}}},
			wantOK:    false,
			wantKind:  "property-removed",
		},
		{
			name:     "renamed argument fails",
			baseline: map[string]map[string]any{"get_example": baseSchema},
			generated: map[string]map[string]any{"get_example": {"type": "object", "additionalProperties": false, "properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "Activity ID."},
			}, "required": []string{"id"}}},
			wantOK:   false,
			wantKind: "property-removed",
		},
		{
			name:      "new tool passes",
			baseline:  map[string]map[string]any{"get_example": baseSchema},
			generated: map[string]map[string]any{"get_example": baseSchema, "get_new_tool": {"type": "object", "additionalProperties": false, "properties": map[string]any{}}},
			wantOK:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baselineDir := t.TempDir()
			currentDir := t.TempDir()
			for name, schema := range tc.baseline {
				writeTestSnapshot(t, baselineDir, testSnapshot(t, name, schema))
			}
			generated := map[string]Snapshot{}
			for name, schema := range tc.generated {
				snapshot := testSnapshot(t, name, schema)
				generated[name] = snapshot
				writeTestSnapshot(t, currentDir, snapshot)
			}
			report, err := CheckSchemaStability(baselineDir, currentDir, generated)
			if err != nil {
				t.Fatalf("CheckSchemaStability() error = %v", err)
			}
			if report.OK() != tc.wantOK {
				t.Fatalf("CheckSchemaStability().OK() = %v, want %v; failures = %#v", report.OK(), tc.wantOK, report.Failures)
			}
			if tc.wantKind != "" && !hasFailureKind(report, tc.wantKind) {
				t.Fatalf("failures = %#v, want kind %q", report.Failures, tc.wantKind)
			}
		})
	}
}

func TestCheckSchemaStabilityMissingBaselineFails(t *testing.T) {
	_, err := CheckSchemaStability(filepath.Join(t.TempDir(), "missing"), t.TempDir(), map[string]Snapshot{})
	if err == nil || !strings.Contains(err.Error(), "baseline snapshot directory") {
		t.Fatalf("CheckSchemaStability() error = %v, want missing baseline error", err)
	}
}

func testSnapshot(t *testing.T, name string, schema map[string]any) Snapshot {
	t.Helper()
	raw, err := CanonicalJSON(schema)
	if err != nil {
		t.Fatalf("CanonicalJSON() error = %v", err)
	}
	return Snapshot{ToolName: name, Raw: raw, Schema: schema}
}

func writeTestSnapshot(t *testing.T, dir string, snapshot Snapshot) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, snapshot.ToolName+".json"), snapshot.Raw, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func hasFailureKind(report SchemaReport, kind string) bool {
	for _, failure := range report.Failures {
		if failure.Kind == kind {
			return true
		}
	}
	return false
}

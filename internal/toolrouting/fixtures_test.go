package toolrouting

import (
	"os"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
)

func TestLoadFixture(t *testing.T) {
	file, err := os.Open("testdata/cases.json")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer file.Close()

	fixture, err := LoadFixture(file, knownToolSet())
	if err != nil {
		t.Fatalf("LoadFixture() error = %v", err)
	}
	if len(fixture.Cases) < 8 {
		t.Fatalf("loaded %d cases, want at least 8", len(fixture.Cases))
	}
	seen := map[string]Case{}
	for _, c := range fixture.Cases {
		seen[c.ID] = c
	}
	if got := seen["activity-details-read"].ExpectedFirstTool; got == nil || *got != toolcatalog.GetActivityDetails {
		t.Fatalf("activity details expected tool = %v, want %s", got, toolcatalog.GetActivityDetails)
	}
	if got := seen["event-delete-hidden-in-safe-mode"].ExpectedFirstTool; got != nil {
		t.Fatalf("safe-mode delete expected tool = %v, want nil no-tool", *got)
	}
	if got := seen["event-delete-full-mode"].ExpectedFirstTool; got == nil || *got != toolcatalog.DeleteEvent {
		t.Fatalf("full-mode delete expected tool = %v, want %s", got, toolcatalog.DeleteEvent)
	}
}

func TestLoadFixtureRejectsUnknownTool(t *testing.T) {
	body := `{"version":1,"cases":[{"id":"bad","prompt":"do it","expected_first_tool":"missing_tool","catalog_mode":"core_safe","toolset":"core","delete_mode":"safe"}]}`
	_, err := LoadFixture(strings.NewReader(body), knownToolSet())
	if err == nil || !strings.Contains(err.Error(), `unknown tool "missing_tool"`) {
		t.Fatalf("LoadFixture() error = %v, want unknown tool", err)
	}
}

func TestLoadFixtureRejectsInconsistentCatalogMode(t *testing.T) {
	body := `{"version":1,"cases":[{"id":"bad","prompt":"do it","expected_first_tool":"get_activities","catalog_mode":"core_safe","toolset":"full","delete_mode":"safe"}]}`
	_, err := LoadFixture(strings.NewReader(body), knownToolSet())
	if err == nil || !strings.Contains(err.Error(), "does not match toolset/delete_mode") {
		t.Fatalf("LoadFixture() error = %v, want catalog mismatch", err)
	}
}

func TestCompareResult(t *testing.T) {
	tool := toolcatalog.GetActivities
	cases := []struct {
		name   string
		caseIn Case
		actual string
		want   Result
	}{
		{
			name:   "matching tool",
			caseIn: Case{ID: "read", ExpectedFirstTool: &tool},
			actual: toolcatalog.GetActivities,
			want:   Result{CaseID: "read", Expected: toolcatalog.GetActivities, Actual: toolcatalog.GetActivities, Pass: true},
		},
		{
			name:   "expected no tool",
			caseIn: Case{ID: "safe-delete"},
			actual: "",
			want:   Result{CaseID: "safe-delete", Expected: NoToolName, Actual: NoToolName, NoTool: true, Pass: true},
		},
		{
			name:   "mismatch",
			caseIn: Case{ID: "wrong", ExpectedFirstTool: &tool},
			actual: toolcatalog.GetEvents,
			want:   Result{CaseID: "wrong", Expected: toolcatalog.GetActivities, Actual: toolcatalog.GetEvents, Pass: false, Detail: "expected get_activities, got get_events"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CompareResult(tc.caseIn, tc.actual, "")
			if got != tc.want {
				t.Fatalf("CompareResult() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func knownToolSet() map[string]struct{} {
	tools := toolcatalog.AllToolNames()
	out := make(map[string]struct{}, len(tools))
	for _, name := range tools {
		out[name] = struct{}{}
	}
	return out
}

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
	if len(fixture.Cases) < 11 {
		t.Fatalf("loaded %d cases, want at least 11", len(fixture.Cases))
	}
	seen := map[string]Case{}
	for _, c := range fixture.Cases {
		seen[c.ID] = c
	}
	if got := seen["activity-details-read"].ExpectedFirstTool; got == nil || *got != toolcatalog.GetActivityDetails {
		t.Fatalf("activity details expected tool = %v, want %s", got, toolcatalog.GetActivityDetails)
	}
	for _, id := range []string{"activity-window-explicit-date-range", "activity-window-sport-filter-runs", "activity-window-oldest-in-period"} {
		if got := seen[id].ExpectedFirstTool; got == nil || *got != toolcatalog.GetActivities {
			t.Fatalf("%s expected tool = %v, want %s", id, got, toolcatalog.GetActivities)
		}
		if !strings.Contains(seen[id].Notes, "get_activity_details") || !strings.Contains(seen[id].Notes, "get_training_summary") {
			t.Fatalf("%s notes = %q, want disambiguation from details and summary tools", id, seen[id].Notes)
		}
	}
	if got := seen["event-delete-hidden-in-safe-mode"].ExpectedFirstTool; got == nil || *got != toolcatalog.ICUvisorListAdvancedCapabilities {
		t.Fatalf("safe-mode delete expected tool = %v, want %s", got, toolcatalog.ICUvisorListAdvancedCapabilities)
	}
	if got := seen["event-delete-full-mode"].ExpectedFirstTool; got == nil || *got != toolcatalog.DeleteEvent {
		t.Fatalf("full-mode delete expected tool = %v, want %s", got, toolcatalog.DeleteEvent)
	}
	for _, id := range []string{"race-a-event-create", "race-b-event-create", "race-c-event-create"} {
		if got := seen[id].ExpectedFirstTool; got == nil || *got != toolcatalog.AddOrUpdateEvent {
			t.Fatalf("%s expected tool = %v, want %s", id, got, toolcatalog.AddOrUpdateEvent)
		}
		if !strings.Contains(seen[id].Notes, "add_race_event") {
			t.Fatalf("%s notes = %q, want add_race_event negative fixture note", id, seen[id].Notes)
		}
	}
	if _, ok := knownToolSet()["add_race_event"]; ok {
		t.Fatal("known tool set unexpectedly includes add_race_event; race cases should route to add_or_update_event")
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

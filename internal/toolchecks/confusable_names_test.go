package toolchecks

import (
	"context"
	"errors"
	"testing"
)

func TestGenerateToolCatalogUsesCallerContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := GenerateToolCatalog(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GenerateToolCatalog() error = %v, want context.Canceled", err)
	}
}

func TestCheckConfusableCatalog(t *testing.T) {
	tests := []struct {
		name    string
		catalog []ToolInfo
		wantOK  bool
	}{
		{
			name: "confusable pair fails",
			catalog: []ToolInfo{
				{Name: "get_activity_details", Description: "Get activity data by activity_id."},
				{Name: "get_activity_streams", Description: "Get activity data by activity_id."},
			},
			wantOK: false,
		},
		{
			name: "rewritten pair passes",
			catalog: []ToolInfo{
				{Name: "get_activity_details", Description: "Get one activity's terse metadata and metrics by activity_id."},
				{Name: "get_activity_streams", Description: "Get canonical activity stream channels by activity_id."},
			},
			wantOK: true,
		},
		{
			name: "future event prefix participates",
			catalog: []ToolInfo{
				{Name: "get_event_by_id", Description: "Fetch calendar event data by event_id."},
				{Name: "get_event_messages", Description: "Fetch calendar event data by event_id."},
			},
			wantOK: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report := CheckConfusableCatalog(tc.catalog, DefaultConfusableThreshold)
			if report.OK() != tc.wantOK {
				t.Fatalf("CheckConfusableCatalog().OK() = %v, want %v; pairs = %#v", report.OK(), tc.wantOK, report.Pairs)
			}
		})
	}
}

func TestFirstDescriptionSentencePreservesDottedTokens(t *testing.T) {
	description := "List intervals.icu calendar events for a bounded date range. Returns terse rows."
	got := FirstDescriptionSentence(description)
	want := "List intervals.icu calendar events for a bounded date range."
	if got != want {
		t.Fatalf("FirstDescriptionSentence() = %q, want %q", got, want)
	}
}

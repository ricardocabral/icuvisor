package prompts

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFuelingReviewPromptValidatesModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		arguments map[string]string
		wantError string
		wantText  string
	}{
		{name: "default completed range", wantText: "resolve athlete-local offsets -14 and -1"},
		{name: "activity mode", arguments: map[string]string{"activity_id": "ride-123"}, wantText: "Scope: activity_id=ride-123"},
		{name: "date range mode", arguments: map[string]string{"start_date": "2026-05-01", "end_date": "2026-05-14"}, wantText: "Scope: start_date=2026-05-01, end_date=2026-05-14"},
		{name: "activity and date conflict", arguments: map[string]string{"activity_id": "ride-123", "start_date": "2026-05-01", "end_date": "2026-05-14"}, wantError: "activity_id cannot be combined"},
		{name: "one-sided range", arguments: map[string]string{"start_date": "2026-05-01"}, wantError: "provide both start_date and end_date"},
		{name: "date-time start", arguments: map[string]string{"start_date": "2026-05-01T07:00:00", "end_date": "2026-05-14"}, wantError: "invalid start_date; provide YYYY-MM-DD"},
		{name: "reversed range", arguments: map[string]string{"start_date": "2026-05-14", "end_date": "2026-05-01"}, wantError: "end_date must be on or after start_date"},
		{name: "over 90 days", arguments: map[string]string{"start_date": "2026-01-01", "end_date": "2026-04-01"}, wantError: "90 athlete-local days or fewer"},
		{name: "malformed race date", arguments: map[string]string{"race_date": "2026-06-07T00:00:00"}, wantError: "invalid race_date; provide YYYY-MM-DD"},
		{name: "race name requires date", arguments: map[string]string{"race_name": "A Race"}, wantError: "missing race_date; provide YYYY-MM-DD"},
		{name: "same-day race lookup", arguments: map[string]string{"race_date": "2026-06-07", "race_name": "A Race"}, wantText: "oldest/newest equal to that athlete-local date and limit:100"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := FuelingReviewPrompt().Handler(context.Background(), Request{Arguments: tc.arguments})
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("Handler() error = %v, want containing %q", err, tc.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(result.Messages) != 1 || !strings.Contains(result.Messages[0].Text, tc.wantText) {
				t.Fatalf("Handler() result = %#v, want text containing %q", result, tc.wantText)
			}
		})
	}
}

func TestFuelingReviewPromptHasReadOnlyProductBoundaries(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, FuelingReviewPrompt(), nil)
	for _, want := range []string{
		"read-only: do not call write/delete tools, include_full, streams, or raw payloads",
		"food/product library",
		"carbohydrate/calorie/sodium/fluid/sweat-rate targets",
		"qualified sports dietitian or clinician",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("fueling review prompt missing %q:\n%s", want, text)
		}
	}
}

func TestFuelingReviewPortablePackContract(t *testing.T) {
	t.Parallel()

	packBytes, err := os.ReadFile(filepath.Join("..", "..", "docs", "prompts", "client-prompt-packs", "fueling-review.md"))
	if err != nil {
		t.Fatalf("read fueling review pack: %v", err)
	}
	pack := string(packBytes)
	for _, want := range []string{
		"Registry prompt: `fueling_review`",
		"`activity_id` selects one activity and is mutually exclusive",
		"athlete-local YYYY-MM-DD dates, never date-times",
		"offsets -14 and -1",
		"get_athlete_profile",
		"resolve_calendar_dates",
		"get_activity_details",
		"get_activities",
		"include_unnamed: true",
		"next_page_token",
		"get_training_summary",
		"get_events",
		"limit: 100",
		"_meta.truncated",
		"fields: [\"kcalConsumed\", \"carbohydrates\", \"protein\", \"fatTotal\"]",
		"calories_intake, carbs_g, protein_g, and fat_g",
		"`carbs_ingested_g`",
		"`carbs_used_g`",
		"calories_burned",
		"exact code and call its meaning unknown",
		"Sourced activity evidence",
		"Sourced daily-wellness evidence",
		"Sourced race/calendar context",
		"Labelled calculations",
		"Coverage and data gaps",
		"General educational guidance",
		"`moving_time_seconds` as the only duration basis",
		"`logged carbs/hour = carbs_ingested_g / (moving_time_seconds / 3600)`",
		"non-negative numeric `carbs_ingested_g`",
		"Zero is eligible and produces `0 g/h`",
		"missing/zero/non-positive moving time",
		"Never use `carbs_used_g`, calories_burned, training load, wellness daily totals",
		"never call write or delete tools",
		"sodium, fluid, or sweat-rate targets",
		"qualified sports dietitian or clinician",
	} {
		if !strings.Contains(pack, want) {
			t.Fatalf("fueling review pack missing %q:\n%s", want, pack)
		}
	}
}

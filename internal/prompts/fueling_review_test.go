package prompts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

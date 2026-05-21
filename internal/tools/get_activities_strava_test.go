package tools

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	stravaSyncChainFixture      = "../intervals/testdata/activities/strava_sync_chain_empty_stubs.json"
	wantWahooStravaWorkaround   = "Open the intervals.icu Connections page, choose Wahoo, and click Download old data so historical activities are re-imported directly from Wahoo instead of through Strava's restricted API."
	wantUnknownStravaWorkaround = "Open the intervals.icu Connections page for the activity's original device provider and click Download old data so historical activities are re-imported directly from that provider instead of through Strava's restricted API."
)

func TestIsStravaBlockedGoldenBranches(t *testing.T) {
	t.Parallel()

	stringPtr := func(value string) *string { return &value }
	tests := []struct {
		name     string
		activity intervals.Activity
		want     bool
	}{
		{
			name: "Strava upstream marker present",
			activity: intervals.Activity{
				Source: stringPtr("Strava"),
				Raw:    map[string]any{"id": "strava-marker", "source": "Strava", "start_date_local": "2026-01-01T07:00:00"},
			},
			want: true,
		},
		{
			name: "N/A HR and cadence on power-only file is not a Strava stub",
			activity: intervals.Activity{
				Raw: map[string]any{
					"id":                 "power-only",
					"start_date_local":   "2026-01-01T07:00:00",
					"average_heartrate":  "N/A",
					"average_cadence":    "N/A",
					"weighted_average_w": 247,
				},
			},
			want: false,
		},
		{
			name: "manual entry with meaningful fields is not blocked",
			activity: intervals.Activity{
				Source: stringPtr("manual"),
				Raw:    map[string]any{"id": "manual", "source": "manual", "name": "Mobility", "type": "Other", "start_date_local": "2026-01-01T07:00:00"},
			},
			want: false,
		},
		{
			name:     "empty upstream object is treated as blocked stub",
			activity: intervals.Activity{Raw: map[string]any{}},
			want:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isStravaBlocked(tc.activity); got != tc.want {
				t.Fatalf("isStravaBlocked() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStravaBlockedWorkaroundInfersAllowedProviderFromExplicitEvidence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  map[string]any
		want string
	}{
		{name: "wahoo external id", raw: map[string]any{"external_id": "wahoo-synthetic-12345"}, want: wantWahooStravaWorkaround},
		{name: "unallowlisted sync chain", raw: map[string]any{"external_id": "mywhoosh-synthetic-23456", "source": "Strava"}, want: wantUnknownStravaWorkaround},
		{name: "missing provider evidence", raw: map[string]any{"source": "Strava", "_note": "hidden"}, want: wantUnknownStravaWorkaround},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := stravaBlockedWorkaround(tc.raw); got != tc.want {
				t.Fatalf("stravaBlockedWorkaround() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsStravaBlockedNumericSyncChainStubs(t *testing.T) {
	t.Parallel()

	activities := loadActivityFixtureFile(t, stravaSyncChainFixture)
	if len(activities) != 3 {
		t.Fatalf("fixture activities = %d, want Wahoo/MyWhoosh/TrainerRoad cases", len(activities))
	}
	for _, activity := range activities {
		t.Run(activity.ID, func(t *testing.T) {
			t.Parallel()
			if strings.HasPrefix(activity.ID, "i") {
				t.Fatalf("activity id = %q, want numeric/no-i stub id", activity.ID)
			}
			if !isStravaBlocked(activity) {
				t.Fatalf("isStravaBlocked(%#v) = false, want true for numeric sync-chain empty stub", activity.Raw)
			}
		})
	}
}

func loadActivityFixtureFile(t *testing.T, path string) []intervals.Activity {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read activity fixture %s: %v", path, err)
	}
	var activities []intervals.Activity
	if err := json.Unmarshal(data, &activities); err != nil {
		t.Fatalf("decode activity fixture %s: %v", path, err)
	}
	return activities
}

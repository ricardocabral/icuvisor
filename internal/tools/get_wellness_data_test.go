package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeWellnessClient struct {
	fakeProfileClient
	rows   []intervals.Wellness
	params intervals.WellnessParams
}

func (f *fakeWellnessClient) ListWellness(ctx context.Context, params intervals.WellnessParams) ([]intervals.Wellness, error) {
	f.params = params
	return f.rows, nil
}

func TestGetWellnessDataFixtures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		file   string
		assert func(t *testing.T, row map[string]any)
	}{
		{
			name: "polar fresh exactly 24h is not stale",
			file: "polar_fresh.json",
			assert: func(t *testing.T, row map[string]any) {
				native := nestedMap(t, row, "_native", "polar")
				if native["ans_charge"] != 6.5 || native["nightly_recharge_status"] != float64(5) {
					t.Fatalf("polar native = %+v", native)
				}
				prov := provenanceFor(t, row, "readiness")
				if prov["source"] != "polar" || prov["native_scale"] != "1-6 Polar nightly_recharge_status" || prov["fetched_at"] != "2026-05-10T00:00:00Z" {
					t.Fatalf("readiness provenance = %+v", prov)
				}
				if meta := row["_meta"].(map[string]any); meta["stale"] == true {
					t.Fatalf("exactly 24h bridge marked stale: %+v", meta)
				}
			},
		},
		{
			name: "polar stale older than 24h",
			file: "polar_stale.json",
			assert: func(t *testing.T, row map[string]any) {
				meta := row["_meta"].(map[string]any)
				if meta["stale"] != true || meta["stale_reason"] != "polar bridge data is older than 24h for this wellness date" {
					t.Fatalf("stale meta = %+v", meta)
				}
				prov := provenanceFor(t, row, "sleepScore")
				if prov["source"] != "polar" || prov["native_scale"] != "1-100 Polar sleep_score" {
					t.Fatalf("sleepScore provenance = %+v", prov)
				}
			},
		},
		{
			name: "garmin body battery native fields",
			file: "garmin_body_battery.json",
			assert: func(t *testing.T, row map[string]any) {
				native := nestedMap(t, row, "_native", "garmin")
				if native["body_battery_min"] != float64(32) || native["body_battery_max"] != float64(85) {
					t.Fatalf("garmin native = %+v", native)
				}
				prov := provenanceFor(t, row, "readiness")
				if prov["source"] != "garmin" {
					t.Fatalf("readiness provenance = %+v", prov)
				}
			},
		},
		{
			name: "oura raw sleep score",
			file: "oura_sleep_score.json",
			assert: func(t *testing.T, row map[string]any) {
				native := nestedMap(t, row, "_native", "oura")
				if native["sleep_score"] != float64(83) {
					t.Fatalf("oura native = %+v", native)
				}
				prov := provenanceFor(t, row, "sleepScore")
				if prov["source"] != "oura" || prov["native_scale"] != "0-100 Oura sleep score" {
					t.Fatalf("sleepScore provenance = %+v", prov)
				}
			},
		},
		{
			name: "manual only subjective scales without provenance",
			file: "manual_only.json",
			assert: func(t *testing.T, row map[string]any) {
				meta := row["_meta"].(map[string]any)
				if _, ok := meta["provenance"]; ok {
					t.Fatalf("manual-only row has provenance: %+v", meta)
				}
				scales := meta["scales"].(map[string]any)
				for _, field := range []string{"feel", "sleepQuality", "fatigue", "soreness", "stress", "mood", "motivation", "injury"} {
					if scales[field] == "" {
						t.Fatalf("scales missing %s: %+v", field, scales)
					}
				}
			},
		},
		{
			name: "custom null fields and unknown source provenance",
			file: "custom_fields.json",
			assert: func(t *testing.T, row map[string]any) {
				if _, ok := row["custom_null_metric"]; ok {
					t.Fatalf("custom null metric was not stripped: %+v", row)
				}
				meta := row["_meta"].(map[string]any)
				if !containsAny(meta["missing_fields"].([]any), "custom_null_metric") || !containsAny(meta["missing_fields"].([]any), "feel") || !containsAny(meta["missing_fields"].([]any), "sleepQuality") {
					t.Fatalf("missing_fields = %+v", meta["missing_fields"])
				}
				prov := provenanceFor(t, row, "sleepScore")
				if prov["source"] != "unknown" || prov["native_scale"] != "0-100 device nightly score" {
					t.Fatalf("unknown source provenance = %+v", prov)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			row := shapedFixtureRow(t, tt.file, false)
			tt.assert(t, row)
		})
	}
}

func TestGetWellnessDataNullStrippingAndIncludeFull(t *testing.T) {
	t.Parallel()

	defaultRow := shapedFixtureRow(t, "custom_fields.json", false)
	if _, ok := defaultRow["custom_null_metric"]; ok {
		t.Fatalf("default row kept custom_null_metric: %+v", defaultRow)
	}
	if defaultRow["custom_recovery_note"] != "late caffeine" {
		t.Fatalf("custom_recovery_note = %v", defaultRow["custom_recovery_note"])
	}

	fullRow := shapedFixtureRow(t, "custom_fields.json", true)
	if _, ok := fullRow["custom_null_metric"]; !ok {
		t.Fatalf("include_full row dropped custom_null_metric: %+v", fullRow)
	}
	fullMeta := fullRow["_meta"].(map[string]any)
	if _, ok := fullMeta["missing_fields"]; ok {
		t.Fatalf("include_full emitted missing_fields: %+v", fullMeta)
	}
	if _, ok := fullRow["full"].(map[string]any); !ok {
		t.Fatalf("include_full missing raw full payload: %+v", fullRow)
	}
}

func shapedFixtureRow(t *testing.T, fixture string, includeFull bool) map[string]any {
	t.Helper()
	row := loadWellnessFixture(t, fixture)
	client := &fakeWellnessClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", MeasurementPreference: "METRIC"}},
		rows:              []intervals.Wellness{row},
	}
	tool := newGetWellnessDataTool(client, client, "test", "UTC", false)
	args := `{"oldest":"2026-05-11","newest":"2026-05-15"}`
	if includeFull {
		args = `{"oldest":"2026-05-11","newest":"2026-05-15","include_full":true}`
	}
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(args)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	var shaped map[string]any
	if err := json.Unmarshal([]byte(resultText(t, result)), &shaped); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return shaped["wellness"].([]any)[0].(map[string]any)
}

func loadWellnessFixture(t *testing.T, fixture string) intervals.Wellness {
	t.Helper()
	path := filepath.Join("..", "intervals", "testdata", "wellness", fixture)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixture, err)
	}
	var row intervals.Wellness
	if err := json.Unmarshal(data, &row); err != nil {
		t.Fatalf("decode fixture %s: %v", fixture, err)
	}
	return row
}

func nestedMap(t *testing.T, row map[string]any, keys ...string) map[string]any {
	t.Helper()
	current := row
	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		if !ok {
			t.Fatalf("%s is %T, want map in row %+v", key, current[key], row)
		}
		current = next
	}
	return current
}

func provenanceFor(t *testing.T, row map[string]any, field string) map[string]any {
	t.Helper()
	return nestedMap(t, row, "_meta", "provenance", field)
}

func containsAny(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

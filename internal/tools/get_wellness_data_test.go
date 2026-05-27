package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
				scales := meta["scales"].(map[string]any)
				if scales["sleepScore"] != "0-100 (device-imported nightly score)" {
					t.Fatalf("canonical sleepScore scale = %+v", scales["sleepScore"])
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
				if prov["source"] != "garmin" || prov["native_scale"] != "0-100 Garmin Body Battery" {
					t.Fatalf("readiness provenance = %+v", prov)
				}
			},
		},
		{
			name: "whoop sleep performance and recovery native fields",
			file: "whoop_sleep_recovery.json",
			assert: func(t *testing.T, row map[string]any) {
				native := nestedMap(t, row, "_native", "whoop")
				if native["sleep_performance_percentage"] != float64(87) || native["recovery_score"] != float64(66) {
					t.Fatalf("whoop native = %+v", native)
				}
				sleep := provenanceFor(t, row, "sleepScore")
				if sleep["source"] != "whoop" || sleep["native_scale"] != "0-100 WHOOP sleep performance percentage" {
					t.Fatalf("sleepScore provenance = %+v", sleep)
				}
				readiness := provenanceFor(t, row, "readiness")
				if readiness["source"] != "whoop" || readiness["native_scale"] != "0-100 WHOOP recovery score" {
					t.Fatalf("readiness provenance = %+v", readiness)
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
			name: "manual only subjective scales and nutrition fields without provenance",
			file: "manual_only.json",
			assert: func(t *testing.T, row map[string]any) {
				for key, want := range map[string]float64{"calories_intake": 2400, "carbs_g": 320.5, "protein_g": 132.25, "fat_g": 78.75} {
					if row[key] != want {
						t.Fatalf("%s = %v, want %v in row %+v", key, row[key], want, row)
					}
				}
				for _, key := range []string{"kcalConsumed", "carbohydrates", "protein", "fatTotal", "calories", "calories_total"} {
					if _, ok := row[key]; ok {
						t.Fatalf("manual row exposed legacy/ambiguous %s: %+v", key, row)
					}
				}
				meta := row["_meta"].(map[string]any)
				if _, ok := meta["provenance"]; ok {
					t.Fatalf("manual-only row has provenance: %+v", meta)
				}
				semantics := meta["field_semantics"].(map[string]any)
				for _, key := range []string{"calories_intake", "carbs_g", "protein_g", "fat_g"} {
					if semantics[key] == "" {
						t.Fatalf("field_semantics missing %s: %+v", key, semantics)
					}
				}
				scales := meta["scales"].(map[string]any)
				for _, field := range []string{"feel", "sleepQuality", "fatigue", "soreness", "stress", "mood", "motivation"} {
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
				if prov["source"] != "unknown" || prov["native_scale"] != "unknown" {
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

func TestGetWellnessDataCurrentDayRangeAddsAsOfMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		rows:              []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-05-24","sleepQuality":3,"weight":null}`)},
	}
	tool := newGetWellnessDataToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-24","newest":"2026-05-24"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	meta := payload["_meta"].(map[string]any)
	assertSaoPauloAsOfMeta(t, meta)
	if meta["include_full"] != false || meta["oldest"] != "2026-05-24" || meta["newest"] != "2026-05-24" {
		t.Fatalf("wellness meta = %#v, want preserved range/include_full", meta)
	}
	row := payload["wellness"].([]any)[0].(map[string]any)
	if _, ok := row["weight"]; ok {
		t.Fatalf("wellness row kept null weight in terse mode: %#v", row)
	}
}

func TestGetWellnessDataPastRangeOmitsAsOfMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		rows:              []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-05-23","sleepQuality":3,"weight":null}`)},
	}
	tool := newGetWellnessDataToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-23"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	meta := payload["_meta"].(map[string]any)
	assertNoAsOfMeta(t, meta)
	if _, ok := meta["timezone"]; ok {
		t.Fatalf("timezone present without current-day as-of metadata: %#v", meta)
	}
	if meta["include_full"] != false || meta["oldest"] != "2026-05-01" || meta["newest"] != "2026-05-23" {
		t.Fatalf("wellness meta = %#v, want preserved range/include_full", meta)
	}
	row := payload["wellness"].([]any)[0].(map[string]any)
	if _, ok := row["weight"]; ok {
		t.Fatalf("wellness row kept null weight in terse mode: %#v", row)
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

	absentNutritionRow := shapedFixtureRow(t, "custom_fields.json", false)
	for _, key := range []string{"calories_intake", "carbs_g", "protein_g", "fat_g", "kcalConsumed", "carbohydrates", "protein", "fatTotal"} {
		if _, ok := absentNutritionRow[key]; ok {
			t.Fatalf("absent nutrition row emitted %s: %+v", key, absentNutritionRow)
		}
	}
}

func TestGetWellnessDataIncludeFullRetainsRawNutritionNames(t *testing.T) {
	t.Parallel()

	row := shapedFixtureRow(t, "manual_only.json", true)
	if row["calories_intake"] != float64(2400) || row["carbs_g"] != 320.5 || row["protein_g"] != 132.25 || row["fat_g"] != 78.75 {
		t.Fatalf("include_full row nutrition = %+v, want disambiguated keys", row)
	}
	for _, key := range []string{"kcalConsumed", "carbohydrates", "protein", "fatTotal"} {
		if _, ok := row[key]; ok {
			t.Fatalf("include_full row leaked top-level legacy %s: %+v", key, row)
		}
	}
	full := row["full"].(map[string]any)
	for key, want := range map[string]float64{"kcalConsumed": 2400, "carbohydrates": 320.5, "protein": 132.25, "fatTotal": 78.75} {
		if full[key] != want {
			t.Fatalf("full[%s] = %v, want %v in raw nutrition payload", key, full[key], want)
		}
	}
}

func TestGetWellnessDataHydrationSemanticsAndIncludeFull(t *testing.T) {
	t.Parallel()

	raw := `{"id":"2026-05-16","hydration":3,"hydrationVolume":1.75}`
	terse := shapedInlineWellnessRow(t, raw, false)
	if terse["hydration"] != float64(3) || terse["hydrationVolume"] != 1.75 {
		t.Fatalf("terse hydration row = %+v, want distinct hydration and hydrationVolume", terse)
	}
	for _, key := range []string{"hydration_ml", "hydration_liters", "hydration_volume"} {
		if _, ok := terse[key]; ok {
			t.Fatalf("terse hydration row emitted guessed/renamed %s: %+v", key, terse)
		}
	}
	semantics := terse["_meta"].(map[string]any)["field_semantics"].(map[string]any)
	if !strings.Contains(semantics["hydration"].(string), "unit semantics are not inferred") {
		t.Fatalf("hydration semantics = %#v, want no inferred unit", semantics)
	}
	if !strings.Contains(semantics["hydrationVolume"].(string), "preserved separately") {
		t.Fatalf("hydrationVolume semantics = %#v, want distinction", semantics)
	}
	if _, ok := terse["full"]; ok {
		t.Fatalf("terse hydration row included full payload: %+v", terse)
	}

	fullRow := shapedInlineWellnessRow(t, raw, true)
	if fullRow["hydration"] != float64(3) || fullRow["hydrationVolume"] != 1.75 {
		t.Fatalf("full hydration row = %+v, want distinct top-level fields", fullRow)
	}
	full := fullRow["full"].(map[string]any)
	if full["hydration"] != float64(3) || full["hydrationVolume"] != 1.75 {
		t.Fatalf("full raw hydration = %+v, want upstream names preserved under full", full)
	}
}

func TestGetWellnessDataNullHydrationDoesNotEmitSemantics(t *testing.T) {
	t.Parallel()

	row := shapedInlineWellnessRow(t, `{"id":"2026-05-17","hydration":null,"hydrationVolume":null}`, false)
	if _, ok := row["hydration"]; ok {
		t.Fatalf("null hydration was not stripped: %+v", row)
	}
	if _, ok := row["hydrationVolume"]; ok {
		t.Fatalf("null hydrationVolume was not stripped: %+v", row)
	}
	if meta, ok := row["_meta"].(map[string]any); ok {
		if semantics, ok := meta["field_semantics"].(map[string]any); ok {
			if semantics["hydration"] != nil || semantics["hydrationVolume"] != nil {
				t.Fatalf("null hydration left stale field_semantics: %+v", semantics)
			}
		}
	}
}

func shapedFixtureRow(t *testing.T, fixture string, includeFull bool) map[string]any {
	t.Helper()
	return shapedWellnessRow(t, loadWellnessFixture(t, fixture), includeFull)
}

func shapedInlineWellnessRow(t *testing.T, raw string, includeFull bool) map[string]any {
	t.Helper()
	var row intervals.Wellness
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		t.Fatalf("decode inline wellness row: %v", err)
	}
	return shapedWellnessRow(t, row, includeFull)
}

func shapedWellnessRow(t *testing.T, row intervals.Wellness, includeFull bool) map[string]any {
	t.Helper()
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

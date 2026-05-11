package streams

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestCanonicalKeyKnownVariants(t *testing.T) {
	tests := []struct {
		in    string
		want  string
		known bool
	}{
		{in: "groundContactTime", want: "ground_contact_time", known: true},
		{in: "GroundContactTime", want: "ground_contact_time", known: true},
		{in: "ground_contact_time", want: "ground_contact_time", known: true},
		{in: "heartRate", want: "heart_rate", known: true},
		{in: "heartrate", want: "heart_rate", known: true},
		{in: "Watts", want: "watts", known: true},
		{in: "CustomMetricValue", want: "custom_metric_value", known: false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, known := CanonicalKey(tt.in)
			if got != tt.want || known != tt.known {
				t.Fatalf("CanonicalKey(%q) = %q, %t; want %q, %t", tt.in, got, known, tt.want, tt.known)
			}
		})
	}
}

func TestCanonicalizeRowPreservesCollidingSampleArrays(t *testing.T) {
	power := make([]any, 2, 4)
	power[0], power[1] = float64(250), float64(260)
	watts := []any{float64(245), float64(255)}
	row := map[string]any{
		"Power": power,
		"Watts": watts,
	}

	got := CanonicalizeRow(row)
	wantSeries := []any{power, watts}
	if !reflect.DeepEqual(got["watts"], wantSeries) {
		t.Fatalf("canonical watts = %#v, want %#v", got["watts"], wantSeries)
	}
	if len(power) != 2 || cap(power) != 4 {
		t.Fatalf("input power slice mutated: len=%d cap=%d values=%#v", len(power), cap(power), power)
	}
	meta := got["_meta"].(map[string]any)
	collisions := meta["stream_key_collisions"].(map[string][]string)
	if !reflect.DeepEqual(collisions["watts"], []string{"Power", "Watts"}) {
		t.Fatalf("collisions = %#v, want Power/Watts", collisions)
	}
}

func TestCanonicalizeRowPreservesUnknownOriginalSpellings(t *testing.T) {
	got := CanonicalizeRow(map[string]any{
		"CustomMetricValue":   float64(42),
		"custom-metric-value": float64(43),
	})
	if !reflect.DeepEqual(got["custom_metric_value"], []any{float64(42), float64(43)}) {
		t.Fatalf("custom metric collision = %#v", got["custom_metric_value"])
	}
	meta := got["_meta"].(map[string]any)
	unknowns := meta["unknown_stream_keys"].([]string)
	if !reflect.DeepEqual(unknowns, []string{"CustomMetricValue", "custom-metric-value"}) {
		t.Fatalf("unknown stream keys = %#v", unknowns)
	}
}

func TestCanonicalizeRowsFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/casing_variants.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	got := CanonicalizeRows(rows)
	if len(got) != 2 {
		t.Fatalf("row count = %d, want 2", len(got))
	}
	for i, row := range got {
		for _, key := range []string{"time", "watts", "heart_rate", "ground_contact_time", "vertical_oscillation"} {
			if _, ok := row[key]; !ok {
				t.Fatalf("row %d missing canonical key %q: %#v", i, key, row)
			}
		}
		if _, ok := row["groundContactTime"]; ok {
			t.Fatalf("row %d retained non-canonical groundContactTime: %#v", i, row)
		}
	}
	unknowns0 := got[0]["_meta"].(map[string]any)["unknown_stream_keys"].([]string)
	if !reflect.DeepEqual(unknowns0, []string{"CustomMetricValue"}) {
		t.Fatalf("row 0 unknowns = %#v", unknowns0)
	}
	unknowns1 := got[1]["_meta"].(map[string]any)["unknown_stream_keys"].([]string)
	if !reflect.DeepEqual(unknowns1, []string{"custom-metric-value"}) {
		t.Fatalf("row 1 unknowns = %#v", unknowns1)
	}
}

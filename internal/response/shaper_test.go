package response

import (
	"encoding/json"
	"testing"
	"time"
)

func TestShapeNullStrippingPreservesNonNullZeroValues(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		want  map[string]any
	}{
		{
			name: "root scalar values",
			input: map[string]any{
				"zero":  float64(0),
				"empty": "",
				"false": false,
				"null":  nil,
			},
			want: map[string]any{
				"zero":  float64(0),
				"empty": "",
				"false": false,
				"_meta": map[string]any{
					"fields_present": []any{"empty", "false", "zero"},
					"missing_fields": []any{"null"},
					"server_version": "dev",
				},
			},
		},
		{
			name: "nested object values",
			input: map[string]any{
				"nested": map[string]any{"zero": float64(0), "empty": "", "false": false, "null": nil},
			},
			want: map[string]any{
				"nested": map[string]any{"zero": float64(0), "empty": "", "false": false},
				"_meta": map[string]any{
					"fields_present": []any{"nested"},
					"missing_fields": []any{"nested.null"},
					"server_version": "dev",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Shape(tt.input, Options{})
			if err != nil {
				t.Fatalf("Shape() error = %v", err)
			}
			assertJSONEqual(t, got, tt.want)
		})
	}
}

func TestShapeStripMetadataReportsPresentAndMissingFields(t *testing.T) {
	got, err := Shape(map[string]any{
		"b_present": true,
		"a_present": "",
		"nested":    map[string]any{"keep": float64(0), "drop": nil},
		"missing":   nil,
	}, Options{})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"a_present": "",
		"b_present": true,
		"nested":    map[string]any{"keep": float64(0)},
		"_meta": map[string]any{
			"fields_present": []any{"a_present", "b_present", "nested"},
			"missing_fields": []any{"missing", "nested.drop"},
			"server_version": "dev",
		},
	})
}

func TestShapePreservesNullArrayElements(t *testing.T) {
	input := map[string]any{
		"samples": []any{1, nil, 2, map[string]any{"hrv": nil, "zero": 0}},
	}
	got, err := Shape(input, Options{})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"samples": []any{float64(1), nil, float64(2), map[string]any{"zero": float64(0)}},
		"_meta": map[string]any{
			"fields_present": []any{"samples"},
			"missing_fields": []any{"samples[3].hrv"},
			"server_version": "dev",
		},
	})
}

func TestShapeRowCollectionsIndependently(t *testing.T) {
	input := map[string]any{
		"rows": []any{
			map[string]any{"date": "2026-05-11", "hrv": nil},
			map[string]any{"date": "2026-05-12", "hrv": 42},
		},
		"debug": nil,
		"_meta": map[string]any{"next_page_token": "next"},
	}
	got, err := Shape(input, Options{RowCollections: []string{"rows"}})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"rows": []any{
			map[string]any{
				"date": "2026-05-11",
				"_meta": map[string]any{
					"fields_present": []any{"date"},
					"missing_fields": []any{"hrv"},
				},
			},
			map[string]any{"date": "2026-05-12", "hrv": float64(42)},
		},
		"_meta": map[string]any{
			"next_page_token": "next",
			"fields_present":  []any{"rows"},
			"missing_fields":  []any{"debug"},
			"server_version":  "dev",
		},
	})
}

func TestShapeOwnsUnitMetadata(t *testing.T) {
	got, err := Shape(map[string]any{
		"distance": 5,
		"_meta":    map[string]any{"units": map[string]any{"system": "imperial"}, "source": "test"},
	}, Options{})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"distance": float64(5),
		"_meta": map[string]any{
			"source":         "test",
			"server_version": "dev",
		},
	})
}

func TestShapeStripsCallerUnitMetadataFromRowCollections(t *testing.T) {
	got, err := Shape(map[string]any{
		"rows": []any{
			map[string]any{
				"distance": float64(5),
				"_meta": map[string]any{
					"source": "row",
					"units":  map[string]any{"system": "imperial", "distance": "mi"},
				},
			},
		},
		"_meta": map[string]any{"units": map[string]any{"system": "imperial", "distance": "mi"}},
	}, Options{RowCollections: []string{"rows"}, UnitSystem: UnitSystemMetric})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"rows": []any{
			map[string]any{
				"distance": float64(5),
				"_meta":    map[string]any{"source": "row"},
			},
		},
		"_meta": map[string]any{
			"server_version": "dev",
			"units":          map[string]any{"system": "metric", "distance": "km", "pace": "min/km", "speed": "km/h"},
		},
	})
}

func TestShapeAddsUnitMetadata(t *testing.T) {
	got, err := Shape(map[string]any{"distance_mi": 3.1}, Options{UnitSystem: UnitSystemImperial})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"distance_mi": float64(3.1),
		"_meta": map[string]any{
			"server_version": "dev",
			"units":          map[string]any{"system": "imperial", "distance": "mi", "pace": "min/mi", "speed": "mph"},
		},
	})
}

func TestShapeRequiresObjectWrapper(t *testing.T) {
	if _, err := Shape([]any{map[string]any{"name": "athlete"}}, Options{}); err == nil {
		t.Fatal("Shape() error = nil, want object wrapper error")
	}
	if _, err := Shape("athlete", Options{}); err == nil {
		t.Fatal("Shape() scalar error = nil, want object wrapper error")
	}
}

func TestRegisteredScaleLabelsReturnsRegistryCopy(t *testing.T) {
	labels := RegisteredScaleLabels()
	if labels["feel"] != "1-5 (athlete-reported feel)" || labels["sleepQuality"] != "1-4 (athlete-entered, 1=poor 4=great)" {
		t.Fatalf("registered scale labels = %+v", labels)
	}
	if _, ok := labels["injury"]; ok {
		t.Fatalf("injury should be free text, not a registered scale: %+v", labels)
	}
	labels["feel"] = "mutated"
	if RegisteredScaleLabels()["feel"] != "1-5 (athlete-reported feel)" {
		t.Fatal("RegisteredScaleLabels returned mutable registry state")
	}
}

func TestShapeDoesNotAddScalesForUnregisteredFields(t *testing.T) {
	got, err := Shape(map[string]any{"unknown_scale": 4}, Options{})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"unknown_scale": float64(4),
		"_meta":         map[string]any{"server_version": "dev"},
	})
}

func TestShapeAddsScalesForRegisteredFields(t *testing.T) {
	got, err := Shape(map[string]any{"fatigue": 2, "feel": 4, "injury": "left knee", "mood": 5, "motivation": 4, "name": "athlete", "sleepQuality": 3, "sleepScore": 87, "soreness": 2, "stress": 3}, Options{})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"fatigue":      float64(2),
		"feel":         float64(4),
		"injury":       "left knee",
		"mood":         float64(5),
		"motivation":   float64(4),
		"name":         "athlete",
		"sleepQuality": float64(3),
		"sleepScore":   float64(87),
		"soreness":     float64(2),
		"stress":       float64(3),
		"_meta": map[string]any{
			"scales": map[string]any{
				"fatigue":      "1-5 (athlete-reported fatigue)",
				"feel":         "1-5 (athlete-reported feel)",
				"mood":         "1-5 (athlete-reported mood)",
				"motivation":   "1-5 (athlete-reported motivation)",
				"sleepQuality": "1-4 (athlete-entered, 1=poor 4=great)",
				"sleepScore":   "0-100 (device-imported nightly score)",
				"soreness":     "1-5 (athlete-reported soreness)",
				"stress":       "1-5 (athlete-reported stress)",
			},
			"server_version": "dev",
		},
	})
}

func TestShapeRemovesStaleCallerSuppliedScales(t *testing.T) {
	got, err := Shape(map[string]any{
		"unknown_scale": 4,
		"_meta": map[string]any{
			"scales": map[string]any{"unknown_scale": "1-5"},
			"source": "test",
		},
	}, Options{})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"unknown_scale": float64(4),
		"_meta": map[string]any{
			"source":         "test",
			"server_version": "dev",
		},
	})
}

func TestShapeAddsDeleteModeMetadata(t *testing.T) {
	SetDeleteMode("full")
	t.Cleanup(func() { SetDeleteMode("safe") })

	got, err := Shape(map[string]any{"name": "athlete"}, Options{})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	meta := got.(map[string]any)["_meta"].(map[string]any)
	if meta["delete_mode"] != "full" {
		t.Fatalf("delete_mode = %v, want full", meta["delete_mode"])
	}
}

func TestShapeDebugMetadataGate(t *testing.T) {
	input := map[string]any{
		"name":       "athlete",
		"fetched_at": "2026-05-11T12:00:00Z",
		"query_type": "profile",
	}

	got, err := Shape(input, Options{ServerVersion: "v0.2.0"})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"name":  "athlete",
		"_meta": map[string]any{"server_version": "v0.2.0"},
	})

	got, err = Shape(input, Options{
		ServerVersion: "v0.2.0",
		DebugMetadata: true,
		QueryType:     "profile",
		FetchedAt:     time.Date(2026, 5, 11, 12, 0, 0, 0, time.FixedZone("BRT", -3*60*60)),
	})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"name":       "athlete",
		"fetched_at": "2026-05-11T15:00:00Z",
		"query_type": "profile",
		"_meta":      map[string]any{"server_version": "v0.2.0"},
	})
}

func TestShapeDebugNullsDoNotLeakThroughMissingFields(t *testing.T) {
	input := map[string]any{
		"name":       "athlete",
		"fetched_at": nil,
		"query_type": nil,
		"rows": []any{
			map[string]any{"date": "2026-05-11", "fetched_at": nil, "hrv": nil},
		},
	}
	got, err := Shape(input, Options{RowCollections: []string{"rows"}})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"name": "athlete",
		"rows": []any{
			map[string]any{
				"date": "2026-05-11",
				"_meta": map[string]any{
					"fields_present": []any{"date"},
					"missing_fields": []any{"hrv"},
				},
			},
		},
		"_meta": map[string]any{"server_version": "dev"},
	})
}

func TestDebugMetadataFromEnv(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "on", value: "true", want: true},
		{name: "off", value: "false", want: false},
		{name: "invalid", value: "yes", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(EnvDebugMetadata, tt.value)
			if got := DebugMetadataFromEnv(); got != tt.want {
				t.Fatalf("DebugMetadataFromEnv() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestParseDebugMetadata(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "true", in: "true", want: true},
		{name: "mixed case", in: " TRUE ", want: true},
		{name: "false", in: "false", want: false},
		{name: "invalid", in: "yes", want: false},
		{name: "empty", in: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseDebugMetadata(tt.in); got != tt.want {
				t.Fatalf("ParseDebugMetadata(%q) = %t, want %t", tt.in, got, tt.want)
			}
		})
	}
}

func TestShapeIncludeFullNullConvention(t *testing.T) {
	type row struct {
		Keep *float64 `json:"keep"`
		Omit *float64 `json:"omit,omitempty"`
	}
	got, err := Shape(row{}, Options{IncludeFull: true})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"keep":  nil,
		"_meta": map[string]any{"server_version": "dev"},
	})
}

func assertJSONEqual(t *testing.T, got any, want any) {
	t.Helper()
	want = withDefaultDeleteMode(want)
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal got: %v", err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal want: %v", err)
	}
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("JSON mismatch\ngot:  %s\nwant: %s", gotJSON, wantJSON)
	}
}

func withDefaultDeleteMode(value any) any {
	root, ok := value.(map[string]any)
	if !ok {
		return value
	}
	out := cloneExpectedMap(root)
	addDefaultDeleteModeToExpected(out)
	return out
}

func addDefaultDeleteModeToExpected(value any) {
	switch typed := value.(type) {
	case map[string]any:
		if meta, ok := typed["_meta"].(map[string]any); ok {
			if _, hasServerVersion := meta["server_version"]; hasServerVersion {
				meta["delete_mode"] = "safe"
			}
		}
		for _, item := range typed {
			addDefaultDeleteModeToExpected(item)
		}
	case []any:
		for _, item := range typed {
			addDefaultDeleteModeToExpected(item)
		}
	}
}

func cloneExpectedMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		switch typed := value.(type) {
		case map[string]any:
			out[key] = cloneExpectedMap(typed)
		case []any:
			items := make([]any, len(typed))
			for i, item := range typed {
				if itemMap, ok := item.(map[string]any); ok {
					items[i] = cloneExpectedMap(itemMap)
				} else {
					items[i] = item
				}
			}
			out[key] = items
		default:
			out[key] = value
		}
	}
	return out
}

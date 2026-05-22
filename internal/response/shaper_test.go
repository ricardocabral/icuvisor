package response

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/safety"
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

func TestShapeAddsCatalogHashAndOwnsSchemaMetadata(t *testing.T) {
	resetRuntimeCatalogMetadataForTest()
	t.Cleanup(resetRuntimeCatalogMetadataForTest)
	setRuntimeCatalogMetadataForTest("v0.5.0", "current-hash")

	got, err := Shape(map[string]any{
		"name": "athlete",
		"_meta": map[string]any{
			"catalog_hash":          "spoofed",
			"schema_changed":        true,
			"schema_change_message": "spoofed",
			"previous_version":      "v0.0.1",
			"current_version":       "v0.0.2",
			"previous_catalog_hash": "old-spoofed",
			"source":                "preserved",
		},
	}, Options{ServerVersion: "v0.5.0"})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	meta := got.(map[string]any)["_meta"].(map[string]any)
	if meta["catalog_hash"] != "current-hash" || meta["source"] != "preserved" {
		t.Fatalf("_meta = %#v, want authoritative catalog hash and preserved source", meta)
	}
	for _, key := range []string{"schema_changed", "schema_change_message", "previous_version", "current_version", "previous_catalog_hash"} {
		if _, ok := meta[key]; ok {
			t.Fatalf("_meta[%q] present in steady-state metadata: %#v", key, meta)
		}
	}
}

func TestShapeSchemaChangeMetadata(t *testing.T) {
	resetRuntimeCatalogMetadataForTest()
	t.Cleanup(resetRuntimeCatalogMetadataForTest)
	setRuntimeCatalogMetadataForTest("v0.4.1", "old-hash")
	if _, err := Shape(map[string]any{"name": "first"}, Options{ServerVersion: "v0.4.1"}); err != nil {
		t.Fatalf("Shape(first) error = %v", err)
	}
	setRuntimeCatalogMetadataForTest("v0.5.0", "new-hash")

	got, err := Shape(map[string]any{"name": "second"}, Options{ServerVersion: "v0.5.0"})
	if err != nil {
		t.Fatalf("Shape(second) error = %v", err)
	}
	meta := got.(map[string]any)["_meta"].(map[string]any)
	wantMessage := "icuvisor was upgraded from v0.4.1 to v0.5.0 since this conversation started; tool schemas may have changed. Open a new conversation to use the latest tools."
	if meta["schema_changed"] != true || meta["schema_change_message"] != wantMessage || meta["previous_version"] != "v0.4.1" || meta["current_version"] != "v0.5.0" || meta["previous_catalog_hash"] != "old-hash" || meta["catalog_hash"] != "new-hash" {
		t.Fatalf("_meta = %#v, want schema-change metadata", meta)
	}
	if got := schemaChangeMessage("v0.4.1", "v0.5.0"); got != wantMessage {
		t.Fatalf("schemaChangeMessage() = %q, want %q", got, wantMessage)
	}
}

func TestShapeAddsDeleteModeMetadata(t *testing.T) {
	got, err := Shape(map[string]any{"name": "athlete"}, Options{DeleteMode: safety.ModeFull})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	meta := got.(map[string]any)["_meta"].(map[string]any)
	if meta["delete_mode"] != "full" {
		t.Fatalf("delete_mode = %v, want full", meta["delete_mode"])
	}
}

func TestShapeUsesPerCallDeleteModeMetadata(t *testing.T) {
	full, err := Shape(map[string]any{"name": "athlete"}, Options{DeleteMode: safety.ModeFull})
	if err != nil {
		t.Fatalf("Shape(full) error = %v", err)
	}
	none, err := Shape(map[string]any{"name": "athlete"}, Options{DeleteMode: safety.ModeNone})
	if err != nil {
		t.Fatalf("Shape(none) error = %v", err)
	}
	fullMeta := full.(map[string]any)["_meta"].(map[string]any)
	noneMeta := none.(map[string]any)["_meta"].(map[string]any)
	if fullMeta["delete_mode"] != "full" || noneMeta["delete_mode"] != "none" {
		t.Fatalf("delete modes = full:%v none:%v, want full/none", fullMeta["delete_mode"], noneMeta["delete_mode"])
	}
}

func TestShapeAddsToolsetMetadata(t *testing.T) {
	tests := []struct {
		name string
		set  string
		want string
	}{
		{name: "default", set: "", want: "core"},
		{name: "full", set: "full", want: "full"},
		{name: "invalid", set: "surprise", want: "core"},
		{name: "empty", set: "   ", want: "core"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Shape(map[string]any{"name": "athlete"}, Options{Toolset: safety.ParseToolset(tc.set)})
			if err != nil {
				t.Fatalf("Shape() error = %v", err)
			}
			meta := got.(map[string]any)["_meta"].(map[string]any)
			if meta["toolset"] != tc.want {
				t.Fatalf("toolset = %v, want %s", meta["toolset"], tc.want)
			}
		})
	}
}

func TestShapeOverwritesStaleCallerToolsetMetadata(t *testing.T) {
	got, err := Shape(map[string]any{
		"name": "athlete",
		"_meta": map[string]any{
			"toolset": "core",
			"count":   3,
		},
	}, Options{Toolset: safety.ToolsetFull})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	meta := got.(map[string]any)["_meta"].(map[string]any)
	if meta["toolset"] != "full" || meta["count"] != float64(3) {
		t.Fatalf("_meta = %#v, want toolset full and preserved count", meta)
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

func TestShapeJSONConversionContract(t *testing.T) {
	t.Run("json tags omitempty and ignored fields", func(t *testing.T) {
		type input struct {
			Renamed    string   `json:"renamed"`
			Ignored    string   `json:"-"`
			Empty      string   `json:"empty,omitempty"`
			NilPointer *string  `json:"nil_pointer,omitempty"`
			EmptySlice []string `json:"empty_slice,omitempty"`
			KeepZero   int      `json:"keep_zero"`
		}
		got, err := Shape(input{Renamed: "kept", Ignored: "dropped"}, Options{})
		if err != nil {
			t.Fatalf("Shape() error = %v", err)
		}
		assertJSONEqual(t, got, map[string]any{"renamed": "kept", "keep_zero": float64(0), "_meta": map[string]any{"server_version": "dev"}})
	})

	t.Run("raw message decodes as json", func(t *testing.T) {
		type input struct {
			Raw json.RawMessage `json:"raw"`
		}
		got, err := Shape(input{Raw: json.RawMessage(`{"nested":null,"list":[1,null]}`)}, Options{})
		if err != nil {
			t.Fatalf("Shape() error = %v", err)
		}
		assertJSONEqual(t, got, map[string]any{"raw": map[string]any{"list": []any{float64(1), nil}}, "_meta": map[string]any{"fields_present": []any{"raw"}, "missing_fields": []any{"raw.nested"}, "server_version": "dev"}})
	})

	t.Run("text marshaler uses json representation", func(t *testing.T) {
		type input struct {
			Fetched time.Time `json:"fetched"`
		}
		fetched := time.Date(2026, 5, 15, 12, 30, 0, 0, time.UTC)
		got, err := Shape(input{Fetched: fetched}, Options{})
		if err != nil {
			t.Fatalf("Shape() error = %v", err)
		}
		assertJSONEqual(t, got, map[string]any{"fetched": "2026-05-15T12:30:00Z", "_meta": map[string]any{"server_version": "dev"}})
	})

	t.Run("non finite floats fail before shaping", func(t *testing.T) {
		type input struct {
			Value float64 `json:"value"`
		}
		for _, value := range []float64{math.NaN(), math.Inf(1)} {
			if _, err := Shape(input{Value: value}, Options{}); err == nil {
				t.Fatalf("Shape(%v) error = nil, want unsupported value", value)
			}
		}
	})

	t.Run("float32 preserves encoding json byte shape", func(t *testing.T) {
		type input struct {
			Value float32 `json:"value"`
		}
		got, err := Shape(input{Value: float32(1.0 / 3.0)}, Options{})
		if err != nil {
			t.Fatalf("Shape() error = %v", err)
		}
		gotJSON, err := json.Marshal(got)
		if err != nil {
			t.Fatalf("marshal shaped: %v", err)
		}
		if !bytes.Contains(gotJSON, []byte(`"value":0.33333334`)) {
			t.Fatalf("float32 JSON = %s, want encoding/json float32 precision", gotJSON)
		}
	})

	t.Run("json number preserves number semantics", func(t *testing.T) {
		type input struct {
			Value json.Number `json:"value"`
		}
		got, err := Shape(input{Value: json.Number("8.5")}, Options{})
		if err != nil {
			t.Fatalf("Shape(valid json.Number) error = %v", err)
		}
		gotJSON, err := json.Marshal(got)
		if err != nil {
			t.Fatalf("marshal shaped: %v", err)
		}
		if bytes.Contains(gotJSON, []byte(`"value":"8.5"`)) || !bytes.Contains(gotJSON, []byte(`"value":8.5`)) {
			t.Fatalf("json.Number JSON = %s, want numeric value", gotJSON)
		}
		if _, err := Shape(input{Value: json.Number("bad")}, Options{}); err == nil {
			t.Fatal("Shape(invalid json.Number) error = nil, want invalid number error")
		}
	})

	t.Run("duplicate json field names use encoding json dominance", func(t *testing.T) {
		inputType := reflect.StructOf([]reflect.StructField{
			{Name: "A", Type: reflect.TypeOf(0), Tag: `json:"x"`},
			{Name: "B", Type: reflect.TypeOf(0), Tag: `json:"x"`},
		})
		input := reflect.New(inputType).Elem()
		input.Field(0).SetInt(1)
		input.Field(1).SetInt(2)
		got, err := Shape(input.Interface(), Options{})
		if err != nil {
			t.Fatalf("Shape() error = %v", err)
		}
		assertJSONEqual(t, got, map[string]any{"_meta": map[string]any{"server_version": "dev"}})
	})

	t.Run("string tag option uses encoding json representation", func(t *testing.T) {
		type input struct {
			Count int `json:"count,string"`
		}
		got, err := Shape(input{Count: 7}, Options{})
		if err != nil {
			t.Fatalf("Shape() error = %v", err)
		}
		assertJSONEqual(t, got, map[string]any{"count": "7", "_meta": map[string]any{"server_version": "dev"}})
	})

	t.Run("cycles fail with wrapped json error", func(t *testing.T) {
		input := map[string]any{}
		input["self"] = input
		if _, err := Shape(input, Options{}); err == nil {
			t.Fatal("Shape(cycle) error = nil, want wrapped cycle error")
		}
	})
}

func TestShapeDoesNotMutateCallerOwnedInput(t *testing.T) {
	input := map[string]any{
		"name":       "Tempo",
		"fetched_at": "2026-05-15T12:00:00Z",
		"full":       map[string]any{"raw_null": nil, "samples": []any{float64(1), nil}},
	}
	got, err := Shape(input, Options{IncludeFull: true})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	if _, ok := got.(map[string]any)["fetched_at"]; ok {
		t.Fatalf("shaped output kept debug fetched_at: %#v", got)
	}
	if input["fetched_at"] != "2026-05-15T12:00:00Z" {
		t.Fatalf("Shape mutated caller fetched_at: %#v", input)
	}
	full := input["full"].(map[string]any)
	if _, ok := full["raw_null"]; !ok {
		t.Fatalf("Shape mutated caller full map: %#v", full)
	}
}

func TestShapeProvenanceDebugAndScaleWalkerContract(t *testing.T) {
	input := map[string]any{
		"feel":       float64(4),
		"fetched_at": "2026-05-15T12:00:00Z",
		"_metadata":  map[string]any{"sleepScore": float64(82)},
		"_meta": map[string]any{
			"scales": map[string]any{"feel": "stale"},
			"provenance": map[string]any{
				"feel": map[string]any{"source": "manual", "fetched_at": "2026-05-15T11:00:00Z", "query_type": "bridge"},
			},
		},
	}
	got, err := Shape(input, Options{})
	if err != nil {
		t.Fatalf("Shape() error = %v", err)
	}
	assertJSONEqual(t, got, map[string]any{
		"feel":      float64(4),
		"_metadata": map[string]any{"sleepScore": float64(82)},
		"_meta": map[string]any{
			"provenance": map[string]any{"feel": map[string]any{"source": "manual", "fetched_at": "2026-05-15T11:00:00Z", "query_type": "bridge"}},
			"scales": map[string]any{
				"feel":       "1-5 (athlete-reported feel)",
				"sleepScore": "0-100 (device-imported nightly score)",
			},
			"server_version": "dev",
		},
	})
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

func BenchmarkShapeLargeIncludeFull(b *testing.B) {
	payload := map[string]any{"activities": make([]any, 0, 250), "_meta": map[string]any{"include_full": true}}
	for i := 0; i < 250; i++ {
		payload["activities"] = append(payload["activities"].([]any), map[string]any{
			"activity_id":           i,
			"name":                  "Long include_full activity",
			"fetched_at":            "2026-05-15T12:00:00Z",
			"icu_training_load":     nil,
			"power_stream":          []any{180, 190, nil, 205, 210, 215},
			"heart_rate_stream":     []any{120, 125, nil, 130, 135, 140},
			"nested_debug_metadata": map[string]any{"query_type": "debug", "keep": true},
		})
	}
	opts := Options{IncludeFull: true, RowCollections: []string{"activities"}, ServerVersion: "bench"}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Shape(payload, opts); err != nil {
			b.Fatalf("Shape() error = %v", err)
		}
	}
}

func TestShapeGoldenSnapshots(t *testing.T) {
	t.Cleanup(resetRuntimeCatalogMetadataForTest)
	for _, tc := range goldenSnapshotCases() {
		t.Run(tc.name, func(t *testing.T) {
			resetRuntimeCatalogMetadataForTest()
			t.Cleanup(resetRuntimeCatalogMetadataForTest)
			setRuntimeCatalogMetadataForTest("v0.4.0", "golden-catalog-hash")
			got, err := Shape(tc.input, tc.opts)
			if err != nil {
				t.Fatalf("Shape() error = %v", err)
			}
			gotJSON := canonicalGoldenJSON(t, got)
			path := filepath.Join("testdata", tc.fixture)
			if os.Getenv("UPDATE_RESPONSE_GOLDENS") == "1" {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("create testdata dir: %v", err)
				}
				if err := os.WriteFile(path, gotJSON, 0o644); err != nil {
					t.Fatalf("write golden %s: %v", path, err)
				}
				return
			}
			wantJSON, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden %s: %v", path, err)
			}
			if !bytes.Equal(gotJSON, wantJSON) {
				t.Fatalf("golden mismatch for %s; run UPDATE_RESPONSE_GOLDENS=1 go test ./internal/response -run TestShapeGoldenSnapshots\n--- got ---\n%s--- want ---\n%s", tc.fixture, gotJSON, wantJSON)
			}
		})
	}
}

type goldenSnapshotCase struct {
	name    string
	fixture string
	input   any
	opts    Options
}

type snapshotActivitiesResponse struct {
	Activities []snapshotActivityRow  `json:"activities"`
	Meta       snapshotActivitiesMeta `json:"_meta"`
}

type snapshotActivityRow struct {
	ActivityID        string               `json:"activity_id,omitempty"`
	Name              string               `json:"name,omitempty"`
	Sport             string               `json:"sport,omitempty"`
	DistanceKM        *float64             `json:"distance_km,omitempty"`
	MovingTimeSeconds int                  `json:"moving_time_seconds,omitempty"`
	Full              map[string]any       `json:"full,omitempty"`
	Unavailable       *snapshotUnavailable `json:"unavailable,omitempty"`
}

type snapshotUnavailable struct {
	Reason     string `json:"reason"`
	Workaround string `json:"workaround"`
}

type snapshotActivitiesMeta struct {
	PageSize      int    `json:"page_size"`
	NextPageToken string `json:"next_page_token,omitempty"`
	MoreAvailable bool   `json:"more_available"`
	IncludeFull   bool   `json:"include_full"`
}

type snapshotFitnessResponse struct {
	Rows []snapshotFitnessRow `json:"fitness"`
	Meta snapshotFitnessMeta  `json:"_meta"`
}

type snapshotFitnessRow struct {
	Date string   `json:"date"`
	CTL  *float64 `json:"ctl,omitempty"`
	ATL  *float64 `json:"atl,omitempty"`
	TSB  *float64 `json:"tsb,omitempty"`
}

type snapshotFitnessMeta struct {
	ServerVersion string   `json:"server_version"`
	StartDate     string   `json:"start_date"`
	EndDate       string   `json:"end_date"`
	Timezone      string   `json:"timezone"`
	Count         int      `json:"count"`
	IncludeFull   bool     `json:"include_full"`
	SourceTools   []string `json:"source_tools,omitempty"`
}

func goldenSnapshotCases() []goldenSnapshotCase {
	baseOpts := func(includeFull bool, rowCollections ...string) Options {
		return Options{
			IncludeFull:    includeFull,
			RowCollections: rowCollections,
			ServerVersion:  "v0.4.0",
			FetchedAt:      time.Date(2026, 5, 15, 17, 45, 0, 0, time.UTC),
			UnitSystem:     UnitSystemMetric,
			DeleteMode:     safety.ModeSafe,
			Toolset:        safety.ToolsetCore,
		}
	}
	distanceKM := 42.2
	ctl := 51.2
	atl := 56.7
	tsb := -4.8
	return []goldenSnapshotCase{
		{
			name:    "get activities terse",
			fixture: "get_activities_terse.golden.json",
			input: snapshotActivitiesResponse{
				Activities: []snapshotActivityRow{
					{ActivityID: "a1", Name: "Tempo Ride", Sport: "Ride", DistanceKM: &distanceKM, MovingTimeSeconds: 5400},
					{ActivityID: "a2", Sport: "Run", MovingTimeSeconds: 1800, Unavailable: &snapshotUnavailable{Reason: "strava_blocked", Workaround: "connect device directly"}},
				},
				Meta: snapshotActivitiesMeta{PageSize: 2, MoreAvailable: false, IncludeFull: false},
			},
			opts: baseOpts(false, "activities"),
		},
		{
			name:    "get activities include full",
			fixture: "get_activities_full.golden.json",
			input: snapshotActivitiesResponse{
				Activities: []snapshotActivityRow{
					{ActivityID: "a1", Name: "Tempo Ride", Sport: "Ride", Full: map[string]any{"icu_training_load": nil, "power_stream": []any{220.0, nil, 235.0}}},
				},
				Meta: snapshotActivitiesMeta{PageSize: 1, MoreAvailable: false, IncludeFull: true},
			},
			opts: baseOpts(true, "activities"),
		},
		{
			name:    "get fitness",
			fixture: "get_fitness.golden.json",
			input: snapshotFitnessResponse{
				Rows: []snapshotFitnessRow{
					{Date: "2026-05-14", CTL: &ctl, TSB: &tsb},
					{Date: "2026-05-15", ATL: &atl},
				},
				Meta: snapshotFitnessMeta{ServerVersion: "stale", StartDate: "2026-05-14", EndDate: "2026-05-15", Timezone: "UTC", Count: 2, IncludeFull: false, SourceTools: []string{"get_fitness"}},
			},
			opts: baseOpts(false, "fitness"),
		},
		{
			name:    "get events wrapper",
			fixture: "get_events_wrapper.golden.json",
			input: map[string]any{
				"events": []any{
					map[string]any{"event_id": "e1", "name": "Endurance", "target_load": 75, "description": nil},
				},
				"workouts": []any{
					map[string]any{"workout_id": "w1", "name": "3x tempo", "steps": []any{map[string]any{"duration_seconds": 600, "target": nil}}},
				},
				"warnings": []any{nil, "library folder skipped"},
				"_meta":    map[string]any{"operation": "apply_training_plan", "events_created": 1},
			},
			opts: baseOpts(false, "events", "workouts"),
		},
		{
			name:    "wellness provenance",
			fixture: "wellness_provenance.golden.json",
			input: map[string]any{
				"wellness": []any{
					map[string]any{"date": "2026-05-15", "readiness": 72, "feel": 4, "fetched_at": "2026-05-15T09:00:00Z", "query_type": "wellness", "_meta": map[string]any{"provenance": map[string]any{"readiness": map[string]any{"source": "polar", "native_scale": "1-6", "fetched_at": "2026-05-15T08:30:00Z"}}}},
				},
				"_meta": map[string]any{"include_full": false, "stale": false},
			},
			opts: baseOpts(false, "wellness"),
		},
	}
}

func canonicalGoldenJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal golden: %v", err)
	}
	return append(data, '\n')
}

func assertJSONEqual(t *testing.T, got any, want any) {
	t.Helper()
	want = withDefaultCommonMeta(want)
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

func withDefaultCommonMeta(value any) any {
	root, ok := value.(map[string]any)
	if !ok {
		return value
	}
	out := cloneExpectedMap(root)
	addDefaultCommonMetaToExpected(out)
	return out
}

func addDefaultCommonMetaToExpected(value any) {
	switch typed := value.(type) {
	case map[string]any:
		if meta, ok := typed["_meta"].(map[string]any); ok {
			if _, hasServerVersion := meta["server_version"]; hasServerVersion {
				meta["catalog_hash"] = defaultCatalogHash
				meta["delete_mode"] = "safe"
				meta["toolset"] = "core"
			}
		}
		for _, item := range typed {
			addDefaultCommonMetaToExpected(item)
		}
	case []any:
		for _, item := range typed {
			addDefaultCommonMetaToExpected(item)
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

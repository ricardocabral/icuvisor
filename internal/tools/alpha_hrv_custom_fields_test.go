package tools

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestAlphaHRVCanonicalContractIsIntervalOnlyAndSourceLabelled(t *testing.T) {
	t.Parallel()

	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-alpha","type":"Run","hrv":61,"device_name":"synthetic sensor"}`)
	client.intervals = decodeExtendedMetricsIntervals(t, `{"id":"activity-alpha","analyzed":true,"icu_intervals":[{"id":"interval-alpha","label":"steady","average_dfa_a1":0.75}]}`)
	client.powerErr = intervals.ErrNotFound

	result, err := newGetExtendedMetricsTool(client, client, "test", "UTC", false).Handler(context.Background(), Request{
		Name:      getExtendedMetricsName,
		Arguments: json.RawMessage(`{"activity_id":"activity-alpha"}`),
	})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	metrics := payload["metrics"].(map[string]any)
	if _, ok := metrics["dfa_alpha1"]; ok {
		t.Fatalf("activity metrics exposed interval-only dfa_alpha1: %#v", metrics)
	}
	if _, ok := metrics["hrv"]; ok {
		t.Fatalf("activity metrics joined wellness HRV: %#v", metrics)
	}
	rows := payload["intervals"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["dfa_alpha1"] != 0.75 {
		t.Fatalf("intervals = %#v, want interval dfa_alpha1", rows)
	}

	meta := payload["_meta"].(map[string]any)
	provenance := meta["metric_provenance"].(map[string]any)["dfa_alpha1"].(map[string]any)
	for key, want := range map[string]any{
		"source_field":    "average_dfa_a1",
		"response_field":  "dfa_alpha1",
		"scope":           "interval",
		"unit":            "unitless",
		"source_endpoint": "GET /api/v1/activity/{id}/intervals",
		"availability":    "conditional",
	} {
		if provenance[key] != want {
			t.Fatalf("metric_provenance.dfa_alpha1[%s] = %#v, want %#v", key, provenance[key], want)
		}
	}
	if _, present := meta["data_availability"]; present {
		t.Fatalf("device_name alone produced availability diagnostic: %#v", meta["data_availability"])
	}
	units := meta["extended_metric_units"].(map[string]any)
	if units["dfa_alpha1"] != "unitless" {
		t.Fatalf("extended_metric_units.dfa_alpha1 = %#v, want unitless", units["dfa_alpha1"])
	}
	if strings.Contains(strings.ToLower(string(mustJSON(t, payload))), "readiness") {
		t.Fatal("extended metrics response made a readiness claim")
	}
}

func TestAlphaHRVDFAAvailabilityMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		rawValue      string
		wantValue     any
		wantReason    string
		wantInterval  bool
		includeFull   bool
		wantRawExists bool
	}{
		{name: "present", rawValue: `0`, wantValue: float64(0), wantInterval: true},
		{name: "null", rawValue: `null`, wantReason: "dfa_alpha1_null", includeFull: true, wantRawExists: true},
		{name: "absent", rawValue: "", wantReason: "dfa_alpha1_unverified_missing", includeFull: true, wantRawExists: false},
		{name: "wrong type", rawValue: `"0.75"`, wantReason: "dfa_alpha1_malformed", includeFull: true, wantRawExists: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newFakeExtendedMetricsClient(t)
			client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-dfa","type":"Run"}`)
			interval := `{"id":"interval-dfa","label":"only dfa"}`
			if tc.rawValue != "" {
				interval = `{"id":"interval-dfa","label":"only dfa","average_dfa_a1":` + tc.rawValue + `}`
			}
			client.intervals = decodeExtendedMetricsIntervals(t, `{"id":"activity-dfa","analyzed":true,"icu_intervals":[`+interval+`]}`)
			client.powerErr = intervals.ErrNotFound
			args := `{"activity_id":"activity-dfa"}`
			if tc.includeFull {
				args = `{"activity_id":"activity-dfa","include_full":true}`
			}
			result, err := newGetExtendedMetricsTool(client, client, "test", "UTC", false).Handler(context.Background(), Request{Name: getExtendedMetricsName, Arguments: json.RawMessage(args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			if tc.wantInterval {
				rows := payload["intervals"].([]any)
				if len(rows) != 1 || rows[0].(map[string]any)["dfa_alpha1"] != tc.wantValue {
					t.Fatalf("intervals = %#v, want dfa_alpha1=%#v", rows, tc.wantValue)
				}
			} else if _, ok := payload["intervals"]; ok {
				t.Fatalf("metricless interval was retained in terse shape: %#v", payload["intervals"])
			}
			meta := payload["_meta"].(map[string]any)
			availability := meta["data_availability"]
			if tc.wantReason == "" {
				if availability != nil {
					t.Fatalf("data_availability = %#v, want no DFA diagnostic for valid zero", availability)
				}
			} else {
				diagnostics := availability.([]any)
				if len(diagnostics) != 1 {
					t.Fatalf("data_availability = %#v, want one per-interval diagnostic", diagnostics)
				}
				diagnostic := diagnostics[0].(map[string]any)
				if diagnostic["reason"] != tc.wantReason || diagnostic["activity_id"] != "activity-dfa" || diagnostic["interval_id"] != "interval-dfa" {
					t.Fatalf("diagnostic = %#v, want reason/activity/interval", diagnostic)
				}
				if !slices.Equal(diagnostic["source_fields"].([]any), []any{"average_dfa_a1"}) || !slices.Equal(diagnostic["missing_fields"].([]any), []any{"intervals[].dfa_alpha1"}) {
					t.Fatalf("diagnostic fields = %#v, want source and missing fields", diagnostic)
				}
			}
			if tc.includeFull {
				full := payload["full"].(map[string]any)
				rawIntervals := full["intervals"].(map[string]any)["icu_intervals"].([]any)[0].(map[string]any)
				if _, present := rawIntervals["average_dfa_a1"]; present != tc.wantRawExists {
					t.Fatalf("full interval raw key presence = %t, want %t: %#v", present, tc.wantRawExists, rawIntervals)
				}
			}
		})
	}
}

func TestAlphaHRVOptionalAndRestrictedSourcesUseHonestAvailability(t *testing.T) {
	t.Parallel()

	t.Run("interval source unavailable", func(t *testing.T) {
		client := newFakeExtendedMetricsClient(t)
		client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-missing-intervals","type":"Run"}`)
		client.intervalsErr = intervals.ErrNotFound
		result, err := newGetExtendedMetricsTool(client, client, "test", "UTC", false).Handler(context.Background(), Request{Name: getExtendedMetricsName, Arguments: json.RawMessage(`{"activity_id":"activity-missing-intervals"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		meta := resultMap(t, result)["_meta"].(map[string]any)
		if !slices.Equal(meta["unavailable_sources"].([]any), []any{"intervals"}) {
			t.Fatalf("unavailable_sources = %#v", meta["unavailable_sources"])
		}
		diagnostic := meta["data_availability"].([]any)[0].(map[string]any)
		if diagnostic["reason"] != "dfa_alpha1_unverified_missing" || diagnostic["activity_id"] != "activity-missing-intervals" {
			t.Fatalf("diagnostic = %#v", diagnostic)
		}
		if _, hasIntervalID := diagnostic["interval_id"]; hasIntervalID {
			t.Fatalf("activity-level diagnostic has interval_id: %#v", diagnostic)
		}
	})

	t.Run("Strava restriction", func(t *testing.T) {
		client := newFakeExtendedMetricsClient(t)
		client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-strava","source":"Strava","_note":"restricted"}`)
		result, err := newGetExtendedMetricsTool(client, client, "test", "UTC", false).Handler(context.Background(), Request{Name: getExtendedMetricsName, Arguments: json.RawMessage(`{"activity_id":"activity-strava"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		meta := resultMap(t, result)["_meta"].(map[string]any)
		diagnostics := meta["data_availability"].([]any)
		if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["reason"] != "restricted_source" {
			t.Fatalf("data_availability = %#v, want only restricted_source", diagnostics)
		}
	})
}

func TestAlphaHRVCustomActivityFieldsAreExplicitAndSourceLabelled(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{`{"id":"activity-custom","name":"Run","type":"Run","start_date_local":"2026-01-03T07:00:00","alpha_hrv":0,"custom_bool":false,"custom_null":null,"custom_object":{"value":1}}`}, "metric")
	client.customItems = decodeCustomItems(t,
		`{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"alpha_hrv"}}`,
		`{"id":"c2","type":"ACTIVITY_FIELD","content":{"field":"custom_bool"}}`,
		`{"id":"c3","type":"ACTIVITY_FIELD","content":{"field":"custom_null"}}`,
		`{"id":"c4","type":"ACTIVITY_FIELD","content":{"field":"custom_object"}}`,
		`{"id":"c5","type":"ACTIVITY_FIELD","content":{"field":"custom_absent"}}`,
	)
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","custom_fields":["alpha_hrv","custom_bool","custom_null","custom_object","custom_absent"]}`)})
	if err != nil {
		t.Fatalf("get_activities error = %v", err)
	}
	payload := resultMap(t, result)
	row := payload["activities"].([]any)[0].(map[string]any)
	custom := row["custom_fields"].(map[string]any)
	if custom["alpha_hrv"] != float64(0) || custom["custom_bool"] != false {
		t.Fatalf("custom_fields = %#v, want numeric zero and boolean false", custom)
	}
	for _, code := range []string{"custom_null", "custom_object", "custom_absent"} {
		if _, ok := custom[code]; ok {
			t.Fatalf("malformed/missing custom field %q was emitted: %#v", code, custom)
		}
	}
	meta := payload["_meta"].(map[string]any)
	provenance := meta["custom_field_provenance"].(map[string]any)
	for _, code := range []string{"alpha_hrv", "custom_bool", "custom_null", "custom_object", "custom_absent"} {
		entry := provenance[code].(map[string]any)
		for key, want := range map[string]any{
			"source_field":   code,
			"response_field": "custom_fields." + code,
			"scope":          "activity",
			"source":         "intervals.icu activity custom field",
			"selection":      "explicit",
			"unit":           "not_provided",
			"scale":          "not_provided",
			"algorithm":      "not_provided",
			"physiology":     "not_provided",
			"source_device":  "not_provided",
		} {
			if entry[key] != want {
				t.Fatalf("provenance[%s][%s] = %#v, want %#v", code, key, entry[key], want)
			}
		}
	}
	found := map[string]bool{}
	for _, raw := range meta["data_availability"].([]any) {
		diagnostic := raw.(map[string]any)
		found[diagnostic["reason"].(string)+":"+diagnostic["activity_id"].(string)] = true
		if diagnostic["activity_id"] != "activity-custom" {
			t.Fatalf("diagnostic not associated with activity: %#v", diagnostic)
		}
		code := ""
		switch diagnostic["reason"] {
		case "custom_field_null":
			code = "custom_null"
		case "custom_field_malformed":
			code = "custom_object"
		case "custom_field_absent":
			code = "custom_absent"
		}
		if code == "" || !slices.Equal(diagnostic["requested"].([]any), []any{code}) || !slices.Equal(diagnostic["source_fields"].([]any), []any{code}) || !slices.Equal(diagnostic["missing_fields"].([]any), []any{"custom_fields." + code}) {
			t.Fatalf("diagnostic = %#v, want exact selected-code association for %q", diagnostic, code)
		}
	}
	for _, want := range []string{"custom_field_null:activity-custom", "custom_field_malformed:activity-custom", "custom_field_absent:activity-custom"} {
		if !found[want] {
			t.Fatalf("data_availability missing %s: %#v", want, meta["data_availability"])
		}
	}
	serialized := strings.ToLower(string(mustJSON(t, payload)))
	if strings.Contains(serialized, "alpha_hrv_readiness") || strings.Contains(serialized, "readiness") || strings.Contains(serialized, "threshold") || strings.Contains(serialized, "medical") {
		t.Fatal("custom field produced an AlphaHRV/readiness/threshold/medical conclusion")
	}
}

func TestAlphaHRVRestrictedActivityKeepsCustomProvenanceWithoutValueDiagnostic(t *testing.T) {
	t.Parallel()

	t.Run("activity list", func(t *testing.T) {
		client := newFakeActivitiesClient(t, []string{`{"id":"activity-strava-custom","name":"Imported","type":"Run","source":"Strava","_note":"restricted","alpha_hrv":0}`}, "metric")
		client.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"alpha_hrv"}}`)
		tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)
		result, err := tool.Handler(context.Background(), Request{Name: getActivitiesName, Arguments: json.RawMessage(`{"oldest":"2026-01-01","custom_fields":["alpha_hrv"]}`)})
		if err != nil {
			t.Fatalf("get_activities error = %v", err)
		}
		meta := resultMap(t, result)["_meta"].(map[string]any)
		if _, ok := meta["custom_field_provenance"].(map[string]any)["alpha_hrv"]; !ok {
			t.Fatalf("custom provenance missing on restricted list response: %#v", meta)
		}
		diagnostics := meta["data_availability"].([]any)
		if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["reason"] != "restricted_source" {
			t.Fatalf("list data_availability = %#v, want only restricted_source", diagnostics)
		}
	})

	t.Run("activity detail", func(t *testing.T) {
		activity := decodeActivityFixture(t, `{"id":"activity-strava-detail","icu_athlete_id":"i12345","name":"Imported","type":"Run","source":"Strava","_note":"restricted","alpha_hrv":0}`)
		client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity}
		client.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"alpha_hrv"}}`)
		tool := newGetActivityDetailsToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)
		result, err := tool.Handler(context.Background(), Request{Name: getActivityDetailsName, Arguments: json.RawMessage(`{"activity_id":"activity-strava-detail","custom_fields":["alpha_hrv"]}`)})
		if err != nil {
			t.Fatalf("get_activity_details error = %v", err)
		}
		meta := resultMap(t, result)["_meta"].(map[string]any)
		if _, ok := meta["custom_field_provenance"].(map[string]any)["alpha_hrv"]; !ok {
			t.Fatalf("custom provenance missing on restricted detail response: %#v", meta)
		}
		diagnostics := meta["data_availability"].([]any)
		if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["reason"] != "restricted_source" {
			t.Fatalf("detail data_availability = %#v, want only restricted_source", diagnostics)
		}
	})
}

func TestAlphaHRVCustomFieldDetailProvenanceAndRawEscapeHatch(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"activity-detail-custom","icu_athlete_id":"i12345","name":"Run","type":"Run","start_date_local":"2026-01-03T07:00:00","alpha_hrv":0,"custom_bool":false,"custom_null":null,"custom_object":{"value":1}}`)
	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity}
	client.customItems = decodeCustomItems(t,
		`{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"alpha_hrv"}}`,
		`{"id":"c2","type":"ACTIVITY_FIELD","content":{"field":"custom_bool"}}`,
		`{"id":"c3","type":"ACTIVITY_FIELD","content":{"field":"custom_null"}}`,
		`{"id":"c4","type":"ACTIVITY_FIELD","content":{"field":"custom_object"}}`,
	)
	tool := newGetActivityDetailsToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: getActivityDetailsName, Arguments: json.RawMessage(`{"activity_id":"activity-detail-custom","custom_fields":["alpha_hrv","custom_bool","custom_null","custom_object"],"include_full":true}`)})
	if err != nil {
		t.Fatalf("get_activity_details error = %v", err)
	}
	payload := resultMap(t, result)
	row := payload["activity"].(map[string]any)
	custom := row["custom_fields"].(map[string]any)
	if custom["alpha_hrv"] != float64(0) || custom["custom_bool"] != false {
		t.Fatalf("detail custom_fields = %#v, want zero and false", custom)
	}
	if _, ok := custom["custom_null"]; ok {
		t.Fatalf("null custom field was emitted: %#v", custom)
	}
	if _, ok := custom["custom_object"]; ok {
		t.Fatalf("non-scalar custom field was emitted: %#v", custom)
	}
	meta := payload["_meta"].(map[string]any)
	if _, ok := meta["custom_field_provenance"].(map[string]any)["alpha_hrv"]; !ok {
		t.Fatalf("detail custom provenance missing alpha_hrv: %#v", meta)
	}
	full := row["full"].(map[string]any)
	if full["alpha_hrv"] != float64(0) || full["custom_bool"] != false || full["custom_null"] != nil || full["custom_object"] == nil {
		t.Fatalf("include_full did not preserve selected raw evidence: %#v", full)
	}
}

func TestAlphaHRVCustomFieldDefaultDoesNotListDefinitions(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{`{"id":"activity-default","name":"Run","type":"Run","start_date_local":"2026-01-03T07:00:00","alpha_hrv":0}`}, "metric")
	client.customItemsErr = context.Canceled
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("default get_activities error = %v", err)
	}
	row := resultMap(t, result)["activities"].([]any)[0].(map[string]any)
	if _, ok := row["custom_fields"]; ok {
		t.Fatalf("default activity row exposed custom fields: %#v", row)
	}
	if client.customItemsCalls != 0 {
		t.Fatalf("ListCustomItems calls = %d, want 0 for default read", client.customItemsCalls)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal test value: %v", err)
	}
	return encoded
}

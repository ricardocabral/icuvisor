package tools

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeWellnessWriterClient struct {
	fakeProfileClient
	row   intervals.Wellness
	rows  []intervals.Wellness
	calls []intervals.WriteWellnessParams
	err   error
}

func (f *fakeWellnessWriterClient) UpdateWellness(ctx context.Context, params intervals.WriteWellnessParams) (intervals.Wellness, error) {
	f.calls = append(f.calls, params)
	if len(f.rows) > 0 {
		row := f.rows[0]
		f.rows = f.rows[1:]
		return row, f.err
	}
	return f.row, f.err
}

func TestUpdateWellnessSchemaDocumentsRangesUnitsAndReadOnlyFields(t *testing.T) {
	t.Parallel()

	tool := newUpdateWellnessTool(&fakeWellnessWriterClient{}, &fakeProfileClient{}, "test", "UTC", false)
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)

	feel := props["feel"].(map[string]any)
	if feel["minimum"] != 1 || feel["maximum"] != 5 || !strings.Contains(feel["description"].(string), "1-5") {
		t.Fatalf("feel schema = %#v, want legacy 1-5 scale description", feel)
	}
	for _, field := range []string{"fatigue", "mood", "motivation", "soreness", "stress"} {
		prop := props[field].(map[string]any)
		if prop["minimum"] != 1 || prop["maximum"] != 5 || !strings.Contains(prop["description"].(string), "1-5") {
			t.Fatalf("%s schema = %#v, want 1-5 scale description", field, prop)
		}
	}
	sleep := props["sleepQuality"].(map[string]any)
	if sleep["minimum"] != 1 || sleep["maximum"] != 4 || !strings.Contains(sleep["description"].(string), "1-4") {
		t.Fatalf("sleepQuality schema = %#v, want 1-4 scale", sleep)
	}
	weight := props["weight"].(map[string]any)
	if weight["minimum"] != 0 || !strings.Contains(weight["description"].(string), "preferred weight unit") || !strings.Contains(weight["description"].(string), "kg") {
		t.Fatalf("weight schema = %#v, want preferred-unit to kg boundary docs", weight)
	}
	injury := props["injury"].(map[string]any)
	if injury["type"] != "string" || strings.Contains(injury["description"].(string), "scale") {
		t.Fatalf("injury schema = %#v, want free-text non-scale field", injury)
	}
	for _, readOnly := range []string{"sleepScore", "_native"} {
		if _, ok := props[readOnly]; ok {
			t.Fatalf("schema exposes read-only field %s", readOnly)
		}
	}
	for _, field := range []string{"examples", "input_examples"} {
		for _, example := range schemaExamples(t, tool.InputSchema.(map[string]any)[field]) {
			if _, ok := example["feel"]; ok {
				t.Fatalf("schema %s advertises unsupported wellness write field feel: %#v", field, example)
			}
		}
	}
}

func TestUpdateWellnessRejectsUnsupportedFeelBeforeWrite(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-01","feel":4,"fatigue":2,"soreness":2,"stress":3,"mood":4,"motivation":4,"sleepQuality":3,"locked":true}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want unsupported feel error")
	}
	if message, ok := PublicErrorMessage(err); !ok || message != "field_not_writable: feel (not accepted by intervals.icu wellness write)" {
		t.Fatalf("PublicErrorMessage() = %q, %v; want unsupported feel message", message, ok)
	}
	if len(client.calls) != 0 {
		t.Fatalf("write calls = %#v, want none for unsupported feel", client.calls)
	}
	if len(result.Content) != 0 || result.StructuredContent != nil || result.IsError {
		t.Fatalf("result = %#v, want empty result on validation error", result)
	}
}

func TestUpdateWellnessRejectsOutOfRangeAndReadOnlyArgumentsBeforeWrite(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)
	for _, raw := range []string{
		`{"date":"2026-05-01","fatigue":6}`,
		`{"date":"2026-05-01","restingHR":-1}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("Handler(%s) error = %v, want ErrInvalidInput", raw, err)
		}
	}
	for _, raw := range []string{
		`{"date":"2026-05-01","sleepQuality":5}`,
		`{"date":"2026-05-01","bodyFat":-1}`,
		`{"date":"2026-05-01","spO2":200}`,
		`{"date":"2026-05-01","vo2max":-1}`,
		`{"date":"2026-05-01","abdomen":-1}`,
		`{"date":"2026-05-01","respiration":-1}`,
		`{"date":"2026-05-01","menstrualPhase":""}`,
		`{"date":"2026-05-01","sleepScore":88}`,
		`{"date":"2026-05-01","_native":{"polar":{"sleep_score":90}}}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
	readOnlyCases := map[string]string{
		`{"date":"2026-05-01","sleepScore":88}`:                        "field_not_writable: sleepScore (device-managed)",
		`{"date":"2026-05-01","_native":{"polar":{"sleep_score":90}}}`: "field_not_writable: _native (bridge-managed)",
	}
	for raw, want := range readOnlyCases {
		_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)})
		if message, ok := PublicErrorMessage(err); !ok || message != want {
			t.Fatalf("PublicErrorMessage(%s) = %q, %v; want %q", raw, message, ok, want)
		}
	}
	if len(client.calls) != 0 {
		t.Fatalf("write calls = %#v, want none after validation failures", client.calls)
	}
}

func TestUpdateWellnessNewFieldsRoundTripToParamsAndFieldsUpdated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       string
		field     string
		wantParam func(intervals.WriteWellnessParams) bool
	}{
		{name: "spO2", raw: `{"date":"2026-05-01","spO2":97}`, field: "spO2", wantParam: func(p intervals.WriteWellnessParams) bool { return p.SpO2 != nil && *p.SpO2 == 97 }},
		{name: "vo2max", raw: `{"date":"2026-05-01","vo2max":52.5}`, field: "vo2max", wantParam: func(p intervals.WriteWellnessParams) bool { return p.VO2Max != nil && *p.VO2Max == 52.5 }},
		{name: "abdomen", raw: `{"date":"2026-05-01","abdomen":81.2}`, field: "abdomen", wantParam: func(p intervals.WriteWellnessParams) bool { return p.Abdomen != nil && *p.Abdomen == 81.2 }},
		{name: "respiration", raw: `{"date":"2026-05-01","respiration":13.5}`, field: "respiration", wantParam: func(p intervals.WriteWellnessParams) bool { return p.Respiration != nil && *p.Respiration == 13.5 }},
		{name: "menstrualPhase", raw: `{"date":"2026-05-01","menstrualPhase":"luteal"}`, field: "menstrualPhase", wantParam: func(p intervals.WriteWellnessParams) bool {
			return p.MenstrualPhase != nil && *p.MenstrualPhase == "luteal"
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeWellnessWriterClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric", Timezone: "UTC"}},
				row:               decodeWellnessRow(t, `{"id":"2026-05-01"}`),
			}
			tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.raw)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(client.calls) != 1 || !tc.wantParam(client.calls[0]) {
				t.Fatalf("write calls = %#v, want %s update", client.calls, tc.field)
			}
			meta := resultMap(t, result)["_meta"].(map[string]any)
			if !containsAny(meta["fields_updated"].([]any), tc.field) {
				t.Fatalf("fields_updated = %#v, want %s", meta["fields_updated"], tc.field)
			}
		})
	}
}

func TestUpdateWellnessWritesAllNewFieldsTogether(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric", Timezone: "UTC"}},
		row:               decodeWellnessRow(t, `{"id":"2026-05-01"}`),
	}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-01","spO2":98,"vo2max":54.1,"abdomen":80.2,"respiration":12.5,"menstrualPhase":"follicular"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.SpO2 == nil || *call.SpO2 != 98 || call.VO2Max == nil || *call.VO2Max != 54.1 || call.Abdomen == nil || *call.Abdomen != 80.2 || call.Respiration == nil || *call.Respiration != 12.5 || call.MenstrualPhase == nil || *call.MenstrualPhase != "follicular" {
		t.Fatalf("write call = %#v, want all new fields", call)
	}
	fields := resultMap(t, result)["_meta"].(map[string]any)["fields_updated"].([]any)
	for _, field := range []string{"spO2", "vo2max", "abdomen", "respiration", "menstrualPhase"} {
		if !containsAny(fields, field) {
			t.Fatalf("fields_updated = %#v, want %s", fields, field)
		}
	}
}

func TestUpdateWellnessFatigueOnlyDoesNotZeroWeight(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric", Timezone: "UTC"}},
		row:               decodeWellnessRow(t, `{"id":"2026-05-01","fatigue":2,"weight":70.5}`),
	}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

	if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-01","fatigue":2}`)}); err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	if client.calls[0].Weight != nil {
		t.Fatalf("weight param = %#v, want nil for omitted weight", client.calls[0].Weight)
	}
}

func TestUpdateWellnessLockedFollowUpSurfacesLockStateInMeta(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric", Timezone: "UTC"}},
		rows: []intervals.Wellness{
			decodeWellnessRow(t, `{"id":"2026-05-01","locked":true}`),
			decodeWellnessRow(t, `{"id":"2026-05-01","locked":true,"fatigue":2}`),
		},
	}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

	if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-01","locked":true}`)}); err != nil {
		t.Fatalf("locked Handler() error = %v", err)
	}
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-01","fatigue":2}`)})
	if err != nil {
		t.Fatalf("follow-up Handler() error = %v", err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["locked"] != true {
		t.Fatalf("meta = %#v, want locked=true surfaced", meta)
	}
	if len(client.calls) != 2 || client.calls[0].Locked == nil || !*client.calls[0].Locked || client.calls[1].Fatigue == nil || *client.calls[1].Fatigue != 2 {
		t.Fatalf("calls = %#v, want locked then fatigue follow-up", client.calls)
	}
}

func TestUpdateWellnessResponseUsesWellnessReadShape(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric", Timezone: "UTC"}},
		row:               decodeWellnessRow(t, `{"id":"2026-05-01","feel":4,"sleepScore":88,"polar_sleep_score":90,"bridge_fetched_at":"2026-05-01T01:00:00Z","injury":"left knee","weight":null,"locked":false}`),
	}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-01","injury":"left knee"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["wellness"].(map[string]any)
	if row["injury"] != "left knee" || row["locked"] != false {
		t.Fatalf("wellness row = %#v, want injury text and locked=false preserved", row)
	}
	assertKeyAbsent(t, row, "weight")
	meta := row["_meta"].(map[string]any)
	scales := meta["scales"].(map[string]any)
	if scales["feel"] == "" || scales["sleepScore"] == "" || scales["injury"] != nil {
		t.Fatalf("scales = %#v, want feel/sleepScore only for these fields", scales)
	}
	if meta["delete_mode"] != "safe" {
		t.Fatalf("row meta = %#v, want delete_mode", meta)
	}
	if !containsAny(meta["missing_fields"].([]any), "weight") || !containsAny(meta["fields_present"].([]any), "injury") {
		t.Fatalf("row meta = %#v, want missing/present field metadata", meta)
	}
	prov := meta["provenance"].(map[string]any)["sleepScore"].(map[string]any)
	if prov["source"] != "polar" || prov["native_scale"] != "1-100 Polar sleep_score" {
		t.Fatalf("sleepScore provenance = %#v", prov)
	}
}

func TestUpdateWellnessIncludeFullPreservesRawNullKeys(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric", Timezone: "UTC"}},
		row:               decodeWellnessRow(t, `{"id":"2026-05-01","injury":"left knee","weight":null,"locked":false}`),
	}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-01","injury":"left knee","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["wellness"].(map[string]any)
	full := row["full"].(map[string]any)
	assertKeyPresentNil(t, full, "weight")
	if row["locked"] != false {
		t.Fatalf("wellness row = %#v, want locked=false preserved", row)
	}
	if _, ok := row["_meta"].(map[string]any); !ok {
		t.Fatalf("wellness row = %#v, want nested row metadata preserved", row)
	}
}

func TestUpdateWellnessRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
}

func TestUpdateWellnessConvertsPreferredWeightToUpstreamKilograms(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "imperial", WeightPrefLB: true, Timezone: "UTC"}},
		row:               decodeWellnessRow(t, `{"id":"2026-05-01","weight":70,"feel":4}`),
	}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-01","weight":154.323584,"fatigue":2}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].Weight == nil {
		t.Fatalf("write calls = %#v, want weight update", client.calls)
	}
	if math.Abs(*client.calls[0].Weight-70) > 0.000001 {
		t.Fatalf("upstream weight kg = %.9f, want 70", *client.calls[0].Weight)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["weight_input_unit"] != "lb" || meta["weight_upstream_unit"] != "kg" {
		t.Fatalf("meta = %#v, want lb input/kg upstream", meta)
	}
}

func schemaExamples(t *testing.T, raw any) []map[string]any {
	t.Helper()
	items, ok := raw.([]map[string]any)
	if ok {
		return items
	}
	generic, ok := raw.([]any)
	if !ok {
		t.Fatalf("schema examples = %#v, want list", raw)
	}
	examples := make([]map[string]any, 0, len(generic))
	for _, item := range generic {
		example, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("schema example = %#v, want object", item)
		}
		examples = append(examples, example)
	}
	return examples
}

func decodeWellnessRow(t *testing.T, raw string) intervals.Wellness {
	t.Helper()
	var row intervals.Wellness
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		t.Fatalf("decode wellness row: %v", err)
	}
	return row
}

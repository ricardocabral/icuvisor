package tools

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeWellnessWriterClient struct {
	fakeProfileClient
	row   intervals.Wellness
	calls []intervals.WriteWellnessParams
	err   error
}

func (f *fakeWellnessWriterClient) UpdateWellness(ctx context.Context, params intervals.WriteWellnessParams) (intervals.Wellness, error) {
	f.calls = append(f.calls, params)
	return f.row, f.err
}

func TestUpdateWellnessSchemaDocumentsRangesUnitsAndReadOnlyFields(t *testing.T) {
	t.Parallel()

	tool := newUpdateWellnessTool(&fakeWellnessWriterClient{}, &fakeProfileClient{}, "test", "UTC", false)
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)

	for _, field := range []string{"feel", "fatigue", "mood", "motivation", "soreness", "stress"} {
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
}

func TestUpdateWellnessRejectsOutOfRangeAndReadOnlyArgumentsBeforeWrite(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)
	for _, raw := range []string{
		`{"date":"2026-05-01","feel":6}`,
		`{"date":"2026-05-01","sleepQuality":5}`,
		`{"date":"2026-05-01","bodyFat":-1}`,
		`{"date":"2026-05-01","sleepScore":88}`,
		`{"date":"2026-05-01","_native":{"polar":{"sleep_score":90}}}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
	if len(client.calls) != 0 {
		t.Fatalf("write calls = %#v, want none after validation failures", client.calls)
	}
}

func TestUpdateWellnessConvertsPreferredWeightToUpstreamKilograms(t *testing.T) {
	t.Parallel()

	client := &fakeWellnessWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "imperial", WeightPrefLB: true, Timezone: "UTC"}},
		row:               decodeWellnessRow(t, `{"id":"2026-05-01","weight":70,"feel":4}`),
	}
	tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-01","weight":154.323584,"feel":4}`)})
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

func decodeWellnessRow(t *testing.T, raw string) intervals.Wellness {
	t.Helper()
	var row intervals.Wellness
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		t.Fatalf("decode wellness row: %v", err)
	}
	return row
}

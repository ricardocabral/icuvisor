package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestUpdateWellnessHydrationBoundariesAndResponseMetadata(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		hydration int
	}{
		{name: "minimum accepted", hydration: 1},
		{name: "maximum accepted", hydration: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeWellnessWriterClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric", Timezone: "UTC"}},
				row:               decodeWellnessRow(t, fmt.Sprintf(`{"id":"2026-05-01","hydration":%d,"hydrationVolume":2.5}`, tc.hydration)),
			}
			tool := newUpdateWellnessTool(client, client, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(fmt.Sprintf(`{"date":"2026-05-01","hydration":%d}`, tc.hydration))})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(client.calls) != 1 || client.calls[0].Hydration == nil || *client.calls[0].Hydration != tc.hydration {
				t.Fatalf("writer calls = %#v, want hydration=%d", client.calls, tc.hydration)
			}

			payload := resultMap(t, result)
			meta := payload["_meta"].(map[string]any)
			if fields := meta["fields_updated"].([]any); len(fields) != 1 || fields[0] != "hydration" {
				t.Fatalf("fields_updated = %#v, want [hydration]", fields)
			}
			wellness := payload["wellness"].(map[string]any)
			if wellness["hydration"] != float64(tc.hydration) || wellness["hydrationVolume"] != 2.5 {
				t.Fatalf("wellness = %#v, want separate hydration rating and volume", wellness)
			}
			if got := wellness["_meta"].(map[string]any)["scales"].(map[string]any)["hydration"]; got != "1-4 (athlete-reported hydration)" {
				t.Fatalf("hydration response scale = %#v", got)
			}
		})
	}
}

func TestUpdateWellnessRejectsInvalidHydrationAndVolumeBeforeWriter(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		raw  string
	}{
		{name: "below minimum", raw: `{"date":"2026-05-01","hydration":0}`},
		{name: "above maximum", raw: `{"date":"2026-05-01","hydration":5}`},
		{name: "volume is not writable", raw: `{"date":"2026-05-01","hydrationVolume":2.5}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeWellnessWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
			tool := newUpdateWellnessTool(client, client, "test", "UTC", false)
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.raw)})
			if err == nil {
				t.Fatal("Handler() error = nil, want validation error")
			}
			if len(client.calls) != 0 {
				t.Fatalf("writer calls = %#v, want none", client.calls)
			}
		})
	}
}

func TestUpdateWellnessHydrationSchemaAndExamples(t *testing.T) {
	t.Parallel()

	tool := newUpdateWellnessTool(&fakeWellnessWriterClient{}, &fakeProfileClient{}, "test", "UTC", false)
	schema := tool.InputSchema.(map[string]any)
	properties := schema["properties"].(map[string]any)
	hydration := properties["hydration"].(map[string]any)
	if hydration["minimum"] != 1 || hydration["maximum"] != 4 || hydration["description"] != "1-4 (athlete-reported hydration); hydration scale." {
		t.Fatalf("hydration schema = %#v", hydration)
	}
	if _, ok := properties["hydrationVolume"]; ok {
		t.Fatalf("schema exposes read-only hydrationVolume: %#v", properties["hydrationVolume"])
	}
	for _, field := range []string{"examples", "input_examples"} {
		for _, example := range schemaExamples(t, schema[field]) {
			if _, ok := example["hydrationVolume"]; ok {
				t.Fatalf("%s advertises hydrationVolume write: %#v", field, example)
			}
		}
	}
	if !strings.Contains(tool.Description, "hydration (1-4)") || !strings.Contains(tool.Description, "hydrationVolume is read-only") {
		t.Fatalf("tool description = %q", tool.Description)
	}
}

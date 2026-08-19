package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestUpdateActivityDoesNotAdvertiseGearAssignment(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{}
	tool := newUpdateActivityTool(client, client, "test", false)
	properties := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	if _, ok := properties["gear_id"]; ok {
		t.Fatal("update_activity schema advertises unsupported gear_id")
	}
	if !strings.Contains(strings.ToLower(tool.Description), "gear assignment is not supported") {
		t.Fatalf("description = %q, want unsupported gear assignment guidance", tool.Description)
	}
	outputDescription := strings.ToLower(tool.OutputSchema.(map[string]any)["description"].(string))
	if !strings.Contains(outputDescription, "gear assignment") || !strings.Contains(outputDescription, "unsupported") {
		t.Fatalf("output description = %q, want unsupported gear assignment guidance", outputDescription)
	}
}

func TestUpdateActivityRejectsUnsupportedGearAssignmentsBeforeWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		id   string
	}{
		{name: "known gear id", raw: `{"activity_id":"a1","gear_id":"123"}`, id: "123"},
		{name: "unknown gear id", raw: `{"activity_id":"a1","gear_id":"missing-gear"}`, id: "missing-gear"},
		{name: "empty gear id", raw: `{"activity_id":"a1","gear_id":""}`},
		{name: "null gear id", raw: `{"activity_id":"a1","gear_id":null}`},
		{name: "repeated assignment", raw: `{"activity_id":"a1","gear_id":"123","gear_id":"123"}`, id: "123"},
		{name: "valid name with assignment", raw: `{"activity_id":"a1","name":"Renamed","gear_id":"123"}`, id: "123"},
		{name: "valid description with assignment", raw: `{"activity_id":"a1","description":"Ride note","gear_id":"123"}`, id: "123"},
		{name: "valid carbs with assignment", raw: `{"activity_id":"a1","carbs_ingested_g":90,"gear_id":"123"}`, id: "123"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityUpdaterClient{}
			tool := newUpdateActivityTool(client, client, "test", false)
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.raw)})
			if err == nil {
				t.Fatalf("Handler(%s) error = nil, want unsupported gear assignment rejection", tc.raw)
			}
			publicMessage, ok := PublicErrorMessage(err)
			if !ok || publicMessage != invalidUpdateActivityArgumentsMessage {
				t.Fatalf("PublicErrorMessage = %q, %v; want stable unsupported-assignment message", publicMessage, ok)
			}
			if tc.id != "" && strings.Contains(publicMessage, tc.id) {
				t.Fatalf("public error = %q, must not include supplied gear ID %q", publicMessage, tc.id)
			}
			if len(client.calls) != 0 {
				t.Fatalf("updater calls = %#v, want zero before unsupported assignment is proven", client.calls)
			}
		})
	}
}

func TestUpdateActivityOmittedGearKeepsSparseMetadataWrite(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{
		activity: decodeActivity(t, `{"id":"a1","name":"Renamed"}`),
	}
	tool := newUpdateActivityTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","name":"Renamed"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || !client.calls[0].NameSet || client.calls[0].DescriptionSet || client.calls[0].CarbsIngestedSet {
		t.Fatalf("updater calls = %#v, want name-only sparse update with gear omitted", client.calls)
	}
	fields := resultMap(t, result)["fields_updated"].([]any)
	if len(fields) != 1 || fields[0] != "name" {
		t.Fatalf("fields_updated = %#v, want name only with omitted gear", fields)
	}
}

func TestUpdateActivityUnsupportedGearErrorDoesNotExposeRawPayload(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{}
	tool := newUpdateActivityTool(client, client, "test", false)
	raw := `{"activity_id":"a1","gear_id":"private-gear-id","unexpected":"raw upstream payload"}`
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)})
	publicMessage, ok := PublicErrorMessage(err)
	if !ok || strings.Contains(publicMessage, "private-gear-id") || strings.Contains(publicMessage, "raw upstream payload") {
		t.Fatalf("PublicErrorMessage = %q, %v; must be short and sanitized", publicMessage, ok)
	}
	if len(client.calls) != 0 {
		t.Fatalf("updater calls = %#v, want zero", client.calls)
	}
}

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeActivityUpdaterClient struct {
	fakeProfileClient
	activity intervals.Activity
	calls    []intervals.UpdateActivityParams
	err      error
}

func (f *fakeActivityUpdaterClient) UpdateActivity(ctx context.Context, params intervals.UpdateActivityParams) (intervals.Activity, error) {
	f.calls = append(f.calls, params)
	return f.activity, f.err
}

func decodeActivity(t *testing.T, raw string) intervals.Activity {
	t.Helper()
	var activity intervals.Activity
	if err := json.Unmarshal([]byte(raw), &activity); err != nil {
		t.Fatalf("decode activity: %v", err)
	}
	return activity
}

func TestUpdateActivitySuccessSparseFields(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}},
		activity:          decodeActivity(t, `{"id":"a1","name":"Threshold ride","description":"Held target W","extra":null}`),
	}
	tool := newUpdateActivityTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":" a1 ","name":" Threshold ride ","description":"Held target W","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("calls = %#v, want 1", client.calls)
	}
	call := client.calls[0]
	if call.ActivityID != "a1" || call.Name != "Threshold ride" || !call.NameSet || call.Description != "Held target W" || !call.DescriptionSet {
		t.Fatalf("call = %#v, want trimmed sparse update", call)
	}
	out := resultMap(t, result)
	if out["activity_id"] != "a1" || out["status"] != "updated" {
		t.Fatalf("response = %#v, want updated confirmation", out)
	}
	fields, ok := out["fields_updated"].([]any)
	if !ok || len(fields) != 2 || fields[0] != "description" || fields[1] != "name" {
		t.Fatalf("fields_updated = %#v, want [description name]", out["fields_updated"])
	}
	meta := out["_meta"].(map[string]any)
	if meta["athleteId"] != "i12345" || meta["destructive"] != false || meta["append_only"] != false || meta["source_endpoint"] != "/activity/{activityId}" {
		t.Fatalf("meta = %#v, want non-destructive metadata", meta)
	}
	full := out["full"].(map[string]any)
	assertKeyPresentNil(t, full, "extra")
}

func TestUpdateActivityClearsDescriptionWhenExplicitEmpty(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}},
		activity:          decodeActivity(t, `{"id":"a1"}`),
	}
	tool := newUpdateActivityTool(client, client, "test", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","description":""}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	call := client.calls[0]
	if call.NameSet || !call.DescriptionSet || call.Description != "" {
		t.Fatalf("call = %#v, want explicit description clear without name change", call)
	}
}

func TestUpdateActivityRejectsBadArguments(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{}
	tool := newUpdateActivityTool(client, client, "test", false)
	for _, raw := range []string{
		`{"activity_id":""}`,
		`{"activity_id":"a1"}`,
		`{"activity_id":"a1","name":""}`,
		`{"activity_id":"a1","name":"   "}`,
		`{"activity_id":"a1","name":"x","confirm":true}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestUpdateActivityPublicError(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{err: errors.New("upstream detail")}
	tool := newUpdateActivityTool(client, client, "test", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","name":"x"}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != updateActivityMessage {
		t.Fatalf("PublicErrorMessage = %q, %v; err = %v", message, ok, err)
	}
}

func TestUpdateActivityRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}}}
	tool := newUpdateActivityTool(client, client, "test", false)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	description := strings.ToLower(tool.Description)
	if !strings.Contains(description, "non-destructive") || strings.Contains(description, "confirm") {
		t.Fatalf("description = %q, want non-destructive language without confirm", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"activity_id", "name", "description", "include_full"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
	if _, ok := props["confirm"]; ok {
		t.Fatalf("schema includes forbidden confirm property")
	}
}

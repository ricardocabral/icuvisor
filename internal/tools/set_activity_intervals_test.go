package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestSetActivityIntervalsSerializesWorkoutDocAsDescription(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}},
		activity:          decodeActivity(t, `{"id":"a1","icu_intervals":[{"id":1,"name":"Warm up"}],"extra":null}`),
	}
	tool := newSetActivityIntervalsTool(client, client, "test", false)

	args := `{"activity_id":" a1 ","workout_doc":{"steps":[{"description":"Warm up","duration":600,"power":{"value":55,"units":"PERCENT_FTP"}}]},"include_full":true}`
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(args)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("calls = %#v, want 1", client.calls)
	}
	call := client.calls[0]
	if call.ActivityID != "a1" || !call.DescriptionSet || call.NameSet {
		t.Fatalf("call = %#v, want description-only sparse update", call)
	}
	if strings.TrimSpace(call.Description) == "" || !strings.Contains(call.Description, "Warm up") {
		t.Fatalf("description = %q, want serialized DSL containing Warm up", call.Description)
	}
	out := resultMap(t, result)
	if out["activity_id"] != "a1" || out["status"] != "intervals_set" || out["workout_doc_uploaded"] != "description_dsl" {
		t.Fatalf("response = %#v, want intervals_set confirmation", out)
	}
	meta := out["_meta"].(map[string]any)
	if meta["destructive"] != true || meta["interval_source_intent"] != "structured_workout" || meta["athleteId"] != "i12345" {
		t.Fatalf("meta = %#v, want destructive structured-intent meta", meta)
	}
	if _, ok := meta["workout_doc_warning"]; ok {
		t.Fatalf("meta includes unexpected workout_doc_warning when intervals.icu rendered intervals: %#v", meta)
	}
}

func TestSetActivityIntervalsMergesProseWhenProvided(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}},
		activity:          decodeActivity(t, `{"id":"a1","icu_intervals":[{"id":1}]}`),
	}
	tool := newSetActivityIntervalsTool(client, client, "test", false)

	args := `{"activity_id":"a1","workout_doc":{"steps":[{"description":"Warm up","duration":600,"power":{"value":55,"units":"PERCENT_FTP"}}]},"prose":"Felt strong throughout."}`
	if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(args)}); err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	call := client.calls[0]
	if !strings.Contains(call.Description, "Felt strong throughout.") {
		t.Fatalf("description = %q, want prose preserved verbatim", call.Description)
	}
	if !strings.Contains(call.Description, "Warm up") {
		t.Fatalf("description = %q, want serialized DSL preserved", call.Description)
	}
}

func TestSetActivityIntervalsWarnsWhenUpstreamDidNotRender(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}},
		activity:          decodeActivity(t, `{"id":"a1","icu_intervals":[]}`),
	}
	tool := newSetActivityIntervalsTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","workout_doc":{"steps":[{"description":"Warm up","duration":600,"power":{"value":55,"units":"PERCENT_FTP"}}]}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	warning, _ := meta["workout_doc_warning"].(string)
	if !strings.Contains(warning, "did not parse") {
		t.Fatalf("workout_doc_warning = %q, want unparsed-DSL warning", warning)
	}
}

func TestSetActivityIntervalsRejectsBadArguments(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{}
	tool := newSetActivityIntervalsTool(client, client, "test", false)
	for _, raw := range []string{
		`{"activity_id":"","workout_doc":{"steps":[{"duration":60}]}}`,
		`{"activity_id":"a1"}`,
		`{"activity_id":"a1","workout_doc":null}`,
		`{"activity_id":"a1","workout_doc":{"steps":[]}}`,
		`{"activity_id":"a1","workout_doc":{"steps":[{"duration":60}]},"confirm":true}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestSetActivityIntervalsPublicError(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{err: errors.New("upstream detail")}
	tool := newSetActivityIntervalsTool(client, client, "test", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","workout_doc":{"steps":[{"duration":60}]}}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != setActivityIntervalsMessage {
		t.Fatalf("PublicErrorMessage = %q, %v; err = %v", message, ok, err)
	}
}

func TestSetActivityIntervalsRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}}}
	tool := newSetActivityIntervalsTool(client, client, "test", false)
	if tool.Requirement != RequirementDelete {
		t.Fatalf("requirement = %q, want delete (registered only with ICUVISOR_DELETE_MODE=full)", tool.Requirement)
	}
	description := strings.ToLower(tool.Description)
	if !strings.Contains(description, "destructive") || !strings.Contains(description, "icuvisor_delete_mode=full") || strings.Contains(description, "confirm") {
		t.Fatalf("description = %q, want destructive language without confirm", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"activity_id", "workout_doc", "prose", "include_full"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
	if _, ok := props["confirm"]; ok {
		t.Fatalf("schema includes forbidden confirm property")
	}
}

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeActivityMessageWriterClient struct {
	fakeProfileClient
	message intervals.NewActivityMessage
	calls   []intervals.AddActivityMessageParams
	err     error
}

func (f *fakeActivityMessageWriterClient) AddActivityMessage(ctx context.Context, params intervals.AddActivityMessageParams) (intervals.NewActivityMessage, error) {
	f.calls = append(f.calls, params)
	return f.message, f.err
}

func TestAddActivityMessageSuccessAppendsFreeText(t *testing.T) {
	t.Parallel()

	client := &fakeActivityMessageWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345"}},
		message:           decodeNewActivityMessage(t, `{"id":42,"new_chat":{"id":7},"extra":null}`),
	}
	tool := newAddActivityMessageTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":" a1 ","message":"  keep body verbatim  ","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].ActivityID != "a1" || client.calls[0].Content != "  keep body verbatim  " {
		t.Fatalf("calls = %#v, want trimmed activity_id and verbatim content", client.calls)
	}
	out := resultMap(t, result)
	if out["activity_id"] != "a1" || out["message_id"] != float64(42) || out["status"] != "appended" {
		t.Fatalf("response = %#v, want append confirmation", out)
	}
	meta := out["_meta"].(map[string]any)
	if meta["append_only"] != true || meta["athlete_id"] != "i12345" || meta["include_full"] != true {
		t.Fatalf("meta = %#v, want append-only and normalized athlete ID", meta)
	}
	full := out["full"].(map[string]any)
	if full["extra"] != nil {
		t.Fatalf("full = %#v, want raw upstream response preserving null", full)
	}
}

func TestAddActivityMessageRejectsEmptyMessageAndBadArguments(t *testing.T) {
	t.Parallel()

	client := &fakeActivityMessageWriterClient{}
	tool := newAddActivityMessageTool(client, client, "test", false)
	for _, raw := range []string{
		`{"activity_id":"","message":"note"}`,
		`{"activity_id":"a1","message":""}`,
		`{"activity_id":"a1","message":"   "}`,
		`{"activity_id":"a1","message":"note","confirm":true}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestAddActivityMessagePublicError(t *testing.T) {
	t.Parallel()

	client := &fakeActivityMessageWriterClient{err: errors.New("upstream detail")}
	tool := newAddActivityMessageTool(client, client, "test", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","message":"note"}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != addActivityMessageMessage {
		t.Fatalf("PublicErrorMessage = %q, %v; err = %v", message, ok, err)
	}
}

func TestAddActivityMessageRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeActivityMessageWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345"}}}
	registrar := &collectingRegistrar{}
	if err := NewRegistry(client, "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	tool := findTool(t, registrar.tools, addActivityMessageName)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	description := strings.ToLower(tool.Description)
	if !strings.Contains(description, "non-destructive") || !strings.Contains(description, "never overwrites") || strings.Contains(description, "confirm") {
		t.Fatalf("description = %q, want append-only non-destructive language without confirm", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"activity_id", "message", "include_full"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
	if _, ok := props["confirm"]; ok {
		t.Fatalf("schema includes forbidden confirm property")
	}
}

func decodeNewActivityMessage(t *testing.T, raw string) intervals.NewActivityMessage {
	t.Helper()
	var message intervals.NewActivityMessage
	if err := json.Unmarshal([]byte(raw), &message); err != nil {
		t.Fatalf("decode activity message: %v", err)
	}
	return message
}

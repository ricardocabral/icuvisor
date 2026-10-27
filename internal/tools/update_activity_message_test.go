package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeActivityMessageEditor struct {
	params   intervals.ChangeActivityMessageParams
	remove   bool
	calls    int
	err      error
	before   intervals.ActivityMessage
	timezone string
}

func (f *fakeActivityMessageEditor) UpdateActivityMessage(_ context.Context, p intervals.ChangeActivityMessageParams) (intervals.ActivityMessageChange, error) {
	f.params = p
	f.calls++
	return f.result("updated"), f.err
}
func (f *fakeActivityMessageEditor) DeleteActivityMessage(_ context.Context, p intervals.ChangeActivityMessageParams) (intervals.ActivityMessageChange, error) {
	f.params = p
	f.calls++
	f.remove = true
	return f.result("deleted"), f.err
}
func (f *fakeActivityMessageEditor) result(status string) intervals.ActivityMessageChange {
	before := f.before
	if before.ID == 0 {
		before = intervals.ActivityMessage{ID: 11, Content: "original", Raw: map[string]any{"extra": nil}}
	}
	return intervals.ActivityMessageChange{Status: status, Before: before, Timezone: f.timezone}
}

func TestActivityMessageCorrectionRecoveryAndCapabilities(t *testing.T) {
	for _, remove := range []bool{false, true} {
		t.Run(map[bool]string{false: "update", true: "delete"}[remove], func(t *testing.T) {
			client := &fakeActivityMessageEditor{}
			tool := newUpdateActivityMessageTool(client, "test", "UTC", false)
			raw := `{"activity_id":" a1 ","message_id":11,"message":"fixed"}`
			if remove {
				tool = newDeleteActivityMessageTool(client, "test", "UTC", false)
				raw = `{"activity_id":" a1 ","message_id":11,"include_full":true}`
			}
			result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(raw)})
			if err != nil {
				t.Fatal(err)
			}
			out := resultMap(t, result)
			if client.params.ActivityID != "a1" || client.params.MessageID != 11 || client.remove != remove {
				t.Fatalf("call=%+v remove=%v", client.params, client.remove)
			}
			before := out["before"].(map[string]any)
			if before["content"] != "original" || before["id"] != float64(11) {
				t.Fatalf("recovery=%v", before)
			}
			if remove {
				if tool.Requirement != RequirementDelete {
					t.Fatalf("requirement=%s", tool.Requirement)
				}
				assertKeyPresentNil(t, before["full"].(map[string]any), "extra")
			} else {
				if tool.Requirement != RequirementWrite {
					t.Fatalf("requirement=%s", tool.Requirement)
				}
				if _, ok := before["full"]; ok {
					t.Fatalf("raw fields in terse result: %v", before)
				}
			}
		})
	}
}
func TestActivityMessageCorrectionRejectsInvalidArguments(t *testing.T) {
	for _, raw := range []string{`{}`, `{"activity_id":"a1","message_id":0,"message":"x"}`, `{"activity_id":"a1","message_id":11,"message":" "}`, `{"activity_id":"a1","message_id":11,"message":"x","chat_id":7}`, `{"activity_id":"a1","message_id":11,"message":"x","athlete_id":"i999"}`} {
		client := &fakeActivityMessageEditor{}
		tool := newUpdateActivityMessageTool(client, "test", "UTC", false)
		if _, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(raw)}); err == nil || client.calls != 0 {
			t.Fatalf("args=%s err=%v calls=%d", raw, err, client.calls)
		}
	}
}
func TestActivityMessageCorrectionHidesUpstreamErrors(t *testing.T) {
	client := &fakeActivityMessageEditor{err: errors.New("private upstream detail")}
	tool := newUpdateActivityMessageTool(client, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","message_id":11,"message":"x"}`)})
	if public, ok := PublicErrorMessage(err); !ok || public != changeActivityMessageError {
		t.Fatalf("public=%q err=%v", public, err)
	}
}

func TestDeleteActivityMessageRejectsEditFieldsEvenNull(t *testing.T) {
	for _, value := range []string{`"text"`, `null`} {
		client := &fakeActivityMessageEditor{}
		tool := newDeleteActivityMessageTool(client, "test", "UTC", false)
		_, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","message_id":11,"message":` + value + `}`)})
		if err == nil || client.calls != 0 {
			t.Fatalf("err=%v calls=%d", err, client.calls)
		}
	}
}

func TestActivityMessageRecoveryUsesAthleteTimezone(t *testing.T) {
	for _, remove := range []bool{false, true} {
		for _, profileTimezone := range []string{"America/Sao_Paulo", ""} {
			t.Run(map[bool]string{false: "edit", true: "delete"}[remove]+"/"+profileTimezone, func(t *testing.T) {
				client := &fakeActivityMessageEditor{timezone: profileTimezone, before: intervals.ActivityMessage{ID: 11, Created: "2026-10-21T10:00:00Z", DeletedAt: "2026-10-21T11:00:00Z", Raw: map[string]any{"created": "2026-10-21T10:00:00Z", "deleted": "2026-10-21T11:00:00Z"}}}
				fallback := "UTC"
				if profileTimezone == "" {
					fallback = "America/Sao_Paulo"
				}
				shape := responseShaping{deleteMode: "full", toolset: "full", catalogHash: "activity-timezone-test"}
				tool := newUpdateActivityMessageTool(client, "2.3.4", fallback, true, shape)
				raw := `{"activity_id":"a1","message_id":11,"message":"fixed","include_full":true}`
				if remove {
					tool = newDeleteActivityMessageTool(client, "2.3.4", fallback, true, shape)
					raw = `{"activity_id":"a1","message_id":11,"include_full":true}`
				}
				result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(raw)})
				if err != nil {
					t.Fatal(err)
				}
				payload := resultMap(t, result)
				if payload["_meta"].(map[string]any)["catalog_hash"] != "activity-timezone-test" {
					t.Fatalf("shaping metadata=%v", payload["_meta"])
				}
				if payload["_meta"].(map[string]any)["server_version"] != "2.3.4" {
					t.Fatalf("version metadata=%v", payload["_meta"])
				}
				before := payload["before"].(map[string]any)
				if before["created"] != "2026-10-21T07:00:00-03:00" || before["deleted_at"] != "2026-10-21T08:00:00-03:00" {
					t.Fatalf("recovery times=%v", before)
				}
				full := before["full"].(map[string]any)
				if full["created"] != "2026-10-21T10:00:00Z" || full["deleted"] != "2026-10-21T11:00:00Z" {
					t.Fatalf("original full times=%v", full)
				}
			})
		}
	}
}

func TestActivityMessageInputExamplesExecuteValidScopedCalls(t *testing.T) {
	for _, remove := range []bool{false, true} {
		client := &fakeActivityMessageEditor{}
		tool := newUpdateActivityMessageTool(client, "test", "UTC", false)
		if remove {
			tool = newDeleteActivityMessageTool(client, "test", "UTC", false)
		}
		examples := schemaInputExamples(t, tool.InputSchema.(map[string]any))
		if len(examples) == 0 {
			t.Fatal("missing input examples")
		}
		for _, example := range examples {
			roundTripped := jsonRoundTripExample(t, example)
			if err := validateExampleAgainstSchema(tool.InputSchema.(map[string]any), roundTripped, "$"); err != nil {
				t.Fatal(err)
			}
			assertNoRealAthleteData(t, roundTripped, tool.Name)
			raw, err := json.Marshal(example)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tool.Handler(context.Background(), Request{Arguments: raw}); err != nil {
				t.Fatalf("example=%s error=%v", raw, err)
			}
			if client.params.ActivityID == "" || client.params.MessageID <= 0 || (!remove && client.params.Content == "") {
				t.Fatalf("invalid example call=%+v", client.params)
			}
		}
	}
}

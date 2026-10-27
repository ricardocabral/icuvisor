package tools

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"testing"
)

type fakeActivityMessageEditor struct {
	params intervals.ChangeActivityMessageParams
	remove bool
	calls  int
	err    error
}

func (f *fakeActivityMessageEditor) UpdateActivityMessage(_ context.Context, p intervals.ChangeActivityMessageParams) (intervals.ActivityMessageChange, error) {
	f.params = p
	f.calls++
	return intervals.ActivityMessageChange{Status: "updated", Before: intervals.ActivityMessage{ID: 11, Content: "original", Raw: map[string]any{"extra": nil}}}, f.err
}
func (f *fakeActivityMessageEditor) DeleteActivityMessage(_ context.Context, p intervals.ChangeActivityMessageParams) (intervals.ActivityMessageChange, error) {
	f.params = p
	f.calls++
	f.remove = true
	return intervals.ActivityMessageChange{Status: "deleted", Before: intervals.ActivityMessage{ID: 11, Content: "original", Raw: map[string]any{"extra": nil}}}, f.err
}
func TestActivityMessageCorrectionRecoveryAndCapabilities(t *testing.T) {
	for _, remove := range []bool{false, true} {
		t.Run(map[bool]string{false: "update", true: "delete"}[remove], func(t *testing.T) {
			client := &fakeActivityMessageEditor{}
			tool := newUpdateActivityMessageTool(client, "test", false)
			raw := `{"activity_id":" a1 ","message_id":11,"message":"fixed"}`
			if remove {
				tool = newDeleteActivityMessageTool(client, "test", false)
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
		tool := newUpdateActivityMessageTool(client, "test", false)
		if _, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(raw)}); err == nil || client.calls != 0 {
			t.Fatalf("args=%s err=%v calls=%d", raw, err, client.calls)
		}
	}
}
func TestActivityMessageCorrectionHidesUpstreamErrors(t *testing.T) {
	client := &fakeActivityMessageEditor{err: errors.New("private upstream detail")}
	tool := newUpdateActivityMessageTool(client, "test", false)
	_, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","message_id":11,"message":"x"}`)})
	if public, ok := PublicErrorMessage(err); !ok || public != changeActivityMessageError {
		t.Fatalf("public=%q err=%v", public, err)
	}
}

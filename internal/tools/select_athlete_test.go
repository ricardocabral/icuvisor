package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
)

func TestSelectAthleteNormalizesAndRejectsUnauthorizedTargets(t *testing.T) {
	t.Parallel()

	cfg := coach.Config{DefaultAthleteID: "i123", Athletes: []coach.Athlete{
		{ID: "i123", AllowedTools: []string{toolcatalog.GetAthleteProfile}},
		{ID: "456", AllowedTools: []string{toolcatalog.GetPowerCurves}},
	}}
	tool := newSelectAthleteTool(cfg)
	store := coach.NewSelectionStore(cfg.DefaultAthleteID)
	ctx := coach.WithSelectionContext(context.Background(), coach.SelectionContext{Store: store, Key: "test-session", Scope: "session"})

	result, err := tool.Handler(ctx, Request{Name: tool.Name, Arguments: json.RawMessage(`{"athlete_id":" 456 "}`)})
	if err != nil {
		t.Fatalf("select_athlete numeric target error = %v", err)
	}
	var parsed selectAthleteResponse
	if err := json.Unmarshal([]byte(result.Content[0].Text), &parsed); err != nil {
		t.Fatalf("unmarshal select_athlete response: %v", err)
	}
	if parsed.PreviousAthleteID != "i123" || parsed.NewAthleteID != "456" {
		t.Fatalf("select_athlete response = %#v, want normalized numeric target", parsed)
	}

	_, err = tool.Handler(ctx, Request{Name: tool.Name, Arguments: json.RawMessage(`{"athlete_id":"not-an-id"}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != invalidCoachAthleteFormatMessage {
		t.Fatalf("invalid format public error = %q, %v; err=%v", message, ok, err)
	}

	_, err = tool.Handler(ctx, Request{Name: tool.Name, Arguments: json.RawMessage(`{"athlete_id":"i999"}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != unauthorizedCoachAthleteMessage {
		t.Fatalf("unauthorized target public error = %q, %v; err=%v", message, ok, err)
	}
}

func TestSelectAthleteRejectsCredentialLikeExtraFieldsWithoutChangingSelection(t *testing.T) {
	t.Parallel()

	cfg := coach.Config{DefaultAthleteID: "i123", Athletes: []coach.Athlete{
		{ID: "i123", AllowedTools: []string{toolcatalog.GetAthleteProfile}},
		{ID: "456", AllowedTools: []string{toolcatalog.GetPowerCurves}},
	}}
	tool := newSelectAthleteTool(cfg)
	store := coach.NewSelectionStore(cfg.DefaultAthleteID)
	ctx := coach.WithSelectionContext(context.Background(), coach.SelectionContext{Store: store, Key: "test-session", Scope: "session"})

	_, err := tool.Handler(ctx, Request{Name: tool.Name, Arguments: json.RawMessage(`{"athlete_id":"456","api_key":"secret"}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != invalidSelectAthleteArgumentsMessage {
		t.Fatalf("credential field public error = %q, %v; err=%v", message, ok, err)
	}
	if got := store.Selected("test-session"); got != "i123" {
		t.Fatalf("selected athlete after rejected credential field = %q, want default i123", got)
	}
}

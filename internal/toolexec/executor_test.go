package toolexec

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/tools"
)

func TestExecuteRejectsAthleteIDInLocalModeAsInvalidInput(t *testing.T) {
	t.Parallel()

	called := false
	tool := tools.Tool{
		Name: "get_athlete_profile",
		Handler: func(context.Context, tools.Request) (tools.Result, error) {
			called = true
			return tools.Result{}, nil
		},
	}

	out := Execute(context.Background(), tool, tools.Request{
		Name:      tool.Name,
		Arguments: json.RawMessage(`{"athlete_id":"i67890"}`),
	}, "i12345")
	if !errors.Is(out.Err, tools.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", out.Err)
	}
	if out.PublicMessage != LocalModeAthleteTargetMessage {
		t.Fatalf("public message = %q, want %q", out.PublicMessage, LocalModeAthleteTargetMessage)
	}
	if called {
		t.Fatal("handler was called")
	}
}

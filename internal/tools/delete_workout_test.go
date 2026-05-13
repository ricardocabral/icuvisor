package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeWorkoutDeleterClient struct {
	fakeProfileClient
	calls []string
	err   error
}

func (f *fakeWorkoutDeleterClient) DeleteLibraryWorkout(ctx context.Context, workoutID string) error {
	f.calls = append(f.calls, workoutID)
	return f.err
}

func TestDeleteWorkoutSuccess(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutDeleterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newDeleteWorkoutTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"workout_id":" w-1 "}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0] != "w-1" {
		t.Fatalf("delete calls = %#v, want trimmed workout ID", client.calls)
	}
	out := resultMap(t, result)
	deleted := out["deleted"].(map[string]any)
	if deleted["workout_id"] != "w-1" || deleted["status"] != "deleted" {
		t.Fatalf("deleted = %#v, want delete confirmation", deleted)
	}
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "delete" || meta["source_endpoint"] != workoutLibraryWorkoutsEndpoint {
		t.Fatalf("meta = %#v, want delete metadata", meta)
	}
}

func TestDeleteWorkoutRejectsBadArguments(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutDeleterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newDeleteWorkoutTool(client, client, "test", "UTC", false)
	for _, raw := range []string{
		`{}`,
		`{"workout_id":" "}`,
		`{"workout_id":"w-1","confirm":true}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestDeleteWorkoutRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutDeleterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	registrar := &collectingRegistrar{}
	if err := NewRegistry(client, "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	tool := findTool(t, registrar.tools, deleteWorkoutName)
	if tool.Requirement != RequirementDelete || !tool.RequiresDelete() {
		t.Fatalf("requirement = %q delete=%v, want delete", tool.Requirement, tool.RequiresDelete())
	}
	if !strings.Contains(tool.Description, "ICUVISOR_DELETE_MODE=full") {
		t.Fatalf("description = %q, want full-mode gate language", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	if _, ok := props["workout_id"]; !ok {
		t.Fatalf("schema missing workout_id: %#v", props)
	}
	if _, ok := props["confirm"]; ok || strings.Contains(strings.ToLower(tool.Description), "confirm argument") && len(props) != 1 {
		t.Fatalf("schema includes confirm or unexpected properties: %#v", props)
	}
}

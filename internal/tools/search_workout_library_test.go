package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type fakeSharedLibraryClient struct {
	folders  []intervals.WorkoutFolder
	folder   intervals.WorkoutFolder
	err      error
	calls    int
	params   intervals.WorkoutLibrarySearchParams
	folderID string
	target   string
}

func (f *fakeSharedLibraryClient) SearchWorkoutLibrary(ctx context.Context, p intervals.WorkoutLibrarySearchParams) ([]intervals.WorkoutFolder, error) {
	f.calls++
	f.params = p
	return f.folders, f.err
}
func (f *fakeSharedLibraryClient) AddSharedWorkoutFolder(ctx context.Context, id string) (intervals.WorkoutFolder, error) {
	f.calls++
	f.folderID = id
	f.target, _ = intervals.TargetAthleteIDFromContext(ctx)
	return f.folder, f.err
}

func TestSearchWorkoutLibraryTerseFullAndBounds(t *testing.T) {
	t.Parallel()
	for _, full := range []bool{false, true} {
		t.Run(fmt.Sprint(full), func(t *testing.T) {
			client := &fakeSharedLibraryClient{folders: decodeToolWorkoutFolders(t, `{"id":10,"athlete_id":"i999","name":"Base","type":"PLAN","activity_types":["Run"],"duration_weeks":8,"hours_per_week_min":4,"hours_per_week_max":6,"workout_targets":["PACE"],"num_workouts":32,"description":"Plan detail","blurb":"Plan blurb","shareToken":"secret-token","owner":{"id":"i999","email":"private@example.test","name":"Private owner"},"children":[{"id":1,"description":"heavy step text","workout_doc":{"steps":[{"duration":600}]},"shareToken":"nested-secret"}]}`, `{"id":11,"name":"Beyond requested limit"}`)}
			tool := newSearchWorkoutLibraryTool(client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(fmt.Sprintf(`{"q":" base ","weeks_min":0,"weeks_max":12,"hours_per_week_min":0,"hours_per_week_max":8,"limit":1,"include_full":%t}`, full))})
			if err != nil {
				t.Fatal(err)
			}
			out := resultMap(t, result)
			rows := out["folders"].([]any)
			if len(rows) != 1 {
				t.Fatalf("rows = %v", rows)
			}
			row := rows[0].(map[string]any)
			if row["folder_id"] != "10" || row["duration_weeks"] != float64(8) || row["hours_per_week_min"] != float64(4) || row["num_workouts"] != float64(32) {
				t.Fatalf("row = %#v", row)
			}
			if _, ok := row["description"]; ok != full {
				t.Fatalf("description full=%t: %#v", full, row)
			}
			encoded, _ := json.Marshal(out)
			for _, secret := range []string{"secret-token", "nested-secret", "private@example.test", "Private owner", "heavy step text", "shareToken", "workout_doc"} {
				if strings.Contains(string(encoded), secret) {
					t.Fatalf("sensitive or heavy field %q leaked: %s", secret, encoded)
				}
			}
			meta := out["_meta"].(map[string]any)
			if meta["limit"] != float64(1) || meta["truncated"] != true || meta["limit_reached"] != true {
				t.Fatalf("meta = %#v", meta)
			}
			if _, ok := meta["next_page_token"]; ok {
				t.Fatal("invented cursor")
			}
			if client.params.Query != "base" || client.params.Limit != 1 {
				t.Fatalf("params = %#v", client.params)
			}
		})
	}
}
func TestSearchWorkoutLibraryValidatesBeforeIO(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{}`, `null`, `{"q":"base"}`, `{"q":"base","weeks_min":null,"weeks_max":12,"hours_per_week_min":0,"hours_per_week_max":8}`, `{"q":"base","weeks_min":0,"weeks_max":12,"hours_per_week_min":0,"hours_per_week_max":8,"limit":0}`, `{"q":"base","weeks_min":0,"weeks_max":12,"hours_per_week_min":0,"hours_per_week_max":8,"page":2}`} {
		t.Run(raw, func(t *testing.T) {
			client := &fakeSharedLibraryClient{}
			tool := newSearchWorkoutLibraryTool(client, "test", false)
			if _, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(raw)}); err == nil {
				t.Fatal("wanted validation error")
			}
			if client.calls != 0 {
				t.Fatal("invalid arguments reached client")
			}
		})
	}
}
func TestSearchWorkoutLibraryEmptyAndErrors(t *testing.T) {
	t.Parallel()
	for _, cause := range []error{nil, errors.New("upstream private details"), context.Canceled, context.DeadlineExceeded} {
		t.Run(fmt.Sprint(cause), func(t *testing.T) {
			client := &fakeSharedLibraryClient{err: cause}
			tool := newSearchWorkoutLibraryTool(client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"q":"base","weeks_min":0,"weeks_max":12,"hours_per_week_min":0,"hours_per_week_max":8}`)})
			if cause == nil {
				if err != nil {
					t.Fatal(err)
				}
				if len(resultMap(t, result)["folders"].([]any)) != 0 {
					t.Fatal("expected empty results")
				}
				if client.params.Limit != 3 {
					t.Fatalf("default limit = %d", client.params.Limit)
				}
				return
			}
			if !errors.Is(err, cause) {
				t.Fatalf("error = %v, wanted cause %v", err, cause)
			}
		})
	}
}
func TestSharedLibraryToolDeclarations(t *testing.T) {
	t.Parallel()
	for _, tool := range []Tool{newSearchWorkoutLibraryTool(nil, "test", false), newAddSharedWorkoutFolderTool(nil, "test", false)} {
		if tool.EffectiveToolset() != safety.ToolsetFull {
			t.Fatalf("toolset = %s", tool.EffectiveToolset())
		}
		want := RequirementRead
		if tool.Name == "add_shared_workout_folder" {
			want = RequirementWrite
		}
		if tool.Requirement != want {
			t.Fatalf("%s requirement = %s", tool.Name, tool.Requirement)
		}
		schema := tool.InputSchema.(map[string]any)
		if len(schema["examples"].([]map[string]any)) == 0 {
			t.Fatal("missing example")
		}
		for name, p := range schema["properties"].(map[string]any) {
			if p.(map[string]any)["description"] == "" {
				t.Fatalf("missing description for %s", name)
			}
		}
	}
}

func TestSearchWorkoutLibraryDefaultPopulatedPayloadBudget(t *testing.T) {
	t.Parallel()
	client := &fakeSharedLibraryClient{}
	for i := 0; i < 30; i++ {
		client.folders = append(client.folders, decodeToolWorkoutFolders(t, fmt.Sprintf(`{"id":%d,"name":"Aerobic base plan %d","type":"PLAN","activity_types":["Run","Ride"],"duration_weeks":8,"hours_per_week_min":4,"hours_per_week_max":7,"workout_targets":["POWER","PACE"],"num_workouts":32,"description":"Extended program notes omitted from terse results"}`, 100+i, i+1))[0])
	}
	tool := newSearchWorkoutLibraryTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"q":"base","weeks_min":0,"weeks_max":12,"hours_per_week_min":0,"hours_per_week_max":8}`)})
	if err != nil {
		t.Fatal(err)
	}
	responseBytes := len(result.Content[0].Text)
	estimatedTokens := (responseBytes + 3) / 4
	t.Logf("populated default response = %d bytes, about %d tokens at 4 bytes/token", responseBytes, estimatedTokens)
	if estimatedTokens > 500 {
		t.Errorf("populated default response estimate = %d tokens, want at most about 500", estimatedTokens)
	}
	if got := len(resultMap(t, result)["folders"].([]any)); got != 3 {
		t.Errorf("default returned %d folders, want 3", got)
	}
	if client.params.Limit != 3 {
		t.Errorf("upstream limit = %d, want 3", client.params.Limit)
	}
}

func TestSearchWorkoutLibraryExplicitLimitRange(t *testing.T) {
	t.Parallel()
	for _, limit := range []int{1, 100} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			client := &fakeSharedLibraryClient{}
			for i := 0; i < 100; i++ {
				client.folders = append(client.folders, intervals.WorkoutFolder{ID: fmt.Sprint(i + 1)})
			}
			tool := newSearchWorkoutLibraryTool(client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(fmt.Sprintf(`{"q":"base","weeks_min":0,"weeks_max":12,"hours_per_week_min":0,"hours_per_week_max":8,"limit":%d}`, limit))})
			if err != nil {
				t.Fatal(err)
			}
			if got := len(resultMap(t, result)["folders"].([]any)); got != limit || client.params.Limit != limit {
				t.Fatalf("returned %d, upstream limit %d; want explicit %d", got, client.params.Limit, limit)
			}
		})
	}
}

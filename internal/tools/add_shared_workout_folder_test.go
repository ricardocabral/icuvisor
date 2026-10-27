package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestAddSharedWorkoutFolderTargetAndRedaction(t *testing.T) {
	t.Parallel()
	client := &fakeSharedLibraryClient{folder: decodeToolWorkoutFolders(t, `{"id":10,"athlete_id":"i999","name":"Base","shareToken":"secret","owner":{"email":"private@example.test"}}`)[0]}
	tool := newAddSharedWorkoutFolderTool(client, "test", false)
	result, err := tool.Handler(intervals.WithTargetAthleteID(context.Background(), "i222"), Request{Arguments: json.RawMessage(`{"folder_id":"10"}`)})
	if err != nil {
		t.Fatal(err)
	}
	if client.target != "i222" || client.folderID != "10" {
		t.Fatalf("target %s folder %s", client.target, client.folderID)
	}
	out := resultMap(t, result)
	row := out["folder"].(map[string]any)
	if row["folder_id"] != "10" || out["_meta"].(map[string]any)["verified_in_library"] != true {
		t.Fatalf("output = %#v", out)
	}
	if _, ok := row["shareToken"]; ok {
		t.Fatal("share token leaked")
	}
	if _, ok := row["owner"]; ok {
		t.Fatal("owner leaked")
	}
}
func TestAddSharedWorkoutFolderRejectsInvalidAndUnverified(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{}`, `{"folder_id":"../10"}`, `{"folder_id":"0"}`, `{"folder_id":"10","athlete_id":"i333"}`, `{"folder_id":"10","include_full":true}`} {
		t.Run(raw, func(t *testing.T) {
			client := &fakeSharedLibraryClient{}
			tool := newAddSharedWorkoutFolderTool(client, "test", false)
			if _, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(raw)}); err == nil {
				t.Fatal("wanted error")
			}
			if client.calls != 0 {
				t.Fatal("invalid request reached client")
			}
		})
	}
	for _, id := range []string{"", "11"} {
		client := &fakeSharedLibraryClient{folder: intervals.WorkoutFolder{ID: id}}
		tool := newAddSharedWorkoutFolderTool(client, "test", false)
		if _, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"folder_id":"10"}`)}); err == nil {
			t.Fatal("wrong folder reported success")
		}
	}
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded, errors.New("403 details")} {
		client := &fakeSharedLibraryClient{err: cause}
		tool := newAddSharedWorkoutFolderTool(client, "test", false)
		if _, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"folder_id":"10"}`)}); !errors.Is(err, cause) {
			t.Fatalf("error = %v, want %v", err, cause)
		}
	}
}

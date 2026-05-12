package intervals

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkoutLibraryClientListsFoldersAndWorkouts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/athlete/i12345/folders":
			_, _ = w.Write([]byte(`[{"id":10,"type":"FOLDER","name":"Threshold","children":[{"id":2,"name":"FTP","type":"Ride","workout_doc":{"steps":[{"duration":600}]}}]}]`))
		case "/athlete/i12345/workouts":
			_, _ = w.Write([]byte(`[{"id":2,"name":"FTP","type":"Ride","folder_id":10,"workout_doc":{"steps":[{"duration":600}],"name":"raw"}}]`))
		default:
			t.Fatalf("path = %q, want folders or workouts", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	folders, err := client.ListWorkoutFolders(context.Background())
	if err != nil {
		t.Fatalf("ListWorkoutFolders() error = %v", err)
	}
	if len(folders) != 1 || folders[0].ID != "10" || len(folders[0].Children) != 1 || folders[0].Children[0].WorkoutDoc == nil {
		t.Fatalf("folders = %+v, want raw children and workout_doc", folders)
	}
	workouts, err := client.ListLibraryWorkouts(context.Background())
	if err != nil {
		t.Fatalf("ListLibraryWorkouts() error = %v", err)
	}
	if len(workouts) != 1 || workouts[0].ID != "2" || rawIDString(workouts[0].Raw["folder_id"]) != "10" {
		t.Fatalf("workouts = %+v, want folder_id preserved", workouts)
	}
	doc, ok := workouts[0].WorkoutDoc.(map[string]any)
	if !ok || doc["name"] != "raw" {
		t.Fatalf("workout_doc = %#v, want verbatim map", workouts[0].WorkoutDoc)
	}
}

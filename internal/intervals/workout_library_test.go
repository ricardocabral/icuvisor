package intervals

import (
	"context"
	"encoding/json"
	"io"
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

func TestCreateLibraryWorkoutSendsWritableFieldsOnly(t *testing.T) {
	t.Parallel()

	var request struct {
		method string
		path   string
		body   map[string]any
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		request = struct {
			method string
			path   string
			body   map[string]any
		}{method: r.Method, path: r.URL.Path, body: decoded}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"w-1","name":"Tempo","type":"Ride","folder_id":"f-20","description":"- 10m 65%","workout_doc":{"steps":[{"duration":600}]}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	description := "- 10m 65%"
	workout, err := client.CreateLibraryWorkout(context.Background(), WriteWorkoutParams{Name: " Tempo ", FolderID: " f-20 ", Sport: " Ride ", Description: &description, Tags: []string{"tempo", "coach"}})
	if err != nil {
		t.Fatalf("CreateLibraryWorkout() error = %v", err)
	}
	if workout.ID != "w-1" || workout.WorkoutDoc == nil {
		t.Fatalf("workout = %+v, want decoded workout with workout_doc", workout)
	}
	if request.method != http.MethodPost || request.path != "/athlete/i12345/workouts" {
		t.Fatalf("request = %#v, want POST athlete workouts", request)
	}
	body := request.body
	if body["name"] != "Tempo" || body["folder_id"] != "f-20" || body["type"] != "Ride" || body["description"] != description {
		t.Fatalf("body = %#v, want mapped workout fields", body)
	}
	if _, ok := body["workout_doc"]; ok {
		t.Fatalf("body includes workout_doc: %#v", body)
	}
	tags := body["tags"].([]any)
	if len(tags) != 2 || tags[0] != "tempo" || tags[1] != "coach" {
		t.Fatalf("tags = %#v, want preserved tags", tags)
	}
}

func TestCreateLibraryWorkoutRequiresWritableBasics(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.CreateLibraryWorkout(context.Background(), WriteWorkoutParams{Sport: "Ride"}); err == nil {
		t.Fatal("CreateLibraryWorkout() error = nil, want required name error")
	}
	if _, err := client.CreateLibraryWorkout(context.Background(), WriteWorkoutParams{Name: "Tempo"}); err == nil {
		t.Fatal("CreateLibraryWorkout() error = nil, want required sport error")
	}
}

func TestUpdateLibraryWorkoutSendsSparseWritableFieldsOnly(t *testing.T) {
	t.Parallel()

	var request struct {
		method string
		path   string
		body   map[string]any
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		request = struct {
			method string
			path   string
			body   map[string]any
		}{method: r.Method, path: r.URL.Path, body: decoded}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"w-2","name":"Renamed","type":"Ride"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	workout, err := client.UpdateLibraryWorkout(context.Background(), WriteWorkoutParams{WorkoutID: " w-2 ", Name: " Renamed ", NameSet: true})
	if err != nil {
		t.Fatalf("UpdateLibraryWorkout() error = %v", err)
	}
	if workout.ID != "w-2" {
		t.Fatalf("workout ID = %q, want decoded update response", workout.ID)
	}
	if request.method != http.MethodPut || request.path != "/athlete/i12345/workouts/w-2" {
		t.Fatalf("request = %#v, want PUT athlete workouts/{id}", request)
	}
	if len(request.body) != 1 || request.body["name"] != "Renamed" {
		t.Fatalf("body = %#v, want sparse name only", request.body)
	}
}

func TestUpdateLibraryWorkoutCanSendDescriptionTagsAndTopLevelFolder(t *testing.T) {
	t.Parallel()

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoded, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(decoded, &body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"w-3","name":"Tempo","type":"Ride"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	description := "- 15m 70%"
	_, err := client.UpdateLibraryWorkout(context.Background(), WriteWorkoutParams{WorkoutID: "w-3", FolderIDSet: true, Description: &description, DescriptionSet: true, Tags: []string{"base", "new"}, TagsSet: true})
	if err != nil {
		t.Fatalf("UpdateLibraryWorkout() error = %v", err)
	}
	if body["folder_id"] != "" || body["description"] != description {
		t.Fatalf("body = %#v, want explicit top-level folder and description", body)
	}
	tags := body["tags"].([]any)
	if len(tags) != 2 || tags[0] != "base" || tags[1] != "new" {
		t.Fatalf("tags = %#v, want replacement tag list", tags)
	}
	if _, ok := body["type"]; ok {
		t.Fatalf("body = %#v, want omitted sport untouched", body)
	}
}

func TestUpdateLibraryWorkoutRequiresIDAndField(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.UpdateLibraryWorkout(context.Background(), WriteWorkoutParams{Name: "Renamed", NameSet: true}); err == nil {
		t.Fatal("UpdateLibraryWorkout() error = nil, want required workout ID error")
	}
	if _, err := client.UpdateLibraryWorkout(context.Background(), WriteWorkoutParams{WorkoutID: "w-1"}); err == nil {
		t.Fatal("UpdateLibraryWorkout() error = nil, want required sparse field error")
	}
}

package intervals

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLibraryAndCustomItemsFixtureReads(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		routes   map[string]string
		validate func(t *testing.T, client *Client)
	}{
		{
			name: "empty library",
			routes: map[string]string{
				"/athlete/i12345/folders":  "testdata/workout_library/folders_empty.json",
				"/athlete/i12345/workouts": "testdata/workout_library/workouts_empty.json",
			},
			validate: func(t *testing.T, client *Client) {
				folders, err := client.ListWorkoutFolders(context.Background())
				if err != nil {
					t.Fatalf("ListWorkoutFolders() error = %v", err)
				}
				workouts, err := client.ListLibraryWorkouts(context.Background())
				if err != nil {
					t.Fatalf("ListLibraryWorkouts() error = %v", err)
				}
				if len(folders) != 0 || len(workouts) != 0 {
					t.Fatalf("library = %d folders/%d workouts, want empty", len(folders), len(workouts))
				}
			},
		},
		{
			name: "nested workout folders",
			routes: map[string]string{
				"/athlete/i12345/folders":  "testdata/workout_library/folders_nested.json",
				"/athlete/i12345/workouts": "testdata/workout_library/workouts_nested.json",
			},
			validate: func(t *testing.T, client *Client) {
				folders, err := client.ListWorkoutFolders(context.Background())
				if err != nil {
					t.Fatalf("ListWorkoutFolders() error = %v", err)
				}
				if len(folders) != 2 || folders[1].ID != "20" || len(folders[1].Children) != 2 {
					t.Fatalf("folders = %+v, want nested Threshold folder", folders)
				}
				if folders[1].Children[0].WorkoutDoc == nil {
					t.Fatalf("child workout_doc missing: %+v", folders[1].Children[0])
				}
				workouts, err := client.ListLibraryWorkouts(context.Background())
				if err != nil {
					t.Fatalf("ListLibraryWorkouts() error = %v", err)
				}
				if len(workouts) != 3 || rawIDString(workouts[1].Raw["folder_id"]) != "20" || workouts[2].WorkoutDoc == nil {
					t.Fatalf("workouts = %+v, want nested and top-level workout_doc preservation", workouts)
				}
			},
		},
		{
			name: "custom item variants",
			routes: map[string]string{
				"/athlete/i12345/custom-item":   "testdata/custom_items/custom_items_variants.json",
				"/athlete/i12345/custom-item/7": "testdata/custom_items/custom_item_7.json",
			},
			validate: func(t *testing.T, client *Client) {
				items, err := client.ListCustomItems(context.Background())
				if err != nil {
					t.Fatalf("ListCustomItems() error = %v", err)
				}
				seen := map[string]bool{}
				for _, item := range items {
					if item.Type != nil {
						seen[*item.Type] = true
					}
				}
				for _, want := range []string{"FITNESS_CHART", "INPUT_FIELD", "ZONES"} {
					if !seen[want] {
						t.Fatalf("custom item types = %#v, missing %s", seen, want)
					}
				}
				item, err := client.GetCustomItem(context.Background(), "7")
				if err != nil {
					t.Fatalf("GetCustomItem() error = %v", err)
				}
				content, ok := item.Content.(map[string]any)
				if !ok || content["series"] == nil || content["layout"] == nil {
					t.Fatalf("content = %#v, want chart series/layout preserved", item.Content)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fixture, ok := tc.routes[r.URL.Path]
				if !ok {
					t.Fatalf("path = %q, want one of %#v", r.URL.Path, tc.routes)
				}
				data, err := os.ReadFile(fixture)
				if err != nil {
					t.Fatalf("read fixture %s: %v", fixture, err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(data)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			tc.validate(t, client)
		})
	}
}

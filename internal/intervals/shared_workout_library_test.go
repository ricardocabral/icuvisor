package intervals

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestSearchWorkoutLibraryPublicQuery(t *testing.T) {
	t.Parallel()
	var query url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/folder-search" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		query = r.URL.Query()
		_, _ = w.Write([]byte(`[{"id":10,"athlete_id":"i999","name":"Base","type":"PLAN","activity_types":["Run"],"duration_weeks":8,"hours_per_week_min":4,"hours_per_week_max":6,"workout_targets":["PACE"],"num_workouts":32,"shareToken":"secret","owner":{"email":"private@example.test"}}]`))
	}))
	defer server.Close()
	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	got, err := client.SearchWorkoutLibrary(WithTargetAthleteID(context.Background(), "i222"), WorkoutLibrarySearchParams{Query: " base & build ", Type: "PLAN", ActivityTypes: []string{"Run", "Ride"}, WeeksMin: 0, WeeksMax: 12, HoursPerWeekMin: 0, HoursPerWeekMax: 8, WorkoutTargets: []string{"PACE", "HR"}})
	if err != nil {
		t.Fatal(err)
	}
	want := url.Values{"q": {"base & build"}, "type": {"PLAN"}, "activity_types": {"Run", "Ride"}, "weeks_min": {"0"}, "weeks_max": {"12"}, "hours_per_week_min": {"0"}, "hours_per_week_max": {"8"}, "workout_targets": {"PACE", "HR"}, "limit": {"3"}}
	if !reflect.DeepEqual(query, want) {
		t.Fatalf("query = %#v, want %#v", query, want)
	}
	if len(got) != 1 || got[0].ID != "10" || *got[0].AthleteID != "i999" || *got[0].DurationWeeks != 8 || len(got[0].WorkoutTargets) != 1 {
		t.Fatalf("folders = %+v", got)
	}
}

func TestSearchWorkoutLibraryValidation(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		change func(*WorkoutLibrarySearchParams)
	}{
		{"blank query", func(p *WorkoutLibrarySearchParams) { p.Query = " " }},
		{"negative weeks", func(p *WorkoutLibrarySearchParams) { p.WeeksMin = -1 }},
		{"reversed weeks", func(p *WorkoutLibrarySearchParams) { p.WeeksMin = 13 }},
		{"unbounded weeks", func(p *WorkoutLibrarySearchParams) { p.WeeksMax = 521 }},
		{"reversed hours", func(p *WorkoutLibrarySearchParams) { p.HoursPerWeekMin = 9 }},
		{"unbounded hours", func(p *WorkoutLibrarySearchParams) { p.HoursPerWeekMax = 169 }},
		{"negative limit", func(p *WorkoutLibrarySearchParams) { p.Limit = -1 }},
		{"large limit", func(p *WorkoutLibrarySearchParams) { p.Limit = 101 }},
		{"invalid type", func(p *WorkoutLibrarySearchParams) { p.Type = "WORKOUT" }},
		{"invalid sport", func(p *WorkoutLibrarySearchParams) { p.ActivityTypes = []string{"Cycling"} }},
		{"invalid target", func(p *WorkoutLibrarySearchParams) { p.WorkoutTargets = []string{"FTP"} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			params := WorkoutLibrarySearchParams{Query: "base", WeeksMax: 12, HoursPerWeekMax: 8}
			tc.change(&params)
			if _, err := client.SearchWorkoutLibrary(context.Background(), params); err == nil {
				t.Fatal("wanted validation error")
			}
			if calls.Load() != 0 {
				t.Fatal("invalid filters reached upstream")
			}
		})
	}
}

func TestAddSharedWorkoutFolderVerifiesSelectedAthleteLibrary(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, post, get string
		wantErr         bool
	}{
		{"verified shared owner", `[{"id":10,"athlete_id":"i999","name":"Shared base"}]`, `[{"id":10,"athlete_id":"i999","name":"Shared base"}]`, false},
		{"empty post", `[]`, `[{"id":10}]`, true},
		{"wrong post folder", `[{"id":11}]`, `[{"id":10}]`, true},
		{"empty readback", `[{"id":10}]`, `[]`, true},
		{"wrong readback folder", `[{"id":10}]`, `[{"id":11}]`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, r.Method+" "+r.URL.Path)
				if r.Method == http.MethodPost {
					body, _ := io.ReadAll(r.Body)
					if len(body) != 0 {
						t.Errorf("unexpected body %s", body)
					}
					_, _ = w.Write([]byte(tc.post))
				} else {
					_, _ = w.Write([]byte(tc.get))
				}
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			folder, err := client.AddSharedWorkoutFolder(WithTargetAthleteID(context.Background(), "i222"), "10")
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, tc.wantErr)
			}
			if !tc.wantErr && folder.ID != "10" {
				t.Fatalf("folder = %+v", folder)
			}
			if len(calls) == 0 || calls[0] != "POST /athlete/i222/folders/10/share-with-me" {
				t.Fatalf("calls = %v", calls)
			}
			if len(calls) > 1 && calls[1] != "GET /athlete/i222/folders" {
				t.Fatalf("readback target = %v", calls)
			}
		})
	}
}

func TestSharedWorkoutLibraryErrorsAndCancellation(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 404, 429, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(status) }))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
			if _, err := client.SearchWorkoutLibrary(context.Background(), WorkoutLibrarySearchParams{Query: "base", WeeksMax: 12, HoursPerWeekMax: 8}); err == nil {
				t.Fatal("wanted search error")
			}
			calls.Store(0)
			if _, err := client.AddSharedWorkoutFolder(context.Background(), "10"); err == nil {
				t.Fatal("wanted share error")
			}
			if calls.Load() != 1 {
				t.Fatalf("POST retried %d times", calls.Load())
			}
		})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("canceled/invalid call reached upstream") }))
	defer server.Close()
	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	for _, id := range []string{"", "../10", "0", "-1", "2147483648"} {
		if _, err := client.AddSharedWorkoutFolder(context.Background(), id); err == nil {
			t.Errorf("folder %q accepted", id)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.SearchWorkoutLibrary(ctx, WorkoutLibrarySearchParams{Query: "base", WeeksMax: 12, HoursPerWeekMax: 8}); !errors.Is(err, context.Canceled) {
		t.Fatalf("search cancel = %v", err)
	}
	if _, err := client.AddSharedWorkoutFolder(ctx, "10"); !errors.Is(err, context.Canceled) {
		t.Fatalf("share cancel = %v", err)
	}
}

func TestAddSharedWorkoutFolderDoesNotLeakTargetAcrossRequests(t *testing.T) {
	t.Parallel()
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		_, _ = w.Write([]byte(`[{"id":10,"athlete_id":"i999"}]`))
	}))
	defer server.Close()
	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	for _, ctx := range []context.Context{WithTargetAthleteID(context.Background(), "i222"), WithTargetAthleteID(context.Background(), "333"), context.Background()} {
		if _, err := client.AddSharedWorkoutFolder(ctx, "10"); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"POST /athlete/i222/folders/10/share-with-me", "GET /athlete/i222/folders", "POST /athlete/333/folders/10/share-with-me", "GET /athlete/333/folders", "POST /athlete/i12345/folders/10/share-with-me", "GET /athlete/i12345/folders"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
}

func TestAddSharedWorkoutFolderReadbackFailureIsNotSuccess(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 404, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var posts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					posts.Add(1)
					_, _ = w.Write([]byte(`[{"id":10}]`))
					return
				}
				w.WriteHeader(status)
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
			if folder, err := client.AddSharedWorkoutFolder(context.Background(), "10"); err == nil || folder.ID != "" {
				t.Fatalf("folder = %+v, error = %v; wanted unverified write", folder, err)
			}
			if posts.Load() != 1 {
				t.Fatalf("posts = %d, want no write retry", posts.Load())
			}
		})
	}
}

package intervals

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestEventsClientEndpointsUseHTTPFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		path     string
		fixture  string
		call     func(context.Context, *Client) (string, error)
		validate func(*testing.T, *http.Request)
	}{
		{
			name:    "list",
			path:    "/athlete/i12345/events",
			fixture: "testdata/events/inconsistent/synthetic_list.json",
			call: func(ctx context.Context, client *Client) (string, error) {
				rows, err := client.ListEvents(ctx, ListEventsParams{Oldest: "2026-03-01", Newest: "2026-03-31"})
				if err != nil {
					return "", err
				}
				return rows[0].ID, nil
			},
			validate: func(t *testing.T, r *http.Request) {
				t.Helper()
				if got := r.URL.Query().Get("oldest"); got != "2026-03-01" {
					t.Fatalf("oldest = %q, want fixture range start", got)
				}
			},
		},
		{
			name:    "detail",
			path:    "/athlete/i12345/events/synthetic-detail-1",
			fixture: "testdata/events/detail.json",
			call: func(ctx context.Context, client *Client) (string, error) {
				event, err := client.GetEvent(ctx, "synthetic-detail-1")
				if err != nil {
					return "", err
				}
				return event.ID, nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, err := os.ReadFile(tc.fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Path; got != tc.path {
					t.Fatalf("path = %q, want %q", got, tc.path)
				}
				if tc.validate != nil {
					tc.validate(t, r)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(body)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			id, err := tc.call(context.Background(), client)
			if err != nil {
				t.Fatalf("call() error = %v", err)
			}
			if id == "" {
				t.Fatal("id = empty, want decoded fixture ID")
			}
		})
	}
}

func TestListEventsSendsDocumentedQueryAndPreservesRaw(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345/events"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		query := r.URL.Query()
		checks := map[string]string{"oldest": "2026-01-01", "newest": "2026-01-31", "category": "WORKOUT", "calendar_id": "cal-1", "limit": "50", "resolve": "true"}
		for key, want := range checks {
			if got := query.Get(key); got != want {
				t.Fatalf("query %s = %q, want %q", key, got, want)
			}
		}
		if got := query.Get("fields"); got != "" {
			t.Fatalf("query fields = %q, want absent", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":123,"name":null,"category":"WORKOUT","type":"Ride","start_date_local":"2026-01-03","workout_doc":{"steps":[{"duration":600}]}}
		]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	resolve := true
	events, err := client.ListEvents(context.Background(), ListEventsParams{Oldest: "2026-01-01", Newest: "2026-01-31", Category: "WORKOUT", CalendarID: "cal-1", Limit: 50, Resolve: &resolve})
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	if events[0].ID != "123" {
		t.Fatalf("ID = %q, want stringified numeric upstream ID", events[0].ID)
	}
	if rawName, ok := events[0].Raw["name"]; !ok || rawName != nil {
		rawJSON, _ := json.Marshal(events[0].Raw)
		t.Fatalf("raw name = %#v (present %v), raw = %s; want present nil", rawName, ok, rawJSON)
	}
	if events[0].WorkoutDoc == nil {
		t.Fatal("WorkoutDoc = nil, want preserved nested workout_doc")
	}
}

func TestListEventsRequiresDateRange(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.ListEvents(context.Background(), ListEventsParams{Newest: "2026-01-31"}); err == nil {
		t.Fatal("ListEvents() error = nil, want required oldest error")
	}
	if _, err := client.ListEvents(context.Background(), ListEventsParams{Oldest: "2026-01-01"}); err == nil {
		t.Fatal("ListEvents() error = nil, want required newest error")
	}
}

func TestGetEventSendsAthleteScopedDetailPathAndPreservesRaw(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345/events/evt-123"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got := r.URL.RawQuery; got != "" {
			t.Fatalf("query = %q, want empty", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"evt-123","name":"Long run","category":"WORKOUT","start_date_local":"2026-01-03","description":null}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	event, err := client.GetEvent(context.Background(), " evt-123 ")
	if err != nil {
		t.Fatalf("GetEvent() error = %v", err)
	}
	if event.ID != "evt-123" || event.Category == nil || *event.Category != "WORKOUT" {
		t.Fatalf("event = %+v, want decoded detail row", event)
	}
	if rawDescription, ok := event.Raw["description"]; !ok || rawDescription != nil {
		t.Fatalf("raw description = %#v (present %v), want present nil", rawDescription, ok)
	}
}

func TestGetEventRequiresID(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.GetEvent(context.Background(), " "); err == nil {
		t.Fatal("GetEvent() error = nil, want required ID error")
	}
}

func TestAddOrUpdateEventSendsCreateAndUpdateBodies(t *testing.T) {
	t.Parallel()

	requests := make([]struct {
		method string
		path   string
		body   map[string]any
	}, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		requests = append(requests, struct {
			method string
			path   string
			body   map[string]any
		}{method: r.Method, path: r.URL.Path, body: decoded})
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			_, _ = w.Write([]byte(`{"id":"evt-9","category":"WORKOUT","start_date_local":"2026-06-02"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"evt-8","category":"WORKOUT","start_date_local":"2026-06-01"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	description := "  preserve\ntext  "
	targetLoad := 75.0
	distance := 30000.0
	moving := 3600
	elapsed := 3900
	created, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{Date: "2026-06-01", Category: "WORKOUT", Name: "Tempo", Description: &description, Tags: []string{"tempo", "coach"}, TargetLoad: &targetLoad, DistanceMeters: &distance, MovingTimeSeconds: &moving, ElapsedTimeSeconds: &elapsed})
	if err != nil {
		t.Fatalf("AddOrUpdateEvent(create) error = %v", err)
	}
	updated, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: " evt-9 ", Date: "2026-06-02", Category: "WORKOUT"})
	if err != nil {
		t.Fatalf("AddOrUpdateEvent(update) error = %v", err)
	}
	if created.ID != "evt-8" || updated.ID != "evt-9" {
		t.Fatalf("ids = %q/%q, want decoded create/update IDs", created.ID, updated.ID)
	}
	if len(requests) != 2 {
		t.Fatalf("request count = %d, want 2", len(requests))
	}
	if requests[0].method != http.MethodPost || requests[0].path != "/athlete/i12345/events" {
		t.Fatalf("create request = %#v, want POST athlete events", requests[0])
	}
	body := requests[0].body
	if body["start_date_local"] != "2026-06-01" || body["category"] != "WORKOUT" || body["name"] != "Tempo" || body["description"] != description {
		t.Fatalf("create body = %#v, want mapped event fields", body)
	}
	if _, ok := body["workout_doc"]; ok {
		t.Fatalf("create body includes workout_doc: %#v", body)
	}
	if _, ok := body["event_id"]; ok {
		t.Fatalf("create body includes event_id: %#v", body)
	}
	tags := body["tags"].([]any)
	if len(tags) != 2 || tags[0] != "tempo" || tags[1] != "coach" {
		t.Fatalf("tags = %#v, want preserved tags", tags)
	}
	if body["load_target"] != float64(75) || body["distance_target"] != float64(30000) || body["time_target"] != float64(3600) || body["elapsed_time_target"] != float64(3900) {
		t.Fatalf("planned metrics body = %#v", body)
	}
	for _, completedKey := range []string{"icu_training_load", "distance", "moving_time", "elapsed_time"} {
		if _, ok := body[completedKey]; ok {
			t.Fatalf("planned write body includes completed metric %s: %#v", completedKey, body)
		}
	}
	if requests[1].method != http.MethodPut || requests[1].path != "/athlete/i12345/events/evt-9" {
		t.Fatalf("update request = %#v, want PUT athlete events/{id}", requests[1])
	}
}

func TestAddOrUpdateEventRequiresWritableBasics(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{Category: "WORKOUT"}); err == nil {
		t.Fatal("AddOrUpdateEvent() error = nil, want required date error")
	}
	if _, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{Date: "2026-01-01"}); err == nil {
		t.Fatal("AddOrUpdateEvent() error = nil, want required category error")
	}
}

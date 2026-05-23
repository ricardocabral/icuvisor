package intervals

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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

func TestAddOrUpdateEventSendsNoteCreateBody(t *testing.T) {
	t.Parallel()

	wantBodyBytes, err := os.ReadFile("testdata/events/note_create_request.json")
	if err != nil {
		t.Fatalf("read NOTE request fixture: %v", err)
	}
	responseBody, err := os.ReadFile("testdata/events/note_create_response.json")
	if err != nil {
		t.Fatalf("read NOTE response fixture: %v", err)
	}
	var wantBody any
	if err := json.Unmarshal(wantBodyBytes, &wantBody); err != nil {
		t.Fatalf("decode NOTE request fixture: %v", err)
	}
	var gotBody any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/athlete/i12345/events/bulk"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	description := "tp-075 captured note fixture"
	event, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{Date: "2026-05-25", Category: "NOTE", Name: "tp-075 fixture note", Description: &description})
	if err != nil {
		t.Fatalf("AddOrUpdateEvent(NOTE create) error = %v", err)
	}
	if event.ID != "EVENT_ID_PLACEHOLDER" || event.Category == nil || *event.Category != "NOTE" {
		t.Fatalf("event = %+v, want decoded NOTE fixture", event)
	}
	if !reflect.DeepEqual(gotBody, wantBody) {
		gotJSON, _ := json.MarshalIndent(gotBody, "", "  ")
		wantJSON, _ := json.MarshalIndent(wantBody, "", "  ")
		t.Fatalf("NOTE create body = %s\nwant %s", gotJSON, wantJSON)
	}
}

func TestAddOrUpdateEventClosesRetryResponseBodyBeforeNextAttempt(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	var closed atomic.Int32
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempt := attempts.Add(1)
		if got, want := req.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := req.URL.Path, "/athlete/i12345/events/evt-9"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if attempt > 1 {
			wantClosed := attempt - 1
			if got := closed.Load(); got != wantClosed {
				t.Fatalf("closed bodies before attempt %d = %d, want %d", attempt, got, wantClosed)
			}
		}

		status := http.StatusServiceUnavailable
		body := `temporary`
		if attempt == 4 {
			status = http.StatusOK
			body = `{"id":"evt-9","category":"WORKOUT","type":"Ride","start_date_local":"2026-06-02T00:00:00"}`
		}
		return &http.Response{
			StatusCode: status,
			Header:     make(http.Header),
			Body:       &countingReadCloser{Reader: strings.NewReader(body), closed: &closed},
		}, nil
	})}
	client := newTestClient(t, "https://example.invalid", httpClient, RetryConfig{MaxAttempts: 4, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond})

	event, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: "evt-9", Date: "2026-06-02", Category: "WORKOUT", Type: "Ride"})
	if err != nil {
		t.Fatalf("AddOrUpdateEvent() error = %v", err)
	}
	if event.ID != "evt-9" {
		t.Fatalf("event ID = %q, want evt-9", event.ID)
	}
	if got := attempts.Load(); got != 4 {
		t.Fatalf("attempts = %d, want 4", got)
	}
	if got := closed.Load(); got != 4 {
		t.Fatalf("closed bodies = %d, want 4", got)
	}
}

func TestAddOrUpdateEventClosesResponseBodyOnDecodeError(t *testing.T) {
	t.Parallel()

	var closed atomic.Int32
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got, want := req.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       &countingReadCloser{Reader: strings.NewReader(`not-json`), closed: &closed},
		}, nil
	})}
	client := newTestClient(t, "https://example.invalid", httpClient, RetryConfig{MaxAttempts: 1})

	_, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: "evt-9", Date: "2026-06-02", Category: "WORKOUT", Type: "Ride"})
	if err == nil {
		t.Fatal("AddOrUpdateEvent() error = nil, want decode error")
	}
	if got := closed.Load(); got != 1 {
		t.Fatalf("closed bodies = %d, want 1", got)
	}
}

func TestAddOrUpdateEventSendsCreateAndUpdateBodies(t *testing.T) {
	t.Parallel()

	requests := make([]struct {
		method string
		path   string
		body   any
	}, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		requests = append(requests, struct {
			method string
			path   string
			body   any
		}{method: r.Method, path: r.URL.Path, body: decoded})
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			_, _ = w.Write([]byte(`{"id":"evt-9","category":"WORKOUT","start_date_local":"2026-06-02"}`))
			return
		}
		_, _ = w.Write([]byte(`[{"id":"evt-8","category":"WORKOUT","type":"Ride","start_date_local":"2026-06-01T00:00:00"}]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	description := "  preserve\ntext  "
	targetLoad := 75.0
	distance := 30000.0
	moving := 3600
	elapsed := 3900
	indoor := true
	created, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{Date: "2026-06-01", Category: "WORKOUT", Type: "VirtualRide", Name: "Tempo", Description: &description, Tags: []string{"tempo", "coach"}, Indoor: &indoor, TargetLoad: &targetLoad, DistanceMeters: &distance, MovingTimeSeconds: &moving, ElapsedTimeSeconds: &elapsed})
	if err != nil {
		t.Fatalf("AddOrUpdateEvent(create) error = %v", err)
	}
	indoorFalse := false
	updated, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: " evt-9 ", Date: "2026-06-02", Category: "WORKOUT", Type: "Ride", Indoor: &indoorFalse})
	if err != nil {
		t.Fatalf("AddOrUpdateEvent(update) error = %v", err)
	}
	if created.ID != "evt-8" || updated.ID != "evt-9" {
		t.Fatalf("ids = %q/%q, want decoded create/update IDs", created.ID, updated.ID)
	}
	if len(requests) != 2 {
		t.Fatalf("request count = %d, want 2", len(requests))
	}
	if requests[0].method != http.MethodPost || requests[0].path != "/athlete/i12345/events/bulk" {
		t.Fatalf("create request = %#v, want POST athlete events/bulk", requests[0])
	}
	createBatch := requests[0].body.([]any)
	if len(createBatch) != 1 {
		t.Fatalf("create body = %#v, want single-event bulk payload", requests[0].body)
	}
	body := createBatch[0].(map[string]any)
	if body["start_date_local"] != "2026-06-01T00:00:00" || body["category"] != "WORKOUT" || body["type"] != "VirtualRide" || body["name"] != "Tempo" || body["description"] != description || body["indoor"] != true {
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
	updateBody := requests[1].body.(map[string]any)
	if updateBody["indoor"] != false {
		t.Fatalf("update body = %#v, want explicit indoor=false", updateBody)
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

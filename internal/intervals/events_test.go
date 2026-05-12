package intervals

import (
	"context"
	"encoding/json"
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

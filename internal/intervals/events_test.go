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
		for _, key := range []string{"calendarId", "start", "end", "fields"} {
			if got := query.Get(key); got != "" {
				t.Fatalf("query %s = %q, want absent", key, got)
			}
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

func TestWriteEventBodyNormalizesDateOnlyAndPreservesTimestamp(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		date string
		want string
	}{
		{name: "date only", date: "2027-01-01", want: "2027-01-01T00:00:00"},
		{name: "full timestamp", date: "2027-01-01T09:30:00", want: "2027-01-01T09:30:00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, err := writeEventBody(WriteEventParams{Date: tc.date, Category: "RACE_A"})
			if err != nil {
				t.Fatalf("writeEventBody() error = %v", err)
			}
			if body.StartDateLocal != tc.want {
				t.Fatalf("start_date_local = %q, want %q", body.StartDateLocal, tc.want)
			}
		})
	}
}

func TestAddOrUpdateEventPreservesOmittedStartTime(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		start string
		date  string
		want  string
		end   string
	}{
		{name: "same day", start: "2026-07-10T07:30:45", date: "2026-07-10", want: "2026-07-10T07:30:45"},
		{name: "reschedule", start: "2026-07-10T07:30:45", date: "2026-07-11", want: "2026-07-11T07:30:45"},
		{name: "fractional seconds", start: "2026-07-10T07:30:45.123", date: "2026-07-11", want: "2026-07-11T07:30:45.123"},
		{name: "date only", start: "2026-07-10", date: "2026-07-11", want: "2026-07-11T00:00:00"},
		{name: "preserved end after new start", start: "2026-07-10T07:30:00", date: "2026-07-11", want: "2026-07-11T07:30:00", end: "2026-07-11T08:30:00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var reads, writes atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/athlete/i12345/events/123" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					reads.Add(1)
					event := map[string]any{"id": 123, "start_date_local": tc.start}
					if tc.end != "" {
						event["end_date_local"] = tc.end
					}
					_ = json.NewEncoder(w).Encode(event)
					return
				}
				writes.Add(1)
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if r.Method != http.MethodPut || body["start_date_local"] != tc.want {
					t.Errorf("request = %s %#v, want PUT with start %s", r.Method, body, tc.want)
				}
				if _, ok := body["end_date_local"]; ok {
					t.Error("omitted end_date_local must stay absent from the write")
				}
				_, _ = w.Write([]byte(`{"id":123,"start_date_local":"` + tc.want + `"}`))
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			_, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: "123", Date: tc.date, Category: "NOTE"})
			if err != nil {
				t.Fatal(err)
			}
			if reads.Load() != 1 || writes.Load() != 1 {
				t.Fatalf("reads/writes = %d/%d, want 1/1", reads.Load(), writes.Load())
			}
		})
	}
}

func TestAddOrUpdateEventWritesExplicitStartTime(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		eventID string
		time    string
		method  string
		path    string
		end     *string
	}{
		{name: "create", time: "07:30", method: http.MethodPost, path: "/athlete/i12345/events/bulk"},
		{name: "update", eventID: "123", time: "17:45", method: http.MethodPut, path: "/athlete/i12345/events/123"},
		{name: "clear time", eventID: "123", time: "00:00", method: http.MethodPut, path: "/athlete/i12345/events/123"},
		{name: "explicit timing without read", eventID: "123", time: "17:45", method: http.MethodPut, path: "/athlete/i12345/events/123", end: stringPtr("2026-07-10T18:45:00")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			want := "2026-07-10T" + tc.time + ":00"
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"id":123,"start_date_local":"2026-07-10T07:30:00","end_date_local":null}`))
					return
				}
				if r.Method != tc.method || r.URL.Path != tc.path {
					t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, tc.method, tc.path)
				}
				var decoded any
				if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if batch, ok := decoded.([]any); ok {
					decoded = batch[0]
				}
				body := decoded.(map[string]any)
				if body["start_date_local"] != want {
					t.Errorf("start_date_local = %v, want %s", body["start_date_local"], want)
				}
				if _, ok := body["start_time"]; ok {
					t.Error("start_time must be encoded into start_date_local")
				}
				w.Header().Set("Content-Type", "application/json")
				event := `{"id":123,"start_date_local":"` + want + `"}`
				if tc.method == http.MethodPost {
					event = "[" + event + "]"
				}
				_, _ = w.Write([]byte(event))
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			event, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: tc.eventID, Date: "2026-07-10", StartTime: &tc.time, EndDateLocal: tc.end, Category: "NOTE", Name: "Reminder"})
			if err != nil {
				t.Fatal(err)
			}
			wantRequests := int32(1)
			if tc.eventID != "" && tc.end == nil {
				wantRequests = 2
			}
			if requests.Load() != wantRequests || event.StartDateLocal == nil || *event.StartDateLocal != want {
				t.Fatalf("requests = %d, event = %#v, want %d requests with start %s", requests.Load(), event, wantRequests, want)
			}
		})
	}
}

func TestAddOrUpdateEventCannotPreserveStartTimeWithoutEvidence(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		status    int
		body      string
		startTime *string
		end       *string
	}{
		{name: "read failure", status: http.StatusNotFound, body: `{}`},
		{name: "missing timestamp", status: http.StatusOK, body: `{"id":123}`},
		{name: "invalid timestamp", status: http.StatusOK, body: `{"id":123,"start_date_local":"invalid"}`},
		{name: "end precedes preserved time", status: http.StatusOK, body: `{"id":123,"start_date_local":"2026-07-10T07:30:00"}`, end: stringPtr("2026-07-11T07:00:00")},
		{name: "preserved end precedes rescheduled start", status: http.StatusOK, body: `{"id":123,"start_date_local":"2026-07-10T07:30:00","end_date_local":"2026-07-10T08:30:00"}`},
		{name: "preserved end precedes explicit start", status: http.StatusOK, body: `{"id":123,"start_date_local":"2026-07-10T07:30:00","end_date_local":"2026-07-11T08:30:00"}`, startTime: stringPtr("09:00")},
		{name: "read failure with explicit start", status: http.StatusNotFound, body: `{}`, startTime: stringPtr("09:00")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var writes atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					writes.Add(1)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
			_, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: "123", Date: "2026-07-11", StartTime: tc.startTime, EndDateLocal: tc.end, Category: "NOTE"})
			if err == nil || writes.Load() != 0 {
				t.Fatalf("error = %v, writes = %d, want failure before writing", err, writes.Load())
			}
		})
	}
}

func TestAddOrUpdateEventSendsRaceCreateAtLocalMidnight(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/athlete/i12345/events/bulk"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		var batch []map[string]any
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if len(batch) != 1 || batch[0]["category"] != "RACE_A" || batch[0]["start_date_local"] != "2027-01-01T00:00:00" {
			t.Fatalf("request body = %#v, want RACE_A at local midnight", batch)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"race-1","category":"RACE_A","start_date_local":"2027-01-01T00:00:00"}]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	if _, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{Date: "2027-01-01", Category: "RACE_A", Name: "Goal race"}); err != nil {
		t.Fatalf("AddOrUpdateEvent(RACE_A create) error = %v", err)
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

func TestAddOrUpdateEventSendsPresentationFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var batch []map[string]any
		if err := json.Unmarshal(body, &batch); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if len(batch) != 1 || batch[0]["color"] != "#ff8800" || batch[0]["not_on_fitness_chart"] != true {
			t.Fatalf("request body = %#v, want presentation fields", batch)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"evt-presentation","category":"NOTE","color":"#ff8800","not_on_fitness_chart":true}]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	color := " #ff8800 "
	notOnFitnessChart := true
	if _, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{Date: "2026-06-02", Category: "NOTE", Name: "Travel", Color: &color, NotOnFitnessChart: &notOnFitnessChart}); err != nil {
		t.Fatalf("AddOrUpdateEvent() error = %v", err)
	}
}

func TestAddOrUpdateEventSendsExplicitFalseForNotOnFitnessChart(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var batch []map[string]any
		if err := json.Unmarshal(body, &batch); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		value, ok := batch[0]["not_on_fitness_chart"]
		if !ok || value != false {
			t.Fatalf("request body = %#v, want explicit not_on_fitness_chart=false", batch)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"evt-visible","category":"NOTE","not_on_fitness_chart":false}]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	notOnFitnessChart := false
	if _, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{Date: "2026-06-03", Category: "NOTE", Name: "Visible", NotOnFitnessChart: &notOnFitnessChart}); err != nil {
		t.Fatalf("AddOrUpdateEvent() error = %v", err)
	}
}

func TestAddOrUpdateEventClosesRetryResponseBodyBeforeNextAttempt(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	var closed atomic.Int32
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"evt-9","start_date_local":"2026-06-02T00:00:00"}`))}, nil
		}
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
		if req.Method == http.MethodGet {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"evt-9","start_date_local":"2026-06-02T00:00:00"}`))}, nil
		}
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
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"evt-9","start_date_local":"2026-06-02T00:00:00"}`))
			return
		}
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
	created, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{ExternalID: " ext-create ", Date: "2026-06-01", Category: "WORKOUT", Type: "VirtualRide", Name: "Tempo", Description: &description, Tags: []string{"tempo", "coach"}, Indoor: &indoor, TargetLoad: &targetLoad, DistanceMeters: &distance, MovingTimeSeconds: &moving, ElapsedTimeSeconds: &elapsed})
	if err != nil {
		t.Fatalf("AddOrUpdateEvent(create) error = %v", err)
	}
	indoorFalse := false
	updated, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: " evt-9 ", ExternalID: "ext-update", Date: "2026-06-02", Category: "WORKOUT", Type: "Ride", Indoor: &indoorFalse})
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
	if body["start_date_local"] != "2026-06-01T00:00:00" || body["external_id"] != "ext-create" || body["category"] != "WORKOUT" || body["type"] != "VirtualRide" || body["name"] != "Tempo" || body["description"] != description || body["indoor"] != true {
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
	if updateBody["external_id"] != "ext-update" || updateBody["indoor"] != false {
		t.Fatalf("update body = %#v, want external_id plus explicit indoor=false", updateBody)
	}
}

func TestAddOrUpdateEventOmitsBlankExternalID(t *testing.T) {
	t.Parallel()

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"evt-blank","start_date_local":"2026-06-02T00:00:00"}`))
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"evt-blank","category":"WORKOUT","start_date_local":"2026-06-02"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	_, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: "evt-blank", ExternalID: "   ", Date: "2026-06-02", Category: "WORKOUT", Type: "Ride"})
	if err != nil {
		t.Fatalf("AddOrUpdateEvent() error = %v", err)
	}
	if _, ok := body["external_id"]; ok {
		t.Fatalf("update body includes blank external_id: %#v", body)
	}
}

func TestAddOrUpdateEventCanSendEmptyTags(t *testing.T) {
	t.Parallel()

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"evt-4","start_date_local":"2026-06-02T00:00:00"}`))
			return
		}
		if r.Method != http.MethodPut || r.URL.Path != "/athlete/i12345/events/evt-4" {
			t.Fatalf("request = %s %s, want PUT event", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"evt-4","category":"WORKOUT","start_date_local":"2026-06-02","tags":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	_, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: "evt-4", Date: "2026-06-02", Category: "WORKOUT", Type: "Ride", TagsSet: true})
	if err != nil {
		t.Fatalf("AddOrUpdateEvent() error = %v", err)
	}
	tags, ok := body["tags"].([]any)
	if !ok || len(tags) != 0 {
		t.Fatalf("tags = %#v, want explicit empty tags array", body["tags"])
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

func TestAddOrUpdateEventCalendarFieldsRoundTripAndSparseUpdate(t *testing.T) {
	t.Parallel()

	var requests []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":123,"start_date_local":"2026-07-10T00:00:00","end_date_local":"2026-07-12T16:00:00"}`))
			return
		}
		if r.URL.Path != "/athlete/i12345/events/bulk" && r.URL.Path != "/athlete/i12345/events/123" {
			t.Errorf("path = %q, want single-event create or numeric-ID update", r.URL.Path)
		}
		var body any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if batch, ok := body.([]any); ok {
			body = batch[0]
		}
		requests = append(requests, body.(map[string]any))
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`[{"id":123,"category":"NOTE","start_date_local":"2026-07-10T00:00:00","end_date_local":"2026-07-12T16:00:00","hide_from_athlete":true,"athlete_cannot_edit":true}]`))
			return
		}
		_, _ = w.Write([]byte(`{"id":123,"category":"NOTE","start_date_local":"2026-07-10T00:00:00","hide_from_athlete":false,"athlete_cannot_edit":false}`))
	}))
	defer server.Close()
	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	end := "2026-07-12T16:00:00"
	trueValue, falseValue := true, false
	created, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{Date: "2026-07-10", EndDateLocal: &end, Category: "NOTE", Name: "Stage race", HideFromAthlete: &trueValue, AthleteCannotEdit: &trueValue})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "123" || created.EndDateLocal == nil || *created.EndDateLocal != end || created.HideFromAthlete == nil || !*created.HideFromAthlete || created.AthleteCannotEdit == nil || !*created.AthleteCannotEdit {
		t.Fatalf("created event = %+v, want numeric ID and calendar fields", created)
	}
	updated, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: "123", Date: "2026-07-10", Category: "NOTE", HideFromAthlete: &falseValue, AthleteCannotEdit: &falseValue})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: "123", Date: "2026-07-10", Category: "NOTE", HideFromAthlete: &falseValue, AthleteCannotEdit: &falseValue}); err != nil {
		t.Fatalf("identical retry: %v", err)
	}
	if updated.HideFromAthlete == nil || *updated.HideFromAthlete || updated.AthleteCannotEdit == nil || *updated.AthleteCannotEdit {
		t.Fatalf("updated event = %+v, want explicit false flags", updated)
	}
	if len(requests) != 3 || requests[0]["end_date_local"] != end || requests[0]["hide_from_athlete"] != true || requests[0]["athlete_cannot_edit"] != true {
		t.Fatalf("create requests = %#v", requests)
	}
	for _, request := range requests[1:] {
		if _, ok := request["end_date_local"]; ok {
			t.Fatalf("update sent omitted end date: %#v", request)
		}
		if request["hide_from_athlete"] != false || request["athlete_cannot_edit"] != false {
			t.Fatalf("update did not send explicit false: %#v", request)
		}
	}
}

func TestEventWriteRequiresUnambiguousIdentityConfirmation(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/events/write_confirmations.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string          `json:"name"`
		EventID  string          `json:"event_id"`
		Response json.RawMessage `json:"response"`
		WantID   string          `json:"want_id"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			writes := 0
			client := newTestClient(t, "https://example.invalid", &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				wantMethod, wantPath := http.MethodPost, "/athlete/i12345/events/bulk"
				if tc.EventID != "" {
					wantMethod, wantPath = http.MethodPut, "/athlete/i12345/events/"+tc.EventID
				}
				if req.Method != wantMethod || req.URL.Path != wantPath {
					t.Fatalf("request = %s %s, want %s %s", req.Method, req.URL.Path, wantMethod, wantPath)
				}
				writes++
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(tc.Response)))}, nil
			})}, RetryConfig{})
			startTime, end := "08:00", "2026-08-10T09:00:00"
			event, err := client.AddOrUpdateEvent(context.Background(), WriteEventParams{EventID: tc.EventID, Date: "2026-08-10", Category: "NOTE", StartTime: &startTime, EndDateLocal: &end})
			if tc.WantID == "" {
				if err == nil || event.ID != "" {
					t.Fatalf("AddOrUpdateEvent() = %#v, %v; want unconfirmed write error", event, err)
				}
			} else if err != nil || event.ID != tc.WantID {
				t.Fatalf("AddOrUpdateEvent() = %#v, %v; want ID %s", event, err, tc.WantID)
			}
			if writes != 1 {
				t.Fatalf("writes = %d, want exactly one attempt", writes)
			}
		})
	}
}

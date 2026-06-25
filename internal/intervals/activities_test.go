package intervals

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListActivitiesSendsQueryAndPreservesRawNulls(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345/activities"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		query := r.URL.Query()
		checks := map[string]string{
			"oldest":   "2026-01-01",
			"newest":   "2026-01-31",
			"route_id": "99",
			"limit":    "51",
			"fields":   "id,name,start_date_local,source,_note",
		}
		for key, want := range checks {
			if got := query.Get(key); got != want {
				t.Fatalf("query %s = %q, want %q", key, got, want)
			}
		}
		for _, key := range []string{"routeId", "start", "end"} {
			if got := query.Get(key); got != "" {
				t.Fatalf("query %s = %q, want absent", key, got)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":"a1","name":null,"type":"Run","start_date_local":"2026-01-30T07:00:00","distance":5000,"icu_training_load":42,"stream_types":["time","distance"],"has_weather":true,"average_weather_temp":22.5,"average_wind_speed":4.1,"prevailing_wind_deg":180},
			{"id":"a2","source":"strava","_note":"Strava activity hidden"}
		]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	activities, err := client.ListActivities(context.Background(), ListActivitiesParams{
		Oldest:  "2026-01-01",
		Newest:  "2026-01-31",
		RouteID: 99,
		Limit:   51,
		Fields:  []string{"id", "name", "start_date_local", "source", "_note"},
	})
	if err != nil {
		t.Fatalf("ListActivities() error = %v", err)
	}
	if len(activities) != 2 {
		t.Fatalf("activity count = %d, want 2", len(activities))
	}
	if activities[0].Name != nil {
		t.Fatalf("Name = %q, want nil pointer for upstream null", *activities[0].Name)
	}
	if rawName, ok := activities[0].Raw["name"]; !ok || rawName != nil {
		rawJSON, _ := json.Marshal(activities[0].Raw)
		t.Fatalf("raw name = %#v (present %v), raw = %s; want present nil", rawName, ok, rawJSON)
	}
	if got, want := activities[0].StreamTypes, []string{"time", "distance"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("StreamTypes = %#v, want %#v", got, want)
	}
	if activities[0].HasWeather == nil || !*activities[0].HasWeather || activities[0].AverageWeatherTemp == nil || *activities[0].AverageWeatherTemp != 22.5 || activities[0].AverageWindSpeed == nil || *activities[0].AverageWindSpeed != 4.1 || activities[0].PrevailingWindDeg == nil || *activities[0].PrevailingWindDeg != 180 {
		t.Fatalf("weather fields = %+v, want decoded activity weather", activities[0])
	}
	if activities[1].Source == nil || *activities[1].Source != "strava" {
		t.Fatalf("Source = %#v, want strava", activities[1].Source)
	}
}

func TestListActivitiesRequiresOldest(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.ListActivities(context.Background(), ListActivitiesParams{}); err == nil {
		t.Fatal("ListActivities() error = nil, want required oldest error")
	}
}

func TestLinkActivityToEventUsesPutActivityPairedEventID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/activity/i147866949"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if decoded["paired_event_id"] != float64(1001) || len(decoded) != 1 {
			t.Fatalf("body = %#v, want only paired_event_id", decoded)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i147866949","paired_event_id":1001,"start_date_local":"2026-05-10T07:00:00"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	activity, err := client.LinkActivityToEvent(context.Background(), LinkActivityToEventParams{ActivityID: " i147866949 ", EventID: " 1001 "})
	if err != nil {
		t.Fatalf("LinkActivityToEvent() error = %v", err)
	}
	if activity.ID != "i147866949" || activity.Raw["paired_event_id"] != float64(1001) {
		t.Fatalf("activity = %+v raw=%#v, want linked response", activity, activity.Raw)
	}
}

func TestUpdateActivitySendsSparsePutPayload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/activity/i147866949"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if decoded["name"] != "Threshold ride" || decoded["description"] != "Felt strong; held target W" || len(decoded) != 2 {
			t.Fatalf("body = %#v, want name+description only", decoded)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i147866949","name":"Threshold ride","description":"Felt strong; held target W"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	activity, err := client.UpdateActivity(context.Background(), UpdateActivityParams{
		ActivityID:     " i147866949 ",
		Name:           "Threshold ride",
		NameSet:        true,
		Description:    "Felt strong; held target W",
		DescriptionSet: true,
	})
	if err != nil {
		t.Fatalf("UpdateActivity() error = %v", err)
	}
	if activity.ID != "i147866949" || activity.Name == nil || *activity.Name != "Threshold ride" {
		t.Fatalf("activity = %+v, want updated name", activity)
	}
}

func TestUpdateActivityRequiresAtLeastOneField(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.UpdateActivity(context.Background(), UpdateActivityParams{ActivityID: "a1"}); err == nil {
		t.Fatal("UpdateActivity() error = nil, want validation error")
	}
	if _, err := client.UpdateActivity(context.Background(), UpdateActivityParams{Name: "x", NameSet: true}); err == nil {
		t.Fatal("UpdateActivity() error = nil, want activity ID required")
	}
}

func TestUpdateActivitySendsExplicitEmptyDescription(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var decoded map[string]any
		_ = json.Unmarshal(body, &decoded)
		if _, ok := decoded["description"]; !ok {
			t.Fatalf("body = %s, want description sent", body)
		}
		if decoded["description"] != "" {
			t.Fatalf("description = %#v, want empty string to clear", decoded["description"])
		}
		_, _ = w.Write([]byte(`{"id":"a1","description":""}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	if _, err := client.UpdateActivity(context.Background(), UpdateActivityParams{ActivityID: "a1", DescriptionSet: true}); err != nil {
		t.Fatalf("UpdateActivity() error = %v", err)
	}
}

func TestLinkActivityToEventRequiresIDs(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	for _, params := range []LinkActivityToEventParams{
		{EventID: "1001"},
		{ActivityID: "a1"},
		{ActivityID: "a1", EventID: "evt-1"},
	} {
		if _, err := client.LinkActivityToEvent(context.Background(), params); err == nil {
			t.Fatalf("LinkActivityToEvent(%#v) error = nil, want validation error", params)
		}
	}
}

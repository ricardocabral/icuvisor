package intervals

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetActivitySendsIntervalsFalseAndPreservesRaw(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/activity/a123"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got := r.URL.Query().Get("intervals"); got != "false" {
			t.Fatalf("intervals query = %q, want false", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"a123","name":null,"type":"Run","start_date_local":"2026-01-02T07:00:00"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	activity, err := client.GetActivity(context.Background(), "a123")
	if err != nil {
		t.Fatalf("GetActivity() error = %v", err)
	}
	if activity.ID != "a123" || activity.Name != nil {
		t.Fatalf("activity = %#v, want id and nil Name", activity)
	}
	if rawName, ok := activity.Raw["name"]; !ok || rawName != nil {
		rawJSON, _ := json.Marshal(activity.Raw)
		t.Fatalf("raw name = %#v present %v raw=%s, want preserved null", rawName, ok, rawJSON)
	}
}

func TestGetActivityIntervalsSendsPathAndPreservesRaw(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/activity/a123/intervals"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"a123","analyzed":true,"icu_intervals":[{"id":"i1","name":"Lap","unit":"MINS_KM","pace":4.2,"nullable":null}],"icu_groups":[{"id":"g1","name":"Work"}]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	dto, err := client.GetActivityIntervals(context.Background(), "a123")
	if err != nil {
		t.Fatalf("GetActivityIntervals() error = %v", err)
	}
	if dto.ID != "a123" || !dto.Analyzed || len(dto.ICUIntervals) != 1 || len(dto.ICUGroups) != 1 {
		t.Fatalf("dto = %#v, want interval and group", dto)
	}
	if got := dto.ICUIntervals[0].Raw["nullable"]; got != nil {
		t.Fatalf("raw nullable = %#v, want nil", got)
	}
}

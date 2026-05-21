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

func TestActivityIntervalDecodesNumericAndStringTimeFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		payload   string
		wantStart string
		wantEnd   string
	}{
		{
			name:      "numeric second offsets",
			payload:   `{"id":"i1","start_time":0,"end_time":240.5}`,
			wantStart: "0",
			wantEnd:   "240.5",
		},
		{
			name:      "string time fields",
			payload:   `{"id":"i1","start_time":"2026-05-01T10:00:00","end_time":"2026-05-01T10:04:00"}`,
			wantStart: "2026-05-01T10:00:00",
			wantEnd:   "2026-05-01T10:04:00",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var interval ActivityInterval
			if err := json.Unmarshal([]byte(tc.payload), &interval); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if interval.StartTime == nil || *interval.StartTime != tc.wantStart {
				t.Errorf("StartTime = %v, want %q", interval.StartTime, tc.wantStart)
			}
			if interval.EndTime == nil || *interval.EndTime != tc.wantEnd {
				t.Errorf("EndTime = %v, want %q", interval.EndTime, tc.wantEnd)
			}
		})
	}
}

func TestIntervalsDTODecodesNumericIntervalTimes(t *testing.T) {
	t.Parallel()
	payload := `{"id":"a1","analyzed":true,"icu_intervals":[{"id":"i1","start_time":0,"end_time":240}],"icu_groups":[]}`
	var dto IntervalsDTO
	if err := json.Unmarshal([]byte(payload), &dto); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(dto.ICUIntervals) != 1 {
		t.Fatalf("ICUIntervals len = %d, want 1", len(dto.ICUIntervals))
	}
	got := dto.ICUIntervals[0]
	if got.EndTime == nil || *got.EndTime != "240" {
		t.Errorf("EndTime = %v, want \"240\"", got.EndTime)
	}
	if raw, ok := got.Raw["end_time"].(float64); !ok || raw != 240 {
		t.Errorf("Raw[end_time] = %#v, want numeric 240 for full-payload responses", got.Raw["end_time"])
	}
}

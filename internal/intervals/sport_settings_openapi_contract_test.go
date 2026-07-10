package intervals

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateSportSettingsOpenAPIContract(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		recalcHRZones bool
	}{
		{name: "recalculates HR zones", recalcHRZones: true},
		{name: "does not recalculate HR zones", recalcHRZones: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Fatalf("method = %s, want PUT", r.Method)
				}
				if r.URL.Path != "/athlete/i12345/sport-settings/7" {
					t.Fatalf("path = %q", r.URL.Path)
				}
				if got, want := r.URL.Query().Get("recalcHrZones"), map[bool]string{true: "true", false: "false"}[tc.recalcHRZones]; got != want {
					t.Fatalf("recalcHrZones = %q, want %q", got, want)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if len(body) != 1 || body["ftp"] != float64(275) {
					t.Fatalf("body = %#v, want only ftp", body)
				}
				if _, ok := body["recalcHrZones"]; ok {
					t.Fatalf("body = %#v, must not contain recalcHrZones", body)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":7,"type":"Ride","ftp":275}`))
			}))
			defer server.Close()

			ftp := 275
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
			if _, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{SportSettingID: 7, RecalcHRZones: tc.recalcHRZones, FTP: &ftp}); err != nil {
				t.Fatalf("UpdateSportSettings() error = %v", err)
			}
		})
	}
}

func TestApplySportSettingsOpenAPIContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/athlete/i12345/sport-settings/7/apply" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.RawQuery != "" {
			t.Fatalf("query = %q, want empty", r.URL.RawQuery)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if len(body) != 0 {
			t.Fatalf("body = %q, want empty", body)
		}
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Fatalf("Content-Type = %q, want empty", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	if err := client.ApplySportSettings(context.Background(), 7); err != nil {
		t.Fatalf("ApplySportSettings() error = %v", err)
	}
}

func TestSportSettingsWriteRejectsInvalidID(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "http://example.invalid", http.DefaultClient, RetryConfig{MaxAttempts: 1})
	if _, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{}); err == nil {
		t.Fatal("UpdateSportSettings() error = nil, want invalid ID error")
	}
	if err := client.ApplySportSettings(context.Background(), 0); err == nil {
		t.Fatal("ApplySportSettings() error = nil, want invalid ID error")
	}
}

package intervals

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWellnessUnmarshalExtractsNativeProviders(t *testing.T) {
	t.Parallel()

	var got Wellness
	raw := `{
		"id":"2026-05-01",
		"sleepScore":88,
		"sleepQuality":3,
		"polar":{"ans_charge":4,"nightly_recharge_status":"ok","sleep_score":91},
		"garmin":{"bodyBatteryMin":25,"bodyBatteryMax":76},
		"oura_sleep_score":82,
		"provider":"Oura Ring",
		"sleep_score":83
	}`
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if got.ID == nil || *got.ID != "2026-05-01" || got.SleepScore == nil || *got.SleepScore != 88 || got.SleepQuality == nil || *got.SleepQuality != 3 {
		t.Fatalf("typed wellness = %+v", got)
	}
	if got.Raw["provider"] != "Oura Ring" {
		t.Fatalf("Raw provider = %#v, want preserved raw fields", got.Raw["provider"])
	}
	if got.Native["polar"]["ans_charge"] != float64(4) || got.Native["garmin"]["body_battery_min"] != float64(25) || got.Native["oura"]["sleep_score"] == nil {
		t.Fatalf("Native providers = %#v", got.Native)
	}
	if !containsString(got.NativeClaimedKeys, "polar") || !containsString(got.NativeClaimedKeys, "garmin") || !containsString(got.NativeClaimedKeys, "oura_sleep_score") || !containsString(got.NativeClaimedKeys, "sleep_score") {
		t.Fatalf("NativeClaimedKeys = %#v", got.NativeClaimedKeys)
	}
}

func TestExtractWellnessNativeEmptyWhenNoProviderFields(t *testing.T) {
	t.Parallel()

	native, claimed := extractWellnessNative(map[string]any{"id": "2026-05-01", "sleepScore": float64(90)})
	if native != nil || claimed != nil {
		t.Fatalf("extractWellnessNative() = (%#v, %#v), want nil provider data", native, claimed)
	}
}

func TestNativeSleepScoreSourceAndDedupe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  map[string]any
		want string
	}{
		{name: "polar source", raw: map[string]any{"source": "Polar Flow"}, want: "polar"},
		{name: "oura integration", raw: map[string]any{"integration": "oura cloud"}, want: "oura"},
		{name: "unknown provider", raw: map[string]any{"provider": "manual"}, want: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := nativeSleepScoreSource(tc.raw); got != tc.want {
				t.Fatalf("nativeSleepScoreSource() = %q, want %q", got, tc.want)
			}
		})
	}

	if got := dedupeStrings([]string{"polar", "", "garmin", "polar", "oura", "garmin"}); strings.Join(got, ",") != "polar,garmin,oura" {
		t.Fatalf("dedupeStrings() = %#v, want stable unique non-empty strings", got)
	}
}

func TestListWellnessBuildsQueryAndDecodesNative(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345/wellness.json"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("oldest"), "2026-05-01"; got != want {
			t.Fatalf("oldest query = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("newest"), "2026-05-07"; got != want {
			t.Fatalf("newest query = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("fields"), "sleepScore,polar_sleep_score"; got != want {
			t.Fatalf("fields query = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"2026-05-01","sleepScore":90,"polar_sleep_score":92}]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	got, err := client.ListWellness(context.Background(), WellnessParams{Oldest: " 2026-05-01 ", Newest: " 2026-05-07 ", Fields: []string{"sleepScore", "", " polar_sleep_score "}})
	if err != nil {
		t.Fatalf("ListWellness() error = %v", err)
	}
	if len(got) != 1 || got[0].ID == nil || *got[0].ID != "2026-05-01" || got[0].Native["polar"]["sleep_score"] != float64(92) {
		t.Fatalf("wellness rows = %#v", got)
	}
}

func TestUpdateWellnessSendsSparseBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/athlete/i12345/wellness/2026-05-01"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["feel"] != float64(4) || len(body) != 1 {
			t.Fatalf("body = %#v, want sparse feel-only update", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"2026-05-01","feel":4,"weight":70.5}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	feel := 4
	got, err := client.UpdateWellness(context.Background(), WriteWellnessParams{Date: " 2026-05-01 ", Feel: &feel})
	if err != nil {
		t.Fatalf("UpdateWellness() error = %v", err)
	}
	if got.Feel == nil || *got.Feel != 4 || got.Weight == nil || *got.Weight != 70.5 {
		t.Fatalf("updated wellness = %#v, want decoded row", got)
	}
}

func TestListWellnessRequiresOldest(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	_, err := client.ListWellness(context.Background(), WellnessParams{Oldest: " \t "})
	if err == nil || !strings.Contains(err.Error(), "oldest is required") {
		t.Fatalf("ListWellness() error = %v, want required oldest", err)
	}
}

func TestListWellnessWrapsHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	_, err := client.ListWellness(context.Background(), WellnessParams{Oldest: "2026-05-01"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("ListWellness() error = %v, want ErrUnauthorized", err)
	}
	if !strings.Contains(err.Error(), "listing wellness") {
		t.Fatalf("ListWellness() error = %q, want operation context", err.Error())
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

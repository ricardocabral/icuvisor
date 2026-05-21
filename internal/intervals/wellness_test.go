package intervals

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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
		"garmin":{"bodyBatteryMin":25,"bodyBatteryMax":76,"sleepScore":79},
		"whoop":{"sleepPerformancePercentage":88,"recoveryScore":64},
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
	if got.Native["polar"]["ans_charge"] != float64(4) || got.Native["garmin"]["body_battery_min"] != float64(25) || got.Native["garmin"]["sleep_score"] != float64(79) || got.Native["whoop"]["sleep_performance_percentage"] != float64(88) || got.Native["whoop"]["recovery_score"] != float64(64) || got.Native["oura"]["sleep_score"] == nil {
		t.Fatalf("Native providers = %#v", got.Native)
	}
	if !containsString(got.NativeClaimedKeys, "polar") || !containsString(got.NativeClaimedKeys, "garmin") || !containsString(got.NativeClaimedKeys, "whoop") || !containsString(got.NativeClaimedKeys, "oura_sleep_score") || !containsString(got.NativeClaimedKeys, "sleep_score") {
		t.Fatalf("NativeClaimedKeys = %#v", got.NativeClaimedKeys)
	}
}

func TestWellnessNutritionFixtureDecodesTypedFields(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/wellness/manual_only.json")
	if err != nil {
		t.Fatalf("read manual fixture: %v", err)
	}
	var got Wellness
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if got.KcalConsumed == nil || *got.KcalConsumed != 2400 {
		t.Fatalf("KcalConsumed = %#v, want 2400", got.KcalConsumed)
	}
	for name, tc := range map[string]struct {
		got  *float64
		want float64
	}{
		"carbohydrates": {got: got.Carbohydrates, want: 320.5},
		"protein":       {got: got.Protein, want: 132.25},
		"fatTotal":      {got: got.FatTotal, want: 78.75},
	} {
		if tc.got == nil || *tc.got != tc.want {
			t.Fatalf("%s = %#v, want %v", name, tc.got, tc.want)
		}
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
		{name: "garmin provider", raw: map[string]any{"provider": "Garmin Connect"}, want: "garmin"},
		{name: "oura integration", raw: map[string]any{"integration": "oura cloud"}, want: "oura"},
		{name: "polar source", raw: map[string]any{"source": "Polar Flow"}, want: "polar"},
		{name: "whoop device", raw: map[string]any{"device": "WHOOP 4.0"}, want: "whoop"},
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
		if body["fatigue"] != float64(2) || len(body) != 1 {
			t.Fatalf("body = %#v, want sparse fatigue-only update", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"2026-05-01","fatigue":2,"weight":70.5}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	fatigue := 2
	got, err := client.UpdateWellness(context.Background(), WriteWellnessParams{Date: " 2026-05-01 ", Fatigue: &fatigue})
	if err != nil {
		t.Fatalf("UpdateWellness() error = %v", err)
	}
	if got.Fatigue == nil || *got.Fatigue != 2 || got.Weight == nil || *got.Weight != 70.5 {
		t.Fatalf("updated wellness = %#v, want decoded row", got)
	}
}

func TestUpdateWellnessRejectsUnsupportedFeelBeforeRequest(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"2026-05-01","feel":4}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	feel := 4
	_, err := client.UpdateWellness(context.Background(), WriteWellnessParams{Date: "2026-05-01", Feel: &feel})
	if err == nil || !strings.Contains(err.Error(), "field_not_writable: feel") {
		t.Fatalf("UpdateWellness() error = %v, want unsupported feel error", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want no upstream call for unsupported feel", requests)
	}
}

func TestUpdateWellnessSendsAcceptedSubjectiveFixtureBody(t *testing.T) {
	t.Parallel()

	var fixture struct {
		Method string         `json:"method"`
		Path   string         `json:"path"`
		Body   map[string]any `json:"body"`
	}
	readJSONFixture(t, "testdata/wellness/subjective_write_request.json", &fixture)
	responseBody, err := os.ReadFile("testdata/wellness/subjective_write_response.json")
	if err != nil {
		t.Fatalf("read response fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, fixture.Method; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		wantPath := strings.ReplaceAll(strings.ReplaceAll(fixture.Path, "{athlete_id}", "i12345"), "{date}", "2026-05-01")
		if got := r.URL.Path; got != wantPath {
			t.Fatalf("path = %q, want %q", got, wantPath)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !mapsEqual(body, fixture.Body) {
			t.Fatalf("body = %#v, want fixture body %#v", body, fixture.Body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	fatigue := 2
	soreness := 2
	stress := 2
	mood := 4
	motivation := 4
	sleepQuality := 3
	locked := true
	got, err := client.UpdateWellness(context.Background(), WriteWellnessParams{Date: "2026-05-01", Fatigue: &fatigue, Soreness: &soreness, Stress: &stress, Mood: &mood, Motivation: &motivation, SleepQuality: &sleepQuality, Locked: &locked})
	if err != nil {
		t.Fatalf("UpdateWellness() error = %v", err)
	}
	if got.Fatigue == nil || *got.Fatigue != 2 || got.Locked == nil || !*got.Locked {
		t.Fatalf("updated wellness = %#v, want decoded fixture response", got)
	}
}

func TestUpdateWellnessSendsNewWritableFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		key    string
		params WriteWellnessParams
		want   any
	}{
		{name: "spO2", key: "spO2", params: WriteWellnessParams{SpO2: float64Ptr(97)}, want: float64(97)},
		{name: "vo2max", key: "vo2max", params: WriteWellnessParams{VO2Max: float64Ptr(52.5)}, want: float64(52.5)},
		{name: "abdomen", key: "abdomen", params: WriteWellnessParams{Abdomen: float64Ptr(81.2)}, want: float64(81.2)},
		{name: "respiration", key: "respiration", params: WriteWellnessParams{Respiration: float64Ptr(13.5)}, want: float64(13.5)},
		{name: "menstrualPhase", key: "menstrualPhase", params: WriteWellnessParams{MenstrualPhase: stringPtr("luteal")}, want: "luteal"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				if got := body[tc.key]; got != tc.want || len(body) != 1 {
					t.Fatalf("body = %#v, want sparse %s=%#v update", body, tc.key, tc.want)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"2026-05-01"}`))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
			tc.params.Date = "2026-05-01"
			if _, err := client.UpdateWellness(context.Background(), tc.params); err != nil {
				t.Fatalf("UpdateWellness() error = %v", err)
			}
		})
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

func TestUpdateWellness422ReturnsValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		body      string
		wantField string
	}{
		{
			name:      "plain text unrecognized wellness field",
			body:      "Unrecognized wellness field [WhoopStrain]",
			wantField: "WhoopStrain",
		},
		{
			name:      "json error shape",
			body:      `{"error":"Unknown field: MyCustomField"}`,
			wantField: "MyCustomField",
		},
		{
			name:      "no field name",
			body:      "invalid request",
			wantField: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
			fatigue := 2
			_, err := client.UpdateWellness(context.Background(), WriteWellnessParams{Date: "2026-05-01", Fatigue: &fatigue})
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("UpdateWellness() error = %v, want ErrValidation", err)
			}
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("UpdateWellness() error = %v, want *ValidationError", err)
			}
			if ve.Field != tc.wantField {
				t.Fatalf("ValidationError.Field = %q, want %q", ve.Field, tc.wantField)
			}
			if strings.Contains(err.Error(), tc.body) {
				t.Fatalf("error %q leaked raw upstream body", err)
			}
		})
	}
}

func TestParseValidationErrorExtractsField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		body      string
		wantField string
		wantMsg   string
	}{
		{
			name:      "wellness bracket pattern",
			body:      "Unrecognized wellness field [feel]",
			wantField: "feel",
			wantMsg:   "Unrecognized wellness field [feel]",
		},
		{
			name:      "custom field colon pattern",
			body:      "Unknown field: WhoopStrain",
			wantField: "WhoopStrain",
			wantMsg:   "Unknown field: WhoopStrain",
		},
		{
			name:      "json message shape",
			body:      `{"message":"Field 'MyScore' is not valid"}`,
			wantField: "MyScore",
			wantMsg:   "Field 'MyScore' is not valid",
		},
		{
			name:      "json error shape",
			body:      `{"error":"Unknown field: SomeCustom"}`,
			wantField: "SomeCustom",
			wantMsg:   "Unknown field: SomeCustom",
		},
		{
			name:      "no field name",
			body:      "internal server error",
			wantField: "",
			wantMsg:   "internal server error",
		},
		{
			name:      "empty body",
			body:      "",
			wantField: "",
			wantMsg:   "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ve := parseValidationError(strings.NewReader(tc.body))
			if ve.Field != tc.wantField {
				t.Fatalf("Field = %q, want %q", ve.Field, tc.wantField)
			}
			if ve.UpstreamMessage != tc.wantMsg {
				t.Fatalf("UpstreamMessage = %q, want %q", ve.UpstreamMessage, tc.wantMsg)
			}
		})
	}
}

func readJSONFixture(t *testing.T, path string, out any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decode fixture %s: %v", path, err)
	}
}

func mapsEqual(got map[string]any, want map[string]any) bool {
	return reflect.DeepEqual(got, want)
}

func float64Ptr(value float64) *float64 {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

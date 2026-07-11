package intervals

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateSportSettingsSendsSparseBodyWithoutApply(t *testing.T) {
	t.Parallel()

	var updateBody map[string]any
	updateRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/athlete/i12345/sport-settings/7":
			updateRequests++
			if r.Method != http.MethodPut {
				t.Fatalf("method = %s, want PUT", r.Method)
			}
			if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":7,"type":"Ride","ftp":275,"lthr":171,"threshold_pace":3.5714285,"pace_units":"MINS_KM","pace_load_type":"RUN"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	ftp := 275
	lthr := 171
	pace := SportSettingsPace{Value: 3.5714285, PaceUnits: "MINS_KM", PaceLoadType: "RUN"}
	got, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{SportSettingID: 7, RecalcHRZones: true, FTP: &ftp, ThresholdHR: &lthr, ThresholdPace: &pace})
	if err != nil {
		t.Fatalf("UpdateSportSettings() error = %v", err)
	}
	if updateRequests != 1 {
		t.Fatalf("update requests = %d, want exactly one with no implicit apply", updateRequests)
	}
	if got.ID != 7 || got.Type != "Ride" || got.FTP != 275 || got.LTHR != 171 || got.ThresholdPace != 3.5714285 || got.PaceLoadType != "RUN" {
		t.Fatalf("updated setting = %+v", got)
	}
	if updateBody["ftp"] != float64(275) || updateBody["lthr"] != float64(171) || updateBody["threshold_pace"] != float64(3.5714285) || updateBody["pace_units"] != "MINS_KM" || updateBody["pace_load_type"] != "RUN" {
		t.Fatalf("update body = %#v, want sparse thresholds", updateBody)
	}
	if updateBody["power_zones"] != nil || updateBody["hr_zones"] != nil || updateBody["pace_zones"] != nil {
		t.Fatalf("update body = %#v, want no zone fields when zones omitted", updateBody)
	}
}

func TestUpdateSportSettingsSendsOnlyIndoorFTP(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/athlete/i12345/sport-settings/7" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.RawQuery != "recalcHrZones=false" {
			t.Fatalf("raw query = %q, want recalcHrZones=false", r.URL.RawQuery)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if len(body) != 1 || body["indoor_ftp"] != float64(245) {
			t.Fatalf("body = %#v, want only indoor_ftp", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"type":"Ride","indoor_ftp":245}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	indoorFTP := 245
	got, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{SportSettingID: 7, IndoorFTP: &indoorFTP})
	if err != nil {
		t.Fatalf("UpdateSportSettings() error = %v", err)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
	if got.IndoorFTP != indoorFTP {
		t.Fatalf("returned indoor FTP = %d, want %d", got.IndoorFTP, indoorFTP)
	}
}

func TestCreateSportSettingsWireMatrixSendsCompleteSparseBodies(t *testing.T) {
	t.Parallel()

	ftp, indoorFTP := 285, 265
	tests := []struct {
		name         string
		params       CreateSportSettingsParams
		wantBody     map[string]any
		wantResponse string
	}{
		{
			name:         "Ride with FTP and indoor FTP",
			params:       CreateSportSettingsParams{Sport: "Ride", FTP: &ftp, IndoorFTP: &indoorFTP},
			wantBody:     map[string]any{"ftp": float64(285), "indoor_ftp": float64(265)},
			wantResponse: `{"id":8,"type":"Ride","ftp":285,"indoor_ftp":265}`,
		},
		{
			name: "Run with canonical threshold pace",
			params: CreateSportSettingsParams{Sport: "Run", ThresholdPace: &SportSettingsPace{
				Value: 1000.0 / 300, PaceUnits: "MINS_KM", PaceLoadType: "RUN",
			}},
			wantBody:     map[string]any{"threshold_pace": 1000.0 / 300, "pace_units": "MINS_KM", "pace_load_type": "RUN"},
			wantResponse: `{"id":9,"type":"Run","threshold_pace":3.3333333333333335,"pace_units":"MINS_KM","pace_load_type":"RUN"}`,
		},
		{
			name: "Swim with canonical threshold pace",
			params: CreateSportSettingsParams{Sport: "Swim", ThresholdPace: &SportSettingsPace{
				Value: 100.0 / 90, PaceUnits: "SECS_100M", PaceLoadType: "SWIM",
			}},
			wantBody:     map[string]any{"threshold_pace": 100.0 / 90, "pace_units": "SECS_100M", "pace_load_type": "SWIM"},
			wantResponse: `{"id":10,"type":"Swim","threshold_pace":1.1111111111111112,"pace_units":"SECS_100M","pace_load_type":"SWIM"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodPost {
					t.Fatalf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != "/athlete/i12345/sport-settings" {
					t.Fatalf("path = %s, want create sport-settings path", r.URL.Path)
				}
				if r.URL.RawQuery != "" {
					t.Fatalf("raw query = %q, want empty", r.URL.RawQuery)
				}

				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				types, ok := body["types"].([]any)
				if !ok || len(types) != 1 || types[0] != tc.params.Sport {
					t.Fatalf("types = %#v, want [%s]", body["types"], tc.params.Sport)
				}
				if len(body) != len(tc.wantBody)+1 {
					t.Fatalf("body = %#v, want complete sparse body %#v plus types", body, tc.wantBody)
				}
				for key, want := range tc.wantBody {
					if got, exists := body[key]; !exists || got != want {
						t.Fatalf("body[%q] = %#v, want %#v in %#v", key, got, want, body)
					}
				}
				for _, forbidden := range []string{"id", "sport_setting_id", "recalcHrZones", "recalc_hr_zones", "apply", "power_zones", "power_zone_names", "hr_zones", "hr_zone_names", "pace_zones", "pace_zone_names"} {
					if _, exists := body[forbidden]; exists {
						t.Fatalf("body = %#v, must not contain %q", body, forbidden)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.wantResponse))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
			if _, err := client.CreateSportSettings(context.Background(), tc.params); err != nil {
				t.Fatalf("CreateSportSettings() error = %v", err)
			}
			if requests != 1 {
				t.Fatalf("requests = %d, want exactly one", requests)
			}
		})
	}
}

func TestSportSettingsThresholdValidationAvoidsHTTP(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatal("invalid parameters must not make an HTTP request")
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	zero := 0
	negative := -1
	nanPace := SportSettingsPace{Value: math.NaN()}
	infinitePace := SportSettingsPace{Value: math.Inf(1)}
	for _, tc := range []struct {
		name      string
		operation string
		call      func() error
	}{
		{name: "update FTP", operation: "updating sport settings", call: func() error {
			_, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{SportSettingID: 7, FTP: &zero})
			return err
		}},
		{name: "update indoor FTP", operation: "updating sport settings", call: func() error {
			_, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{SportSettingID: 7, IndoorFTP: &negative})
			return err
		}},
		{name: "update threshold HR", operation: "updating sport settings", call: func() error {
			_, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{SportSettingID: 7, ThresholdHR: &zero})
			return err
		}},
		{name: "update non-finite pace", operation: "updating sport settings", call: func() error {
			_, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{SportSettingID: 7, ThresholdPace: &nanPace})
			return err
		}},
		{name: "create blank sport", operation: "creating sport settings", call: func() error {
			_, err := client.CreateSportSettings(context.Background(), CreateSportSettingsParams{Sport: "  "})
			return err
		}},
		{name: "create FTP", operation: "creating sport settings", call: func() error {
			_, err := client.CreateSportSettings(context.Background(), CreateSportSettingsParams{Sport: "Ride", FTP: &zero})
			return err
		}},
		{name: "create indoor FTP", operation: "creating sport settings", call: func() error {
			_, err := client.CreateSportSettings(context.Background(), CreateSportSettingsParams{Sport: "Ride", IndoorFTP: &negative})
			return err
		}},
		{name: "create threshold HR", operation: "creating sport settings", call: func() error {
			_, err := client.CreateSportSettings(context.Background(), CreateSportSettingsParams{Sport: "Ride", ThresholdHR: &zero})
			return err
		}},
		{name: "create non-finite pace", operation: "creating sport settings", call: func() error {
			_, err := client.CreateSportSettings(context.Background(), CreateSportSettingsParams{Sport: "Ride", ThresholdPace: &infinitePace})
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if err == nil || !strings.Contains(err.Error(), tc.operation) {
				t.Fatalf("error = %v, want %q validation error", err, tc.operation)
			}
		})
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestCreateSportSettingsAllowsIndoorFTPAboveFTP(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["ftp"] != float64(200) || body["indoor_ftp"] != float64(300) {
			t.Fatalf("body = %#v, want independent valid FTP values", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":8,"type":"Ride","ftp":200,"indoor_ftp":300}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	ftp, indoorFTP := 200, 300
	if _, err := client.CreateSportSettings(context.Background(), CreateSportSettingsParams{Sport: "Ride", FTP: &ftp, IndoorFTP: &indoorFTP}); err != nil {
		t.Fatalf("CreateSportSettings() error = %v", err)
	}
}

func TestUpdateSportSettingsSendsZoneOverwriteFieldsWhenProvided(t *testing.T) {
	t.Parallel()

	var updateBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/athlete/i12345/sport-settings/7" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
			t.Fatalf("decode update body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"type":"Ride","ftp":275}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	_, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{SportSettingID: 7, RecalcHRZones: true, ZonesProvided: true, Zones: []SportSettingsZoneDefinition{{Kind: "power", Boundaries: []float64{100.2, 200.8}, Names: []string{"Z1", "Z2"}}}})
	if err != nil {
		t.Fatalf("UpdateSportSettings() error = %v", err)
	}
	if got := updateBody["power_zones"].([]any); len(got) != 2 || got[0] != float64(100) || got[1] != float64(200) {
		t.Fatalf("power_zones = %#v, want rounded integer boundaries", updateBody["power_zones"])
	}
	if got := updateBody["power_zone_names"].([]any); len(got) != 2 || got[0] != "Z1" || got[1] != "Z2" {
		t.Fatalf("power_zone_names = %#v", updateBody["power_zone_names"])
	}
}

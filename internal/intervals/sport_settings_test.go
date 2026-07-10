package intervals

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateSportSettingsSendsSparseBodyWithoutApply(t *testing.T) {
	t.Parallel()

	var updateBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/athlete/i12345/sport-settings/7":
			if r.Method != http.MethodPut {
				t.Fatalf("method = %s, want PUT", r.Method)
			}
			if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":7,"type":"Ride","ftp":275,"lthr":171,"threshold_pace":255,"pace_units":"MINS_KM"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	ftp := 275
	lthr := 171
	pace := SportSettingsPace{Value: 255, Unit: "MINS_KM"}
	got, err := client.UpdateSportSettings(context.Background(), WriteSportSettingsParams{SportSettingID: 7, RecalcHRZones: true, FTP: &ftp, ThresholdHR: &lthr, ThresholdPace: &pace})
	if err != nil {
		t.Fatalf("UpdateSportSettings() error = %v", err)
	}
	if got.ID != 7 || got.Type != "Ride" || got.FTP != 275 || got.LTHR != 171 || got.ThresholdPace != 255 {
		t.Fatalf("updated setting = %+v", got)
	}
	if updateBody["ftp"] != float64(275) || updateBody["lthr"] != float64(171) || updateBody["threshold_pace"] != float64(255) || updateBody["pace_units"] != "MINS_KM" {
		t.Fatalf("update body = %#v, want sparse thresholds", updateBody)
	}
	if updateBody["power_zones"] != nil || updateBody["hr_zones"] != nil || updateBody["pace_zones"] != nil {
		t.Fatalf("update body = %#v, want no zone fields when zones omitted", updateBody)
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

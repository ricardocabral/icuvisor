package tools

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type fakeSportSettingsWriterClient struct {
	fakeProfileClient
	setting intervals.SportSettings
	calls   []intervals.WriteSportSettingsParams
	err     error
}

func (f *fakeSportSettingsWriterClient) UpdateSportSettings(ctx context.Context, params intervals.WriteSportSettingsParams) (intervals.SportSettings, error) {
	f.calls = append(f.calls, params)
	return f.setting, f.err
}

func TestUpdateSportSettingsSchemaDocumentsInputsAndZoneGate(t *testing.T) {
	t.Parallel()

	tool := newUpdateSportSettingsTool(&fakeSportSettingsWriterClient{}, &fakeProfileClient{}, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
	if !strings.Contains(tool.Description, "get_athlete_profile _meta.warnings") || strings.Contains(strings.ToLower(tool.Description), "credential") {
		t.Fatalf("description = %q, want profile-warning guidance without credential wording", tool.Description)
	}
	schema := tool.InputSchema.(map[string]any)
	required := schema["required"].([]string)
	if !containsString(required, "sport") || !containsString(required, "effective_date") {
		t.Fatalf("required = %#v, want sport and effective_date", required)
	}
	props := schema["properties"].(map[string]any)
	for _, field := range []string{"sport", "effective_date", "ftp", "threshold_hr", "threshold_pace", "zones"} {
		if _, ok := props[field]; !ok {
			t.Fatalf("schema missing field %s", field)
		}
	}
	sport := props["sport"].(map[string]any)
	if len(sport["enum"].([]string)) == 0 || !containsString(sport["enum"].([]string), "Ride") || !containsString(sport["enum"].([]string), "Run") {
		t.Fatalf("sport enum = %#v, want Ride/Run", sport["enum"])
	}
	pace := props["threshold_pace"].(map[string]any)
	if !strings.Contains(pace["description"].(string), "seconds_per_mile") || !strings.Contains(pace["description"].(string), "8:00/mi") {
		t.Fatalf("threshold_pace description = %q, want km/mile pace examples", pace["description"])
	}
	paceProps := pace["properties"].(map[string]any)
	unitEnum := paceProps["unit"].(map[string]any)["enum"].([]string)
	if !containsString(unitEnum, "seconds_per_km") || !containsString(unitEnum, "seconds_per_mile") {
		t.Fatalf("threshold_pace unit enum = %#v, want seconds per km/mile", unitEnum)
	}
	zones := props["zones"].(map[string]any)
	if !strings.Contains(zones["description"].(string), "overwrites prior") || !strings.Contains(zones["description"].(string), "ICUVISOR_DELETE_MODE=full") {
		t.Fatalf("zones description = %q, want overwrite gate warning", zones["description"])
	}
	zoneProps := zones["items"].(map[string]any)["properties"].(map[string]any)
	boundaryDescription := zoneProps["boundaries"].(map[string]any)["description"].(string)
	if !strings.Contains(boundaryDescription, "seconds_per_km") || !strings.Contains(boundaryDescription, "seconds_per_mile") || !strings.Contains(boundaryDescription, "not speed") {
		t.Fatalf("zone boundary description = %q, want explicit pace-duration wording", boundaryDescription)
	}
}

func TestUpdateSportSettingsThresholdFieldsAndPaceConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          string
		wantFTP       *int
		wantHR        *int
		wantPace      bool
		wantPaceValue float64
		wantInputUnit string
		wantFields    []string
	}{
		{name: "ftp only", args: `{"sport":"Run","effective_date":"2026-05-01","ftp":290}`, wantFTP: intPtr(290), wantFields: []string{"ftp"}},
		{name: "threshold hr only", args: `{"sport":"Run","effective_date":"2026-05-01","threshold_hr":171}`, wantHR: intPtr(171), wantFields: []string{"threshold_hr"}},
		{name: "threshold pace seconds per km converts to upstream mile pace", args: `{"sport":"Run","effective_date":"2026-05-01","threshold_pace":{"value":300,"unit":"seconds_per_km"}}`, wantPace: true, wantPaceValue: 482.8032, wantInputUnit: "seconds_per_km", wantFields: []string{"threshold_pace"}},
		{name: "threshold pace seconds per mile stays in upstream mile pace", args: `{"sport":"Run","effective_date":"2026-05-01","threshold_pace":{"value":480,"unit":"seconds_per_mile"}}`, wantPace: true, wantPaceValue: 480, wantInputUnit: "seconds_per_mile", wantFields: []string{"threshold_pace"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newFakeSportSettingsClient(intervals.SportSettings{ID: 8, Types: []string{"Run"}, PaceUnits: "MINS_MILE"})
			client.setting = intervals.SportSettings{ID: 8, Type: "Run", FTP: valueOrZero(tc.wantFTP), FTHR: valueOrZero(tc.wantHR), PaceUnits: "MINS_MILE"}
			tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(client.calls) != 1 {
				t.Fatalf("write calls = %d, want 1", len(client.calls))
			}
			call := client.calls[0]
			if !sameIntPtr(call.FTP, tc.wantFTP) || !sameIntPtr(call.ThresholdHR, tc.wantHR) {
				t.Fatalf("write call = %+v, want ftp=%v threshold_hr=%v", call, tc.wantFTP, tc.wantHR)
			}
			if tc.wantPace {
				if call.ThresholdPace == nil || call.ThresholdPace.Unit != "MINS_MILE" || math.Abs(call.ThresholdPace.Value-tc.wantPaceValue) > 0.0001 {
					t.Fatalf("threshold pace call = %+v, want %.4f sec/mile", call.ThresholdPace, tc.wantPaceValue)
				}
			} else if call.ThresholdPace != nil {
				t.Fatalf("threshold pace call = %+v, want nil", call.ThresholdPace)
			}
			out := resultMap(t, result)
			settings := out["sport_settings"].(map[string]any)
			meta := out["_meta"].(map[string]any)
			fields := meta["fields_updated"].([]any)
			if len(fields) != len(tc.wantFields) || fields[0] != tc.wantFields[0] || meta["recompute_pending"] != true || meta["zones_provided"] != false {
				t.Fatalf("meta = %#v, want fields %v, recompute_pending, and zones_provided=false", meta, tc.wantFields)
			}
			assertKeyAbsent(t, settings, "zone_definitions_overwritten")
			if tc.wantPace && (meta["pace_input_unit"] != tc.wantInputUnit || meta["pace_upstream_unit"] != "MINS_MILE") {
				t.Fatalf("meta = %#v, want pace conversion metadata", meta)
			}
		})
	}
}

func TestUpdateSportSettingsRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 7, Types: []string{"Ride"}})
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
}

func newFakeSportSettingsClient(setting intervals.SportSettings) *fakeSportSettingsWriterClient {
	return &fakeSportSettingsWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC", SportSettings: []intervals.SportSettings{setting}}},
		setting:           setting,
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

func intPtr(value int) *int {
	return &value
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func sameIntPtr(got *int, want *int) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}
	return *got == *want
}

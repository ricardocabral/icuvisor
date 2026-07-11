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
	if len(required) != 1 || !containsString(required, "sport") {
		t.Fatalf("required = %#v, want only sport", required)
	}
	props := schema["properties"].(map[string]any)
	for _, field := range []string{"sport", "recalc_hr_zones", "ftp", "indoor_ftp", "threshold_hr", "threshold_pace", "zones"} {
		if _, ok := props[field]; !ok {
			t.Fatalf("schema missing field %s", field)
		}
	}
	recalcHRZones := props["recalc_hr_zones"].(map[string]any)
	if recalcHRZones["default"] != true || !strings.Contains(recalcHRZones["description"].(string), "defaults to true") {
		t.Fatalf("recalc_hr_zones = %#v, want documented true default", recalcHRZones)
	}
	sport := props["sport"].(map[string]any)
	if len(sport["enum"].([]string)) == 0 || !containsString(sport["enum"].([]string), "Ride") || !containsString(sport["enum"].([]string), "Run") {
		t.Fatalf("sport enum = %#v, want Ride/Run", sport["enum"])
	}
	pace := props["threshold_pace"].(map[string]any)
	if !strings.Contains(pace["description"].(string), "seconds_per_mile") || !strings.Contains(pace["description"].(string), "8:00/mi") || !strings.Contains(pace["description"].(string), "seconds_per_100y") {
		t.Fatalf("threshold_pace description = %q, want km/mile/yards pace examples", pace["description"])
	}
	paceProps := pace["properties"].(map[string]any)
	unitEnum := paceProps["unit"].(map[string]any)["enum"].([]string)
	if !containsString(unitEnum, "seconds_per_km") || !containsString(unitEnum, "seconds_per_mile") || !containsString(unitEnum, "seconds_per_100y") {
		t.Fatalf("threshold_pace unit enum = %#v, want seconds per km/mile/100y", unitEnum)
	}
	zones := props["zones"].(map[string]any)
	if !strings.Contains(zones["description"].(string), "overwrites prior") || !strings.Contains(zones["description"].(string), "ICUVISOR_DELETE_MODE=full") {
		t.Fatalf("zones description = %q, want overwrite gate warning", zones["description"])
	}
	zoneProps := zones["items"].(map[string]any)["properties"].(map[string]any)
	boundaryDescription := zoneProps["boundaries"].(map[string]any)["description"].(string)
	if !strings.Contains(boundaryDescription, "percent-of-threshold-pace") || !strings.Contains(boundaryDescription, "77.5") || !strings.Contains(boundaryDescription, "never durations") {
		t.Fatalf("zone boundary description = %q, want explicit pace-percentage wording", boundaryDescription)
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
		wantEchoValue float64
		wantInputUnit string
		wantFields    []string
		wantRecalc    bool
	}{
		{name: "ftp defaults HR zone recalculation to true", args: `{"sport":"Run","ftp":290}`, wantFTP: intPtr(290), wantFields: []string{"ftp"}, wantRecalc: true},
		{name: "threshold hr does not recalculate HR zones when false", args: `{"sport":"Run","recalc_hr_zones":false,"threshold_hr":171}`, wantHR: intPtr(171), wantFields: []string{"threshold_hr"}, wantRecalc: false},
		{name: "threshold pace seconds per km writes mps with preserved mile display", args: `{"sport":"Run","threshold_pace":{"value":300,"unit":"seconds_per_km"}}`, wantPace: true, wantPaceValue: 1000.0 / 300, wantEchoValue: 482.8032, wantInputUnit: "seconds_per_km", wantFields: []string{"threshold_pace"}, wantRecalc: true},
		{name: "threshold pace seconds per mile writes mps with preserved mile display", args: `{"sport":"Run","threshold_pace":{"value":480,"unit":"seconds_per_mile"}}`, wantPace: true, wantPaceValue: 1609.344 / 480, wantEchoValue: 480, wantInputUnit: "seconds_per_mile", wantFields: []string{"threshold_pace"}, wantRecalc: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newFakeSportSettingsClient(intervals.SportSettings{ID: 8, Types: []string{"Run"}, PaceUnits: "MINS_MILE", PaceLoadType: "RUN"})
			client.setting = intervals.SportSettings{ID: 8, Type: "Run", FTP: valueOrZero(tc.wantFTP), FTHR: valueOrZero(tc.wantHR), PaceUnits: "MINS_MILE", PaceLoadType: "RUN"}
			tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(client.calls) != 1 {
				t.Fatalf("write calls = %d, want 1", len(client.calls))
			}
			call := client.calls[0]
			if !sameIntPtr(call.FTP, tc.wantFTP) || !sameIntPtr(call.ThresholdHR, tc.wantHR) || call.RecalcHRZones != tc.wantRecalc {
				t.Fatalf("write call = %+v, want ftp=%v threshold_hr=%v recalc_hr_zones=%t", call, tc.wantFTP, tc.wantHR, tc.wantRecalc)
			}
			if tc.wantPace {
				if call.ThresholdPace == nil || call.ThresholdPace.PaceUnits != "MINS_MILE" || call.ThresholdPace.PaceLoadType != "RUN" || math.Abs(call.ThresholdPace.Value-tc.wantPaceValue) > 0.0001 {
					t.Fatalf("threshold pace call = %+v, want %.4f m/s with MINS_MILE/RUN metadata", call.ThresholdPace, tc.wantPaceValue)
				}
			} else if call.ThresholdPace != nil {
				t.Fatalf("threshold pace call = %+v, want nil", call.ThresholdPace)
			}
			out := resultMap(t, result)
			settings := out["sport_settings"].(map[string]any)
			meta := out["_meta"].(map[string]any)
			fields := meta["fields_updated"].([]any)
			if len(fields) != len(tc.wantFields) || fields[0] != tc.wantFields[0] || meta["hr_zone_recalculation_requested"] != tc.wantRecalc || meta["zones_provided"] != false {
				t.Fatalf("meta = %#v, want fields %v, hr_zone_recalculation_requested=%t, and zones_provided=false", meta, tc.wantFields, tc.wantRecalc)
			}
			assertKeyAbsent(t, settings, "zone_definitions_overwritten")
			if tc.wantPace {
				if math.Abs(settings["threshold_pace_seconds_per_mile"].(float64)-tc.wantEchoValue) > 0.0001 || settings["pace_units_source"] != "MINS_MILE" || settings["pace_load_type"] != "RUN" {
					t.Fatalf("settings = %#v, want m/s params rendered as the selected mile display", settings)
				}
				assertKeyAbsent(t, settings, "threshold_pace_meters_per_second")
			}
			if tc.wantPace && (meta["pace_input_unit"] != tc.wantInputUnit || meta["pace_display_unit"] != "MINS_MILE" || meta["pace_load_type"] != "RUN") {
				t.Fatalf("meta = %#v, want pace conversion metadata", meta)
			}
		})
	}
}

func TestUpdateSportSettingsWritesIndoorFTP(t *testing.T) {
	t.Parallel()

	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 7, Types: []string{"Ride"}})
	client.setting = intervals.SportSettings{ID: 7, Type: "Ride", IndoorFTP: 260}
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Ride","indoor_ftp":260}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || !sameIntPtr(client.calls[0].IndoorFTP, intPtr(260)) {
		t.Fatalf("write calls = %#v, want indoor FTP 260", client.calls)
	}
	out := resultMap(t, result)
	settings := out["sport_settings"].(map[string]any)
	if settings["indoor_ftp_watts"] != float64(260) {
		t.Fatalf("settings = %#v, want indoor_ftp_watts=260", settings)
	}
	meta := out["_meta"].(map[string]any)
	if fields := meta["fields_updated"].([]any); len(fields) != 1 || fields[0] != "indoor_ftp" {
		t.Fatalf("fields_updated = %#v, want [indoor_ftp]", fields)
	}
}

func TestUpdateSportSettingsWritesYardSwimPace(t *testing.T) {
	t.Parallel()

	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 9, Types: []string{"Swim"}, PaceUnits: "SECS_100M"})
	client.setting = intervals.SportSettings{ID: 9, Type: "Swim", PaceUnits: "SECS_100M"}
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Swim","threshold_pace":{"value":90,"unit":"seconds_per_100y"}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.ThresholdPace == nil || call.ThresholdPace.PaceUnits != "SECS_100M" || call.ThresholdPace.PaceLoadType != "SWIM" || math.Abs(call.ThresholdPace.Value-(91.44/90)) > 0.000001 {
		t.Fatalf("threshold pace call = %+v, want m/s with preserved SECS_100M/SWIM metadata", call.ThresholdPace)
	}
	out := resultMap(t, result)
	settings := out["sport_settings"].(map[string]any)
	if math.Abs(settings["threshold_pace_seconds_per_100m"].(float64)-(100/(91.44/90))) > 0.000001 || settings["pace_units_source"] != "SECS_100M" || settings["pace_load_type"] != "SWIM" {
		t.Fatalf("settings = %#v, want explicit sec/100m echo from m/s", settings)
	}
	assertKeyAbsent(t, settings, "threshold_pace_meters_per_second")
	meta := out["_meta"].(map[string]any)
	if meta["pace_input_unit"] != "seconds_per_100y" || meta["pace_display_unit"] != "SECS_100M" || meta["pace_load_type"] != "SWIM" {
		t.Fatalf("meta = %#v, want yard pace metadata", meta)
	}
}

func TestUpdateSportSettingsUsesReturnedMPSForPaceEcho(t *testing.T) {
	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 8, Types: []string{"Run"}, PaceUnits: "MINS_KM", PaceLoadType: "RUN"})
	client.setting = intervals.SportSettings{ID: 8, Type: "Run", ThresholdPace: 3.5714285, PaceUnits: "MINS_KM", PaceLoadType: "RUN"}
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Run","threshold_pace":{"value":300,"unit":"seconds_per_km"}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].ThresholdPace == nil || math.Abs(client.calls[0].ThresholdPace.Value-(1000.0/300)) > 0.0001 {
		t.Fatalf("write calls = %#v, want m/s request value", client.calls)
	}
	settings := resultMap(t, result)["sport_settings"].(map[string]any)
	if math.Abs(settings["threshold_pace_seconds_per_km"].(float64)-280) > 0.0001 || settings["pace_units_source"] != "MINS_KM" || settings["pace_load_type"] != "RUN" {
		t.Fatalf("settings = %#v, want returned 3.5714285 m/s rendered as 280 s/km", settings)
	}
}

func TestUpdateSportSettingsInfersValidDisplayAndLoadMetadata(t *testing.T) {
	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 8, Types: []string{"Run"}, PaceUnits: "NONE"})
	client.setting = intervals.SportSettings{ID: 8, Type: "Run"}
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Run","threshold_pace":{"value":280,"unit":"seconds_per_km"}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	call := client.calls[0].ThresholdPace
	if call == nil || math.Abs(call.Value-3.5714285714285716) > 0.000001 || call.PaceUnits != "MINS_KM" || call.PaceLoadType != "RUN" {
		t.Fatalf("threshold pace call = %+v, want inferred m/s display/load metadata", call)
	}
	settings := resultMap(t, result)["sport_settings"].(map[string]any)
	if settings["threshold_pace_seconds_per_km"] != float64(280) || settings["pace_units_source"] != "MINS_KM" || settings["pace_load_type"] != "RUN" {
		t.Fatalf("settings = %#v, want inferred display metadata", settings)
	}
}

func TestShapeUpdateSportSettingsPaceFallbacksUseMPS(t *testing.T) {
	value := 3.5714285
	for _, paceUnits := range []string{"NONE", "FEET"} {
		t.Run(paceUnits, func(t *testing.T) {
			payload := shapeUpdateSportSettingsResponse(updateSportSettingsRequest{Sport: "Run"}, intervals.WriteSportSettingsParams{SportSettingID: 8}, intervals.SportSettings{ID: 8, ThresholdPace: value, PaceUnits: paceUnits, PaceLoadType: "RUN"}, updateSportSettingsMeta{})
			echo := payload.SportSettings
			if echo.ThresholdPaceMetersPerSecond == nil || *echo.ThresholdPaceMetersPerSecond != value || echo.PaceUnitsSource != paceUnits || echo.PaceLoadType != "RUN" {
				t.Fatalf("pace fallback = %+v, want unambiguous m/s response", echo)
			}
			if echo.ThresholdPaceSecondsPerKM != nil || echo.ThresholdPaceSecondsPerMile != nil || echo.ThresholdPaceSecondsPer100M != nil {
				t.Fatalf("pace fallback = %+v, want no duration fields", echo)
			}
		})
	}
}

func TestUpdateSportSettingsRejectsLegacyAndUnknownArgumentsBeforeClients(t *testing.T) {
	t.Parallel()

	for _, args := range []string{
		`{"sport":"Ride","effective_date":"2026-05-01","ftp":290}`,
		`{"sport":"Ride","ftp":290,"unknown":true}`,
		`{"sport":"Ride","ftp":290,"recalc_hr_zones":null}`,
	} {
		t.Run(args, func(t *testing.T) {
			client := newFakeSportSettingsClient(intervals.SportSettings{ID: 7, Types: []string{"Ride"}})
			tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(args)})
			if err == nil || err.Error() != invalidUpdateSportSettingsArgumentsMessage {
				t.Fatalf("Handler() error = %v, want terse validation error", err)
			}
			if client.fakeProfileClient.calls != 0 || len(client.calls) != 0 {
				t.Fatalf("profile calls = %d, writer calls = %d, want neither", client.fakeProfileClient.calls, len(client.calls))
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

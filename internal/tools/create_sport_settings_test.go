package tools

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeSportSettingsCreatorClient struct {
	fakeProfileClient
	setting intervals.SportSettings
	calls   []intervals.CreateSportSettingsParams
	err     error
}

func (f *fakeSportSettingsCreatorClient) CreateSportSettings(_ context.Context, params intervals.CreateSportSettingsParams) (intervals.SportSettings, error) {
	f.calls = append(f.calls, params)
	return f.setting, f.err
}

func TestCreateSportSettingsSchemaDocumentsThresholdOnlyCreation(t *testing.T) {
	t.Parallel()

	tool := newCreateSportSettingsTool(&fakeSportSettingsCreatorClient{}, &fakeProfileClient{}, "test", "UTC", false)
	if !strings.Contains(tool.Description, "cannot replace zones") || !strings.Contains(tool.Description, "does not yet have settings") {
		t.Fatalf("description = %q, want threshold-only missing-setting guidance", tool.Description)
	}
	if tool.Requirement != RequirementWrite || tool.EffectiveToolset().String() != "full" {
		t.Fatalf("tool = %#v, want full write tool", tool)
	}
	schema := tool.InputSchema.(map[string]any)
	if required := schema["required"].([]string); len(required) != 1 || required[0] != "sport" {
		t.Fatalf("required = %#v, want only sport", required)
	}
	props := schema["properties"].(map[string]any)
	for _, field := range []string{"sport", "ftp", "indoor_ftp", "threshold_hr", "threshold_pace"} {
		if _, ok := props[field]; !ok {
			t.Fatalf("schema missing field %q", field)
		}
	}
}

func TestCreateSportSettingsSchemaExcludesCredentialsAndUnsafeControls(t *testing.T) {
	t.Parallel()

	tool := newCreateSportSettingsTool(&fakeSportSettingsCreatorClient{}, &fakeProfileClient{}, "test", "UTC", false)
	assertCreateSportSettingsSchemaSafe(t, tool.InputSchema.(map[string]any), false)
}

func TestCreateSportSettingsCreatesThresholdOnlySettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		arguments     string
		created       intervals.SportSettings
		wantFTP       *int
		wantIndoorFTP *int
		wantHR        *int
		wantPace      float64
		wantUnits     string
		wantLoadType  string
		wantEchoKey   string
		wantEchoValue float64
		wantFields    []string
	}{
		{
			name:          "ride FTP and indoor FTP",
			arguments:     `{"sport":"Ride","ftp":285,"indoor_ftp":265}`,
			created:       intervals.SportSettings{ID: 12, Type: "Ride", FTP: 285, IndoorFTP: 265},
			wantFTP:       intPtr(285),
			wantIndoorFTP: intPtr(265),
			wantFields:    []string{"ftp", "indoor_ftp"},
		},
		{
			name:       "ride threshold heart rate",
			arguments:  `{"sport":"Ride","threshold_hr":171}`,
			created:    intervals.SportSettings{ID: 13, Type: "Ride", LTHR: 171},
			wantHR:     intPtr(171),
			wantFields: []string{"threshold_hr"},
		},
		{
			name:          "run threshold pace",
			arguments:     `{"sport":"Run","threshold_pace":{"value":300,"unit":"seconds_per_km"}}`,
			created:       intervals.SportSettings{ID: 14, Type: "Run", ThresholdPace: 1000.0 / 280, PaceUnits: "MINS_KM", PaceLoadType: "RUN"},
			wantPace:      1000.0 / 300,
			wantUnits:     "MINS_KM",
			wantLoadType:  "RUN",
			wantEchoKey:   "threshold_pace_seconds_per_km",
			wantEchoValue: 280,
			wantFields:    []string{"threshold_pace"},
		},
		{
			name:          "swim threshold pace",
			arguments:     `{"sport":"Swim","threshold_pace":{"value":90,"unit":"seconds_per_100y"}}`,
			created:       intervals.SportSettings{ID: 15, Type: "Swim", ThresholdPace: 100.0 / 85, PaceUnits: "SECS_100M", PaceLoadType: "SWIM"},
			wantPace:      metersPer100Yards / 90,
			wantUnits:     "SECS_100Y",
			wantLoadType:  "SWIM",
			wantEchoKey:   "threshold_pace_seconds_per_100m",
			wantEchoValue: 85,
			wantFields:    []string{"threshold_pace"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeSportSettingsCreatorClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
				setting:           tc.created,
			}
			tool := newCreateSportSettingsTool(client, client, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.arguments)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(client.calls) != 1 {
				t.Fatalf("create calls = %#v, want one", client.calls)
			}
			call := client.calls[0]
			if call.Sport != tc.created.Type || !sameIntPtr(call.FTP, tc.wantFTP) || !sameIntPtr(call.IndoorFTP, tc.wantIndoorFTP) || !sameIntPtr(call.ThresholdHR, tc.wantHR) {
				t.Fatalf("create call = %+v, want sport and requested threshold values", call)
			}
			if tc.wantPace == 0 {
				if call.ThresholdPace != nil {
					t.Fatalf("threshold pace = %+v, want nil", call.ThresholdPace)
				}
			} else if call.ThresholdPace == nil || math.Abs(call.ThresholdPace.Value-tc.wantPace) > 0.000001 || call.ThresholdPace.PaceUnits != tc.wantUnits || call.ThresholdPace.PaceLoadType != tc.wantLoadType {
				t.Fatalf("threshold pace = %+v, want %.6f with %s/%s", call.ThresholdPace, tc.wantPace, tc.wantUnits, tc.wantLoadType)
			}

			out := resultMap(t, result)
			settings := out["sport_settings"].(map[string]any)
			if settings["sport"] != tc.created.Type || settings["sport_setting_id"] != float64(tc.created.ID) {
				t.Fatalf("sport settings = %#v, want created sport and ID", settings)
			}
			if tc.wantEchoKey != "" {
				if got, ok := settings[tc.wantEchoKey].(float64); !ok || math.Abs(got-tc.wantEchoValue) > 0.000001 || settings["pace_units_source"] != tc.created.PaceUnits || settings["pace_load_type"] != tc.wantLoadType {
					t.Fatalf("sport settings = %#v, want returned threshold pace rendered in selected units", settings)
				}
			}
			meta := out["_meta"].(map[string]any)
			if meta["operation"] != "create" || !sameAnyStringSlice(meta["fields_created"].([]any), tc.wantFields) {
				t.Fatalf("meta = %#v, want create operation and fields %v", meta, tc.wantFields)
			}
		})
	}
}

func TestCreateSportSettingsRejectsMalformedArgumentsBeforeProfileLookup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		arguments string
	}{
		{name: "blank sport", arguments: `{"sport":" ","ftp":250}`},
		{name: "unsupported sport", arguments: `{"sport":"Triathlon","ftp":250}`},
		{name: "no threshold field", arguments: `{"sport":"Ride"}`},
		{name: "non-positive FTP", arguments: `{"sport":"Ride","ftp":0}`},
		{name: "non-positive indoor FTP", arguments: `{"sport":"Ride","indoor_ftp":-1}`},
		{name: "non-positive threshold heart rate", arguments: `{"sport":"Ride","threshold_hr":0}`},
		{name: "malformed pace", arguments: `{"sport":"Run","threshold_pace":{"value":300}}`},
		{name: "unsupported pace", arguments: `{"sport":"Run","threshold_pace":{"value":300,"unit":"seconds_per_furlong"}}`},
		{name: "confirm is forbidden", arguments: `{"sport":"Ride","ftp":250,"confirm":true}`},
		{name: "recalculation is forbidden", arguments: `{"sport":"Ride","ftp":250,"recalc_hr_zones":true}`},
		{name: "zones are forbidden", arguments: `{"sport":"Ride","ftp":250,"zones":[]}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeSportSettingsCreatorClient{}
			tool := newCreateSportSettingsTool(client, client, "test", "UTC", false)

			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.arguments)})
			if err == nil || !strings.Contains(err.Error(), invalidCreateSportSettingsArgumentsMessage) {
				t.Fatalf("Handler() error = %v, want actionable create-arguments rejection", err)
			}
			if client.fakeProfileClient.calls != 0 {
				t.Fatalf("profile calls = %d, want 0 before malformed-argument rejection", client.fakeProfileClient.calls)
			}
			if len(client.calls) != 0 {
				t.Fatalf("create calls = %#v, want none before malformed-argument rejection", client.calls)
			}
		})
	}
}

func TestCreateSportSettingsRejectsExistingSettingsWithoutCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings []intervals.SportSettings
	}{
		{name: "upstream type", settings: []intervals.SportSettings{{ID: 7, Type: "Ride"}}},
		{name: "upstream types", settings: []intervals.SportSettings{{ID: 8, Types: []string{"Ride"}}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeSportSettingsCreatorClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{SportSettings: tc.settings}}}
			tool := newCreateSportSettingsTool(client, client, "test", "UTC", false)

			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Ride","ftp":250}`)})
			if err == nil || !strings.Contains(err.Error(), sportSettingsAlreadyExistMessage) || !strings.Contains(err.Error(), "use update_sport_settings") {
				t.Fatalf("Handler() error = %v, want duplicate guidance", err)
			}
			if client.fakeProfileClient.calls != 1 {
				t.Fatalf("profile calls = %d, want one duplicate lookup", client.fakeProfileClient.calls)
			}
			if len(client.calls) != 0 {
				t.Fatalf("create calls = %#v, want no write after duplicate lookup", client.calls)
			}
		})
	}
}

func assertCreateSportSettingsSchemaSafe(t *testing.T, schema map[string]any, allowAthleteID bool) {
	t.Helper()

	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("schema = %#v, want closed input object", schema)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties = %#v, want object", schema["properties"])
	}
	allowed := map[string]struct{}{
		"sport": {}, "ftp": {}, "indoor_ftp": {}, "threshold_hr": {}, "threshold_pace": {},
	}
	if allowAthleteID {
		allowed["athlete_id"] = struct{}{}
	}
	if len(properties) != len(allowed) {
		t.Fatalf("schema properties = %#v, want only %#v", properties, allowed)
	}
	for name := range allowed {
		if _, ok := properties[name]; !ok {
			t.Fatalf("schema properties = %#v, missing allowed %q", properties, name)
		}
	}
	for _, forbidden := range []string{
		"api_key", "apikey", "apiKey", "token", "credential", "credential_ref", "credentials",
		"confirm", "recalc_hr_zones", "recalcHrZones", "zones", "power_zones", "power_zone_names",
		"hr_zones", "hr_zone_names", "pace_zones", "pace_zone_names", "apply", "apply_to_activities",
	} {
		if _, ok := properties[forbidden]; ok {
			t.Fatalf("create schema exposes forbidden %q: %#v", forbidden, properties)
		}
	}
}

func sameAnyStringSlice(got []any, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index, value := range want {
		if got[index] != value {
			return false
		}
	}
	return true
}

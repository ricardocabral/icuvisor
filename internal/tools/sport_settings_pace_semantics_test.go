package tools

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestSportSettingsPaceRoundTripsMPSAndSelectedDisplay(t *testing.T) {
	tests := []struct {
		name            string
		sport           string
		paceUnits       string
		paceLoadType    string
		inputSeconds    float64
		inputUnit       string
		wantMPS         float64
		returnedMPS     float64
		responseField   string
		wantDisplaySecs float64
	}{
		{name: "run metric", sport: "Run", paceUnits: "MINS_KM", paceLoadType: "RUN", inputSeconds: 280, inputUnit: "seconds_per_km", wantMPS: 3.5714285, returnedMPS: 3.5714285, responseField: "threshold_pace_seconds_per_km", wantDisplaySecs: 280},
		{name: "run imperial", sport: "Run", paceUnits: "MINS_MILE", paceLoadType: "RUN", inputSeconds: 450.616329012327, inputUnit: "seconds_per_mile", wantMPS: 3.5714285, returnedMPS: 3.5714285, responseField: "threshold_pace_seconds_per_mile", wantDisplaySecs: 450.616329012327},
		{name: "swim 100 meters", sport: "Swim", paceUnits: "SECS_100M", paceLoadType: "SWIM", inputSeconds: 50, inputUnit: "seconds_per_100m", wantMPS: 2, returnedMPS: 2, responseField: "threshold_pace_seconds_per_100m", wantDisplaySecs: 50},
		{name: "swim 100 yards", sport: "Swim", paceUnits: "SECS_100Y", paceLoadType: "SWIM", inputSeconds: 45.72, inputUnit: "seconds_per_100y", wantMPS: 2, returnedMPS: 2, responseField: "threshold_pace_seconds_per_100y", wantDisplaySecs: 45.72},
		{name: "row 500 meters", sport: "Rowing", paceUnits: "SECS_500M", inputSeconds: 125, inputUnit: "seconds_per_500m", wantMPS: 4, returnedMPS: 4, responseField: "threshold_pace_seconds_per_500m", wantDisplaySecs: 125},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setting := intervals.SportSettings{ID: 8, Type: tc.sport, Types: []string{tc.sport}, PaceUnits: tc.paceUnits, PaceLoadType: tc.paceLoadType}
			client := newFakeSportSettingsClient(setting)
			client.setting = intervals.SportSettings{ID: 8, Type: tc.sport, ThresholdPace: tc.returnedMPS, PaceUnits: tc.paceUnits, PaceLoadType: tc.paceLoadType}
			tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

			args, err := json.Marshal(map[string]any{
				"sport":          tc.sport,
				"threshold_pace": map[string]any{"value": tc.inputSeconds, "unit": tc.inputUnit},
			})
			if err != nil {
				t.Fatalf("marshal arguments: %v", err)
			}
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: args})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			call := client.calls[0].ThresholdPace
			if call == nil || math.Abs(call.Value-tc.wantMPS) > 0.000001 || call.PaceUnits != tc.paceUnits || call.PaceLoadType != tc.paceLoadType {
				t.Fatalf("write pace = %+v, want %.7f m/s with %s/%s", call, tc.wantMPS, tc.paceUnits, tc.paceLoadType)
			}
			settings := resultMap(t, result)["sport_settings"].(map[string]any)
			value, ok := settings[tc.responseField].(float64)
			if !ok || math.Abs(value-tc.wantDisplaySecs) > 0.0001 || settings["pace_units_source"] != tc.paceUnits {
				t.Fatalf("settings = %#v, want returned m/s shaped in %s", settings, tc.responseField)
			}
			assertKeyAbsent(t, settings, "threshold_pace_meters_per_second")
		})
	}
}

func TestSportSettingsPaceZonePercentagesRemainUnchanged(t *testing.T) {
	boundaries := []float64{77.5, 100}
	if err := validateSportSettingsZones([]updateSportSettingsZoneRequest{{Kind: "pace", Boundaries: boundaries, Names: []string{"Easy", "Threshold"}}}); err != nil {
		t.Fatalf("validateSportSettingsZones() error = %v", err)
	}
	definitions := sportSettingsZoneDefinitions([]updateSportSettingsZoneRequest{{Kind: "pace", Boundaries: boundaries, Names: []string{"Easy", "Threshold"}}})
	if len(definitions) != 1 || definitions[0].Kind != "pace" || len(definitions[0].Boundaries) != 2 || definitions[0].Boundaries[0] != 77.5 || definitions[0].Boundaries[1] != 100 || definitions[0].Names[1] != "Threshold" {
		t.Fatalf("pace zone definitions = %#v, want unchanged percentages", definitions)
	}
}

package tools

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

const testMetersPer100Yards = 91.44

type swimPaceRoundTripCase struct {
	name               string
	inputUnit          string
	paceUnits          string
	responseField      string
	inputSeconds       float64
	wantMPS            float64
	returnedMPS        float64
	wantDisplaySeconds float64
}

func TestUpdateSportSettingsSwimPaceDecimalRoundTrips(t *testing.T) {
	for _, tc := range swimPaceDecimalRoundTripCases() {
		t.Run(tc.name, func(t *testing.T) {
			setting := intervals.SportSettings{ID: 9, Type: "Swim", Types: []string{"Swim"}, PaceUnits: tc.paceUnits, PaceLoadType: "SWIM"}
			client := newFakeSportSettingsClient(setting)
			client.setting = intervals.SportSettings{ID: 9, Type: "Swim", ThresholdPace: tc.returnedMPS, PaceUnits: tc.paceUnits, PaceLoadType: "SWIM"}
			tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: swimPaceRoundTripArguments(t, tc)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(client.calls) != 1 {
				t.Fatalf("update calls = %#v, want one", client.calls)
			}
			assertSwimPaceRoundTripTransport(t, client.calls[0].ThresholdPace, tc)
			assertSwimPaceRoundTripResponse(t, result, tc)
		})
	}
}

func TestCreateSportSettingsSwimPaceDecimalRoundTrips(t *testing.T) {
	for _, tc := range swimPaceDecimalRoundTripCases() {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeSportSettingsCreatorClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
				setting:           intervals.SportSettings{ID: 10, Type: "Swim", ThresholdPace: tc.returnedMPS, PaceUnits: tc.paceUnits, PaceLoadType: "SWIM"},
			}
			tool := newCreateSportSettingsTool(client, client, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: swimPaceRoundTripArguments(t, tc)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(client.calls) != 1 {
				t.Fatalf("create calls = %#v, want one", client.calls)
			}
			assertSwimPaceRoundTripTransport(t, client.calls[0].ThresholdPace, tc)
			assertSwimPaceRoundTripResponse(t, result, tc)
		})
	}
}

func swimPaceDecimalRoundTripCases() []swimPaceRoundTripCase {
	return []swimPaceRoundTripCase{
		{name: "metric 1:30 returns 2:00", inputUnit: "seconds_per_100m", paceUnits: "SECS_100M", responseField: "threshold_pace_seconds_per_100m", inputSeconds: 90, wantMPS: 100.0 / 90, returnedMPS: 100.0 / 120, wantDisplaySeconds: 120},
		{name: "metric 2:00 returns 1:30", inputUnit: "seconds_per_100m", paceUnits: "SECS_100M", responseField: "threshold_pace_seconds_per_100m", inputSeconds: 120, wantMPS: 100.0 / 120, returnedMPS: 100.0 / 90, wantDisplaySeconds: 90},
		{name: "yard 1:30 returns 2:00", inputUnit: "seconds_per_100y", paceUnits: "SECS_100Y", responseField: "threshold_pace_seconds_per_100y", inputSeconds: 90, wantMPS: testMetersPer100Yards / 90, returnedMPS: testMetersPer100Yards / 120, wantDisplaySeconds: 120},
		{name: "yard 2:00 returns 1:30", inputUnit: "seconds_per_100y", paceUnits: "SECS_100Y", responseField: "threshold_pace_seconds_per_100y", inputSeconds: 120, wantMPS: testMetersPer100Yards / 120, returnedMPS: testMetersPer100Yards / 90, wantDisplaySeconds: 90},
	}
}

func swimPaceRoundTripArguments(t *testing.T, tc swimPaceRoundTripCase) json.RawMessage {
	t.Helper()

	arguments, err := json.Marshal(map[string]any{
		"sport":          "Swim",
		"threshold_pace": map[string]any{"value": tc.inputSeconds, "unit": tc.inputUnit},
	})
	if err != nil {
		t.Fatalf("marshal arguments: %v", err)
	}
	return arguments
}

func assertSwimPaceRoundTripTransport(t *testing.T, pace *intervals.SportSettingsPace, tc swimPaceRoundTripCase) {
	t.Helper()

	if pace == nil || math.Abs(pace.Value-tc.wantMPS) > 0.0000001 || pace.PaceUnits != tc.paceUnits || pace.PaceLoadType != "SWIM" {
		t.Fatalf("threshold pace = %+v, want %.9f m/s with %s/SWIM metadata", pace, tc.wantMPS, tc.paceUnits)
	}
}

func assertSwimPaceRoundTripResponse(t *testing.T, result Result, tc swimPaceRoundTripCase) {
	t.Helper()

	payload := resultMap(t, result)
	settings := payload["sport_settings"].(map[string]any)
	gotDisplay, ok := settings[tc.responseField].(float64)
	if !ok || math.Abs(gotDisplay-tc.wantDisplaySeconds) > 0.0000001 || settings["pace_units_source"] != tc.paceUnits || settings["pace_load_type"] != "SWIM" {
		t.Fatalf("sport settings = %#v, want returned m/s shaped as %s=%.0f with %s/SWIM metadata", settings, tc.responseField, tc.wantDisplaySeconds, tc.paceUnits)
	}
	assertKeyAbsent(t, settings, "threshold_pace_meters_per_second")

	meta := payload["_meta"].(map[string]any)
	if meta["pace_input_unit"] != tc.inputUnit || meta["pace_display_unit"] != tc.paceUnits || meta["pace_load_type"] != "SWIM" {
		t.Fatalf("meta = %#v, want input/display/load pace metadata", meta)
	}
}

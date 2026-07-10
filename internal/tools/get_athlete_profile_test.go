package tools

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/athleteprofile"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type collectingRegistrar struct {
	tools []Tool
}

func (r *collectingRegistrar) AddTool(tool Tool) error {
	r.tools = append(r.tools, tool)
	return nil
}

type failingToolRegistrar struct {
	name string
	err  error
}

func (r failingToolRegistrar) AddTool(tool Tool) error {
	if tool.Name == r.name {
		return r.err
	}
	return nil
}

type fakeProfileClient struct {
	profile intervals.AthleteWithSportSettings
	err     error
	calls   int
	ctx     context.Context
}

type profileContextKey struct{}

func (f *fakeProfileClient) GetAthleteProfile(ctx context.Context) (intervals.AthleteWithSportSettings, error) {
	f.calls++
	f.ctx = ctx
	return f.profile, f.err
}

func TestGetAthleteProfileRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{}
	tool := newGetAthleteProfileTool(client, "v0.1-test", "America/Sao_Paulo", false)
	firstSentence, _, _ := strings.Cut(tool.Description, ".")
	for _, want := range []string{"athlete profile", "FTP", "thresholds", "zones", "sport settings"} {
		if !strings.Contains(firstSentence, want) {
			t.Fatalf("first sentence %q missing %q", firstSentence, want)
		}
	}

	schema, ok := tool.InputSchema.(map[string]any)
	if !ok {
		t.Fatalf("InputSchema type = %T, want map[string]any", tool.InputSchema)
	}
	if schema["type"] != "object" {
		t.Fatalf("schema type = %v, want object", schema["type"])
	}
	if schema["additionalProperties"] != false {
		t.Fatalf("additionalProperties = %v, want false", schema["additionalProperties"])
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties = %T, want map[string]any", schema["properties"])
	}
	if len(properties) != 1 {
		t.Fatalf("schema property count = %d, want 1", len(properties))
	}
	includeFull, ok := properties["include_full"].(map[string]any)
	if !ok {
		t.Fatalf("include_full schema = %T, want map[string]any", properties["include_full"])
	}
	if includeFull["type"] != "boolean" || includeFull["default"] != false {
		t.Fatalf("include_full schema = %#v, want boolean default false", includeFull)
	}
	if includeFull["description"] == "" {
		t.Fatal("include_full description is empty")
	}
	for name := range properties {
		lower := strings.ToLower(name)
		for _, forbidden := range []string{"api_key", "password", "token", "credential", "athlete_id"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("schema property %q contains forbidden %q", name, forbidden)
			}
		}
	}
}

func TestRegistryErrorNamesFailingTool(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	err := NewRegistry(newNoNetworkIntervalsClient(t), "test", "UTC").Register(context.Background(), failingToolRegistrar{name: getFitnessName, err: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Register() error = %v, want wrapped %v", err, wantErr)
	}
	if got := err.Error(); !strings.Contains(got, "registering get_fitness") || strings.Contains(got, getAthleteProfileName) {
		t.Fatalf("Register() error = %q, want failing tool name only", got)
	}
}

func TestGetAthleteProfileHandlerSuccess(t *testing.T) {
	t.Parallel()

	tool, client := newTestProfileTool(t, "v1.2.3", "UTC", intervals.AthleteWithSportSettings{
		ID:                    "i12345",
		Name:                  "Example Athlete",
		FirstName:             "Example",
		LastName:              "Athlete",
		MeasurementPreference: "METRIC",
		WeightPrefLB:          true,
		Fahrenheit:            true,
		Timezone:              "America/Sao_Paulo",
		Locale:                "pt_BR",
		SportSettings: []intervals.SportSettings{{
			ID:             7,
			AthleteID:      "i12345",
			Types:          []string{"Ride"},
			FTP:            250,
			IndoorFTP:      240,
			WPrime:         20000,
			PMax:           900,
			PowerZones:     []int{100, 150, 200},
			PowerZoneNames: []string{"Z1", "Z2", "Z3"},
			LTHR:           170,
			MaxHR:          190,
			HRZones:        []int{120, 140, 160},
			HRZoneNames:    []string{"Z1", "Z2", "Z3"},
			ThresholdPace:  3.5714285,
			PaceUnits:      "MINS_KM",
			PaceZones:      []float64{77.5, 90, 100},
			PaceZoneNames:  []string{"Z1", "Z2", "Z3"},
		}},
	})

	ctx := context.WithValue(context.Background(), profileContextKey{}, "sentinel")
	result, err := tool.Handler(ctx, Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("client calls = %d, want 1", client.calls)
	}
	if client.ctx == nil {
		t.Fatal("client context was not captured")
	}
	if got := client.ctx.Value(profileContextKey{}); got != "sentinel" {
		t.Fatalf("client context value = %v, want sentinel", got)
	}
	response := decodeProfileResult(t, result)
	if response.AthleteID != "i12345" || response.Meta.ServerVersion != "v1.2.3" {
		t.Fatalf("identity/meta = %+v, want normalized athlete and version", response)
	}
	if response.Timezone != "America/Sao_Paulo" || response.Locale != "pt_BR" {
		t.Fatalf("timezone/locale = %q/%q", response.Timezone, response.Locale)
	}
	if response.Units.MeasurementPreference != "metric" || response.Units.Weight != "lb" || response.Units.Temperature != "fahrenheit" {
		t.Fatalf("units = %+v, want metric/lb/fahrenheit", response.Units)
	}
	if len(response.SportSettings) != 1 {
		t.Fatalf("sport setting count = %d, want 1", len(response.SportSettings))
	}
	sport := response.SportSettings[0]
	if sport.FTPWatts != 250 || sport.IndoorFTPWatts != 240 || sport.LTHRBPM != 170 || sport.MaxHRBPM != 190 {
		t.Fatalf("sport thresholds = %+v", sport)
	}
	if sport.ThresholdPaceSecondsPerKM == nil || math.Abs(*sport.ThresholdPaceSecondsPerKM-280) > 0.0001 || len(sport.PaceZonesPercentOfThreshold) != 3 || sport.PaceZonesPercentOfThreshold[0] != 77.5 || sport.PaceZonesPercentOfThreshold[2] != 100 {
		t.Fatalf("km pace fields = %+v", sport)
	}
	if sport.ThresholdPaceSecondsPerMile != nil {
		t.Fatalf("mile pace field should be omitted for MINS_KM: %+v", sport)
	}
	if strings.Contains(resultText(t, result), "pace_zones_seconds_per_") {
		t.Fatalf("profile result retained misleading pace-zone duration fields: %s", resultText(t, result))
	}
	if sport.PaceUnitsSource != "MINS_KM" || sport.PaceDistanceUnit != "km" {
		t.Fatalf("pace metadata = %q/%q", sport.PaceUnitsSource, sport.PaceDistanceUnit)
	}
	if len(response.Meta.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want none for complete sport settings", response.Meta.Warnings)
	}
}

func TestGetAthleteProfileKeepsFTPAndZoneBoundariesSeparate(t *testing.T) {
	t.Parallel()

	response := newGetAthleteProfileResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{{
			Types:          []string{"Ride"},
			FTP:            250,
			IndoorFTP:      235,
			PowerZones:     []int{125, 188, 250, 300},
			PowerZoneNames: []string{"Z1", "Z2", "Boundary matching FTP", "Z4"},
		}},
	}, "test", "UTC")

	if len(response.SportSettings) != 1 {
		t.Fatalf("sport settings count = %d, want 1", len(response.SportSettings))
	}
	sport := response.SportSettings[0]
	if sport.FTPWatts != 250 || sport.IndoorFTPWatts != 235 {
		t.Fatalf("FTP fields = ftp:%d indoor:%d, want separate threshold values", sport.FTPWatts, sport.IndoorFTPWatts)
	}
	if len(sport.PowerZonesWatts) != 4 || sport.PowerZonesWatts[2] != 250 {
		t.Fatalf("power_zones_watts = %#v, want boundary array retaining 250", sport.PowerZonesWatts)
	}
	if sport.PowerZoneNames[2] != "Boundary matching FTP" {
		t.Fatalf("power_zone_names = %#v, want zone-boundary name preserved separately", sport.PowerZoneNames)
	}
	if !strings.Contains(response.Meta.PowerThresholdConvention, "ftp_watts is the upstream sport FTP threshold") || !strings.Contains(response.Meta.ZoneBoundaryConvention, "power_zones_watts and hr_zones_bpm are upstream zone boundary arrays") {
		t.Fatalf("profile semantic metadata = %#v / %#v", response.Meta.PowerThresholdConvention, response.Meta.ZoneBoundaryConvention)
	}
}

func TestGetAthleteProfileDoesNotTreatZoneBoundaryAsIndoorFTP(t *testing.T) {
	t.Parallel()

	response := newGetAthleteProfileResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{{
			Types:          []string{"Ride"},
			FTP:            260,
			PowerZones:     []int{130, 180, 240, 300},
			PowerZoneNames: []string{"Z1", "Z2", "Looks like indoor FTP", "Z4"},
		}},
	}, "test", "UTC")

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("decode response map: %v", err)
	}
	sport := raw["sport_settings"].([]any)[0].(map[string]any)
	if sport["ftp_watts"] != float64(260) {
		t.Fatalf("ftp_watts = %#v, want upstream FTP threshold", sport["ftp_watts"])
	}
	if _, ok := sport["indoor_ftp_watts"]; ok {
		t.Fatalf("sport row = %#v, did not expect absent upstream indoor_ftp to be synthesized", sport)
	}
	zones := sport["power_zones_watts"].([]any)
	if zones[2] != float64(240) || sport["power_zone_names"].([]any)[2] != "Looks like indoor FTP" {
		t.Fatalf("zone fields = %#v / %#v, want numeric boundary kept as zone data", zones, sport["power_zone_names"])
	}
	if !strings.Contains(response.Meta.PowerThresholdConvention, "Absence of indoor_ftp_watts means Icuvisor has no separate indoor FTP") || !strings.Contains(response.Meta.PowerThresholdConvention, "not that it should be inferred from zones") {
		t.Fatalf("power threshold convention = %q, want absent-indoor-FTP limitation", response.Meta.PowerThresholdConvention)
	}
}

func TestGetAthleteProfileReadinessWarnings(t *testing.T) {
	t.Parallel()

	response := newGetAthleteProfileResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{
			{Types: []string{"Ride"}},
			{Types: []string{"Run"}},
			{Types: []string{"Swim"}},
		},
	}, "test", "UTC")

	wantCodes := []string{
		"missing_power_threshold",
		"missing_power_zones",
		"missing_hr_threshold",
		"missing_hr_zones",
		"missing_hr_threshold",
		"missing_hr_zones",
		"missing_pace_threshold",
		"missing_pace_zones",
		"missing_hr_threshold",
		"missing_hr_zones",
		"missing_pace_threshold",
		"missing_pace_zones",
	}
	if got := profileWarningCodes(response.Meta.Warnings); !stringSlicesEqual(got, wantCodes) {
		t.Fatalf("warning codes = %#v, want %#v", got, wantCodes)
	}
	wantActionFields := map[string][]string{
		"missing_power_threshold": {"update_sport_settings", "ftp"},
		"missing_power_zones":     {"update_sport_settings", "zones", "kind=power"},
		"missing_hr_threshold":    {"update_sport_settings", "threshold_hr"},
		"missing_hr_zones":        {"update_sport_settings", "zones", "kind=hr"},
		"missing_pace_threshold":  {"update_sport_settings", "threshold_pace"},
		"missing_pace_zones":      {"update_sport_settings", "zones", "kind=pace"},
	}
	for _, warning := range response.Meta.Warnings {
		if len(warning.SportTypes) != 1 || warning.SportTypes[0] == "" {
			t.Fatalf("warning sport types = %#v", warning)
		}
		if warning.Field == "" || warning.Message == "" || warning.Action == "" {
			t.Fatalf("warning missing actionable context: %#v", warning)
		}
		for _, want := range wantActionFields[warning.Code] {
			if !strings.Contains(warning.Action, want) {
				t.Fatalf("warning action = %q, want %q for code %s", warning.Action, want, warning.Code)
			}
		}
		lower := strings.ToLower(warning.Message + " " + warning.Action)
		for _, forbidden := range []string{"i12345", "api", "credential", "token", "password", "lthr", "fthr"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("warning leaks forbidden text %q: %#v", forbidden, warning)
			}
		}
	}
}

func TestGetAthleteProfileHandlerSerializesReadinessWarnings(t *testing.T) {
	t.Parallel()

	tool, _ := newTestProfileTool(t, "test", "UTC", intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{
			{Types: []string{"Ride"}},
			{Types: []string{"Run"}},
			{Types: []string{"Swim"}},
		},
	})
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	response := decodeProfileResult(t, result)
	wantCodes := []string{
		"missing_power_threshold",
		"missing_power_zones",
		"missing_hr_threshold",
		"missing_hr_zones",
		"missing_hr_threshold",
		"missing_hr_zones",
		"missing_pace_threshold",
		"missing_pace_zones",
		"missing_hr_threshold",
		"missing_hr_zones",
		"missing_pace_threshold",
		"missing_pace_zones",
	}
	if got := profileWarningCodes(response.Meta.Warnings); !stringSlicesEqual(got, wantCodes) {
		t.Fatalf("warning codes = %#v, want %#v", got, wantCodes)
	}
	text := resultText(t, result)
	for _, want := range []string{"\"warnings\"", "\"missing_power_threshold\"", "threshold_hr", "kind=pace", "\"power_threshold_convention\"", "\"zone_boundary_convention\""} {
		if !strings.Contains(text, want) {
			t.Fatalf("serialized response missing %s: %s", want, text)
		}
	}
}

func TestGetAthleteProfileHandlerOmitsReadinessWarningsWhenAliasesComplete(t *testing.T) {
	t.Parallel()

	tool, _ := newTestProfileTool(t, "test", "UTC", intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{
			{Types: []string{"Ride"}, FTP: 250, FTHR: 170, PowerZones: []int{100, 150}, HRZones: []int{120, 140}},
			{Types: []string{"Run"}, FTHR: 170, HRZones: []int{120, 140}, PaceThreshold: 300, PaceUnits: "MINS_KM", PaceZones: []float64{360, 330}},
			{Types: []string{"Swim"}, FTHR: 150, HRZones: []int{120, 140}, PaceThreshold: 90, PaceUnits: "SECS_100M", PaceZones: []float64{100, 90}},
		},
	})
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	response := decodeProfileResult(t, result)
	if len(response.Meta.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want none", response.Meta.Warnings)
	}
	if strings.Contains(resultText(t, result), "\"warnings\"") {
		t.Fatalf("serialized complete profile includes warnings: %s", resultText(t, result))
	}
}

func TestGetAthleteProfileReadinessWarningsPreferTypesOverLegacyType(t *testing.T) {
	t.Parallel()

	response := newGetAthleteProfileResponse(intervals.AthleteWithSportSettings{
		ID:            "i12345",
		SportSettings: []intervals.SportSettings{{Type: "Ride", Types: []string{"Run"}}},
	}, "test", "UTC")
	if got := profileWarningCodes(response.Meta.Warnings); !stringSlicesEqual(got, []string{"missing_hr_threshold", "missing_hr_zones", "missing_pace_threshold", "missing_pace_zones"}) {
		t.Fatalf("warning codes = %#v", got)
	}
	for _, warning := range response.Meta.Warnings {
		if !stringSlicesEqual(warning.SportTypes, []string{"Run"}) {
			t.Fatalf("sport types = %#v, want Run only", warning.SportTypes)
		}
	}

	fallback := newGetAthleteProfileResponse(intervals.AthleteWithSportSettings{
		ID:            "i12345",
		SportSettings: []intervals.SportSettings{{Type: "Ride"}},
	}, "test", "UTC")
	if got := profileWarningCodes(fallback.Meta.Warnings); !stringSlicesEqual(got[:2], []string{"missing_power_threshold", "missing_power_zones"}) {
		t.Fatalf("fallback warning codes = %#v", got)
	}
	if !stringSlicesEqual(fallback.Meta.Warnings[0].SportTypes, []string{"Ride"}) {
		t.Fatalf("fallback sport types = %#v, want Ride", fallback.Meta.Warnings[0].SportTypes)
	}
}

func TestGetAthleteProfileReadinessWarningsSkipNonApplicableSports(t *testing.T) {
	t.Parallel()

	response := newGetAthleteProfileResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{
			{Types: []string{"Strength"}},
			{Types: []string{"Yoga"}},
			{Types: []string{"Other"}},
		},
	}, "test", "UTC")
	if len(response.Meta.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want none for non-applicable sports", response.Meta.Warnings)
	}
}

func TestGetAthleteProfileReadinessWarningsOmittedWhenComplete(t *testing.T) {
	t.Parallel()

	response := newGetAthleteProfileResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{
			{Types: []string{"Ride"}, FTP: 250, LTHR: 170, PowerZones: []int{100, 150}, HRZones: []int{120, 140}},
			{Types: []string{"Run"}, LTHR: 170, HRZones: []int{120, 140}, ThresholdPace: 300, PaceUnits: "MINS_KM", PaceZones: []float64{360, 330}},
			{Types: []string{"Swim"}, LTHR: 150, HRZones: []int{120, 140}, ThresholdPace: 90, PaceUnits: "SECS_100M", PaceZones: []float64{100, 90}},
		},
	}, "test", "UTC")
	if len(response.Meta.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want none", response.Meta.Warnings)
	}
}

func TestGetAthleteProfileIncludeFullDelta(t *testing.T) {
	t.Parallel()

	profile := intervals.AthleteWithSportSettings{
		ID:                    "i12345",
		MeasurementPreference: "IMPERIAL",
		SportSettings: []intervals.SportSettings{{
			ID:        9,
			AthleteID: "i12345",
			Types:     []string{"Run"},
		}},
	}
	tool, _ := newTestProfileTool(t, "test", "UTC", profile)

	defaultResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("default Handler() error = %v", err)
	}
	defaultText := resultText(t, defaultResult)
	for _, forbidden := range []string{"measurement_preference_source", "sport_setting_id", "sport_setting_athlete_id"} {
		if strings.Contains(defaultText, forbidden) {
			t.Fatalf("default response contains full-only field %q: %s", forbidden, defaultText)
		}
	}

	explicitFalseResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"include_full":false}`)})
	if err != nil {
		t.Fatalf("explicit false Handler() error = %v", err)
	}
	if explicitFalseText := resultText(t, explicitFalseResult); explicitFalseText != defaultText {
		t.Fatalf("include_full=false changed default response\ndefault: %s\nfalse:   %s", defaultText, explicitFalseText)
	}

	fullResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"include_full":true}`)})
	if err != nil {
		t.Fatalf("full Handler() error = %v", err)
	}
	fullText := resultText(t, fullResult)
	for _, want := range []string{"measurement_preference_source", "sport_setting_id", "sport_setting_athlete_id", "i12345"} {
		if !strings.Contains(fullText, want) {
			t.Fatalf("full response missing %q: %s", want, fullText)
		}
	}
}

func TestGetAthleteProfileResponseShapingVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		version      string
		fallback     string
		profile      intervals.AthleteWithSportSettings
		wantTimezone string
		wantUnits    GetAthleteProfileUnits
		wantMilePace bool
	}{
		{
			name:         "configured timezone fallback and mile pace",
			version:      "",
			fallback:     "Europe/Lisbon",
			wantTimezone: "Europe/Lisbon",
			wantUnits:    GetAthleteProfileUnits{MeasurementPreference: "imperial", Weight: "lb", Temperature: "celsius"},
			wantMilePace: true,
			profile: intervals.AthleteWithSportSettings{
				ID:           "i12345",
				WeightPrefLB: true,
				SportSettings: []intervals.SportSettings{{
					ThresholdPace: 3.5714285,
					PaceUnits:     "MINS_MILE",
					PaceZones:     []float64{77.5, 100},
				}},
			},
		},
		{
			name:         "default timezone fallback and metric units",
			fallback:     "",
			wantTimezone: config.DefaultTimezone,
			wantUnits:    GetAthleteProfileUnits{MeasurementPreference: "metric", Weight: "kg", Temperature: "celsius"},
			profile:      intervals.AthleteWithSportSettings{ID: "i12345", MeasurementPreference: "METRIC"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := newGetAthleteProfileResponse(tc.profile, tc.version, normalizeTimezoneFallback(tc.fallback))
			if response.Timezone != tc.wantTimezone {
				t.Fatalf("timezone = %q, want %q", response.Timezone, tc.wantTimezone)
			}
			if response.Units != tc.wantUnits {
				t.Fatalf("units = %+v, want %+v", response.Units, tc.wantUnits)
			}
			if response.Meta.ServerVersion == "" {
				t.Fatal("server version is empty")
			}
			if tc.wantMilePace {
				sport := response.SportSettings[0]
				if sport.ThresholdPaceSecondsPerMile == nil || math.Abs(*sport.ThresholdPaceSecondsPerMile-450.616329012327) > 0.0001 || len(sport.PaceZonesPercentOfThreshold) != 2 || sport.PaceDistanceUnit != "mile" || sport.PaceUnitsSource != "MINS_MILE" {
					t.Fatalf("mile pace shaping = %+v", sport)
				}
				if sport.ThresholdPaceSecondsPerKM != nil {
					t.Fatalf("km pace field should be omitted for mile pace: %+v", sport)
				}
			}
		})
	}
}

func TestGetAthleteProfileShapesYardSwimPace(t *testing.T) {
	t.Parallel()

	response := newGetAthleteProfileResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{{
			Types:         []string{"Swim"},
			ThresholdPace: 2,
			PaceUnits:     "SECS_100Y",
			PaceZones:     []float64{77.5, 100},
		}},
	}, "test", "UTC")
	if len(response.SportSettings) != 1 {
		t.Fatalf("sport settings = %d, want 1", len(response.SportSettings))
	}
	sport := response.SportSettings[0]
	if sport.ThresholdPaceSecondsPer100Y == nil || *sport.ThresholdPaceSecondsPer100Y != 45.72 || len(sport.PaceZonesPercentOfThreshold) != 2 || sport.PaceZonesPercentOfThreshold[0] != 77.5 || sport.PaceZonesPercentOfThreshold[1] != 100 {
		t.Fatalf("yard swim pace shaping = %+v", sport)
	}
	if sport.PaceDistanceUnit != "100y" || sport.PaceUnitsSource != "SECS_100Y" {
		t.Fatalf("yard swim pace metadata = %+v", sport)
	}
	if sport.ThresholdPaceSecondsPer100M != nil || sport.ThresholdPaceMetersPerSecond != nil || sport.ThresholdPaceValue != nil {
		t.Fatalf("yard swim pace used wrong fields: %+v", sport)
	}
}

func TestGetAthleteProfilePaceConversionPolicies(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	profile := intervals.AthleteWithSportSettings{
		ID:             "i12345",
		PreferredUnits: "miles",
		SportSettings: []intervals.SportSettings{
			{Types: []string{"Run"}, ThresholdPace: 3.5714285, PaceUnits: "MINS_KM", PaceZones: []float64{77.5, 100}},
			{Types: []string{"Swim"}, ThresholdPace: 2, PaceUnits: "SECS_100M", PaceZones: []float64{80, 100}},
			{Types: []string{"Other"}, ThresholdPace: 7, PaceUnits: "FEET", PaceZones: []float64{77.5, 100}},
		},
	}
	response := newGetAthleteProfileResponse(profile, "test", "UTC")
	if len(response.SportSettings) != 3 {
		t.Fatalf("sport settings = %d, want 3", len(response.SportSettings))
	}
	run := response.SportSettings[0]
	if run.ThresholdPaceSecondsPerKM == nil || math.Abs(*run.ThresholdPaceSecondsPerKM-280) > 0.0001 || len(run.PaceZonesPercentOfThreshold) != 2 || run.PaceDistanceUnit != "km" {
		t.Fatalf("run pace conversion = %+v", run)
	}
	if run.ThresholdPaceSecondsPerMile != nil {
		t.Fatalf("run mile pace should be omitted when pace_units is MINS_KM: %+v", run)
	}
	swim := response.SportSettings[1]
	if swim.ThresholdPaceSecondsPer100M == nil || *swim.ThresholdPaceSecondsPer100M != 50 || len(swim.PaceZonesPercentOfThreshold) != 2 || swim.PaceDistanceUnit != "100m" {
		t.Fatalf("swim pace pass-through = %+v", swim)
	}
	unknown := response.SportSettings[2]
	if unknown.ThresholdPaceMetersPerSecond == nil || *unknown.ThresholdPaceMetersPerSecond != 7 || unknown.ThresholdPaceValue == nil || *unknown.ThresholdPaceValue != 7 || len(unknown.PaceZonesPercentOfThreshold) != 2 || unknown.Meta["unknown_unit"] != "FEET" || unknown.PaceDistanceUnit != "FEET" {
		t.Fatalf("unknown pace pass-through = %+v", unknown)
	}
}

func TestGetAthleteProfileArgumentValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args string
	}{
		{name: "unknown api key", args: `{"api_key":"secret"}`},
		{name: "unknown athlete id", args: `{"athlete_id":"i12345"}`},
		{name: "null", args: `null`},
		{name: "array", args: `[]`},
		{name: "boolean", args: `true`},
		{name: "trailing json", args: `{} {}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tool, client := newTestProfileTool(t, "test", "UTC", intervals.AthleteWithSportSettings{ID: "i12345"})
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err == nil {
				t.Fatal("Handler() error = nil, want invalid arguments error")
			}
			message, ok := PublicErrorMessage(err)
			if !ok || message != invalidGetAthleteProfileArgumentsMessage {
				t.Fatalf("public error = (%q, %v), want invalid args", message, ok)
			}
			if client.calls != 0 {
				t.Fatalf("client calls = %d, want 0", client.calls)
			}
		})
	}
}

func TestGetAthleteProfileErrorMapping(t *testing.T) {
	t.Parallel()

	upstreamErr := errors.New("upstream secret detail")
	tool, _ := newTestProfileToolWithError(t, upstreamErr)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want upstream error")
	}
	message, ok := PublicErrorMessage(err)
	if !ok || message != fetchAthleteProfileMessage {
		t.Fatalf("public error = (%q, %v), want fetch message", message, ok)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("public error leaked internal detail: %q", err.Error())
	}
}

func TestGetAthleteProfileCancellationIsNotMappedToCredentialError(t *testing.T) {
	t.Parallel()

	tool, _ := newTestProfileToolWithError(t, context.Canceled)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Handler() error = %v, want context.Canceled", err)
	}
	if _, ok := PublicErrorMessage(err); ok {
		t.Fatalf("cancellation should not be a public user error: %v", err)
	}
}

func TestGetAthleteProfileMetaUnitsFromPreferredUnits(t *testing.T) {
	t.Parallel()

	tool, _ := newTestProfileTool(t, "test", "UTC", intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "miles"})
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	text := resultText(t, result)
	for _, want := range []string{"\"measurement_preference\":\"imperial\"", "\"units\":{\"distance\":\"mi\"", "\"system\":\"imperial\""} {
		if !strings.Contains(text, want) {
			t.Fatalf("response missing unit metadata %s: %s", want, text)
		}
	}
	meta := result.StructuredContent.(map[string]any)["_meta"].(map[string]any)
	if meta["server_version"] != "test" {
		t.Fatalf("server_version = %v, want test", meta["server_version"])
	}
	units := meta["units"].(map[string]string)
	if units["system"] != "imperial" || units["distance"] != "mi" {
		t.Fatalf("meta units = %+v, want imperial/mi", units)
	}
}

func TestGetAthleteProfileDefaultsEmptyUnitsConsistentlyToMetric(t *testing.T) {
	t.Parallel()

	tool, _ := newTestProfileTool(t, "test", "UTC", intervals.AthleteWithSportSettings{ID: "i12345"})
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	text := resultText(t, result)
	for _, want := range []string{"\"measurement_preference\":\"metric\"", "\"units\":{\"distance\":\"km\"", "\"system\":\"metric\""} {
		if !strings.Contains(text, want) {
			t.Fatalf("response missing default metric unit marker %s: %s", want, text)
		}
	}
}

func TestGetAthleteProfileDebugMetadataOptIn(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}}
	tool := newGetAthleteProfileTool(client, "test", "UTC", true)
	result, err := tool.Handler(context.Background(), Request{Name: getAthleteProfileName, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	text := resultText(t, result)
	for _, want := range []string{"\"query_type\":\"get_athlete_profile\"", "\"fetched_at\":"} {
		if !strings.Contains(text, want) {
			t.Fatalf("debug response missing %s: %s", want, text)
		}
	}
}

func TestGetAthleteProfileOmitsForbiddenDebugAndSecretFields(t *testing.T) {
	t.Parallel()

	tool, _ := newTestProfileTool(t, "test", "UTC", intervals.AthleteWithSportSettings{
		ID:                    "i12345",
		Name:                  "Safe Athlete",
		MeasurementPreference: "IMPERIAL",
		SportSettings: []intervals.SportSettings{{
			ID:        7,
			AthleteID: "i12345",
		}},
	})
	for _, args := range []string{`{}`, `{"include_full":true}`} {
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(args)})
		if err != nil {
			t.Fatalf("Handler(%s) error = %v", args, err)
		}
		lower := strings.ToLower(resultText(t, result))
		if args != `{}` && !strings.Contains(lower, "sport_setting_id") {
			t.Fatalf("full response did not include full-only fields: %s", lower)
		}
		for _, forbidden := range []string{"fetched_at", "query_type", "raw_payload", "raw_upstream", "http://", "https://", "authorization", "header", "credential", "api_key", "token", "basic"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("response contains forbidden %q: %s", forbidden, lower)
			}
		}
	}
}

func newTestProfileTool(t *testing.T, version string, timezoneFallback string, profile intervals.AthleteWithSportSettings) (Tool, *fakeProfileClient) {
	t.Helper()
	client := &fakeProfileClient{profile: profile}
	return newGetAthleteProfileTool(client, version, timezoneFallback, false), client
}

func newTestProfileToolWithError(t *testing.T, err error) (Tool, *fakeProfileClient) {
	t.Helper()
	client := &fakeProfileClient{err: err}
	return newGetAthleteProfileTool(client, "test", "UTC", false), client
}

func decodeProfileResult(t *testing.T, result Result) GetAthleteProfileResponse {
	t.Helper()
	structuredJSON, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var structured GetAthleteProfileResponse
	if err := json.Unmarshal(structuredJSON, &structured); err != nil {
		t.Fatalf("decode structured response: %v", err)
	}
	var textResponse GetAthleteProfileResponse
	if err := json.Unmarshal([]byte(resultText(t, result)), &textResponse); err != nil {
		t.Fatalf("decode text response: %v", err)
	}
	if structured.AthleteID != textResponse.AthleteID || structured.Meta.ServerVersion != textResponse.Meta.ServerVersion {
		t.Fatalf("structured/text mismatch: %+v vs %+v", structured, textResponse)
	}
	return structured
}

func profileWarningCodes(warnings []athleteprofile.ReadinessWarning) []string {
	codes := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		codes = append(codes, warning.Code)
	}
	return codes
}

func stringSlicesEqual(got []string, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func resultText(t *testing.T, result Result) string {
	t.Helper()
	if result.IsError {
		t.Fatal("result IsError = true, want false")
	}
	if len(result.Content) != 1 {
		t.Fatalf("content count = %d, want 1", len(result.Content))
	}
	if result.Content[0].Type != ContentTypeText {
		t.Fatalf("content type = %q, want text", result.Content[0].Type)
	}
	return result.Content[0].Text
}

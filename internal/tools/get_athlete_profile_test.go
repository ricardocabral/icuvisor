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
		ID:                    "12345",
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
			AthleteID:      "12345",
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
			ThresholdPace:  255.5,
			PaceUnits:      "MINS_KM",
			PaceZones:      []float64{330, 300, 270},
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
	if sport.ThresholdPaceSecondsPerKM == nil || *sport.ThresholdPaceSecondsPerKM != 255.5 || len(sport.PaceZonesSecondsPerKM) != 3 {
		t.Fatalf("km pace fields = %+v", sport)
	}
	if sport.ThresholdPaceSecondsPerMile != nil || len(sport.PaceZonesSecondsPerMile) != 0 {
		t.Fatalf("mile pace fields should be omitted for MINS_KM: %+v", sport)
	}
	if sport.PaceUnitsSource != "MINS_KM" || sport.PaceDistanceUnit != "km" {
		t.Fatalf("pace metadata = %q/%q", sport.PaceUnitsSource, sport.PaceDistanceUnit)
	}
}

func TestGetAthleteProfileIncludeFullDelta(t *testing.T) {
	t.Parallel()

	profile := intervals.AthleteWithSportSettings{
		ID:                    "i12345",
		MeasurementPreference: "IMPERIAL",
		SportSettings: []intervals.SportSettings{{
			ID:        9,
			AthleteID: "12345",
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
				ID:           "12345",
				WeightPrefLB: true,
				SportSettings: []intervals.SportSettings{{
					ThresholdPace: 400,
					PaceUnits:     "MINS_MILE",
					PaceZones:     []float64{420, 390},
				}},
			},
		},
		{
			name:         "default timezone fallback and metric units",
			fallback:     "",
			wantTimezone: config.DefaultTimezone,
			wantUnits:    GetAthleteProfileUnits{MeasurementPreference: "metric", Weight: "kg", Temperature: "celsius"},
			profile:      intervals.AthleteWithSportSettings{ID: "12345", MeasurementPreference: "METRIC"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := newGetAthleteProfileResponse(tc.profile, tc.version, normalizeTimezoneFallback(tc.fallback), false)
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
				if sport.ThresholdPaceSecondsPerMile == nil || len(sport.PaceZonesSecondsPerMile) != 2 || sport.PaceDistanceUnit != "mile" || sport.PaceUnitsSource != "MINS_MILE" {
					t.Fatalf("mile pace shaping = %+v", sport)
				}
				if sport.ThresholdPaceSecondsPerKM != nil || len(sport.PaceZonesSecondsPerKM) != 0 {
					t.Fatalf("km pace fields should be omitted for mile pace: %+v", sport)
				}
			}
		})
	}
}

func TestGetAthleteProfilePaceConversionPolicies(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	profile := intervals.AthleteWithSportSettings{
		ID:             "12345",
		PreferredUnits: "miles",
		SportSettings: []intervals.SportSettings{
			{Types: []string{"Run"}, ThresholdPace: 5, PaceUnits: "MINS_KM", PaceZones: []float64{4, 5}},
			{Types: []string{"Swim"}, ThresholdPace: 82, PaceUnits: "SECS_100M", PaceZones: []float64{80, 85}},
			{Types: []string{"Other"}, ThresholdPace: 7, PaceUnits: "FEET", PaceZones: []float64{6, 7}},
		},
	}
	response := newGetAthleteProfileResponse(profile, "test", "UTC", false)
	if len(response.SportSettings) != 3 {
		t.Fatalf("sport settings = %d, want 3", len(response.SportSettings))
	}
	run := response.SportSettings[0]
	if run.ThresholdPaceSecondsPerMile == nil || math.Abs(*run.ThresholdPaceSecondsPerMile-8.04672) > 0.000001 || len(run.PaceZonesSecondsPerMile) != 2 || run.PaceDistanceUnit != "mile" {
		t.Fatalf("run pace conversion = %+v", run)
	}
	if run.ThresholdPaceSecondsPerKM != nil || len(run.PaceZonesSecondsPerKM) != 0 {
		t.Fatalf("run km fields should be omitted after imperial conversion: %+v", run)
	}
	swim := response.SportSettings[1]
	if swim.ThresholdPaceSecondsPer100M == nil || *swim.ThresholdPaceSecondsPer100M != 82 || len(swim.PaceZonesSecondsPer100M) != 2 || swim.PaceDistanceUnit != "100m" {
		t.Fatalf("swim pace pass-through = %+v", swim)
	}
	unknown := response.SportSettings[2]
	if unknown.ThresholdPaceValue == nil || *unknown.ThresholdPaceValue != 7 || len(unknown.PaceZonesValues) != 2 || unknown.Meta["unknown_unit"] != "FEET" || unknown.PaceDistanceUnit != "FEET" {
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
			tool, client := newTestProfileTool(t, "test", "UTC", intervals.AthleteWithSportSettings{ID: "12345"})
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

	tool, _ := newTestProfileTool(t, "test", "UTC", intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "miles"})
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

	tool, _ := newTestProfileTool(t, "test", "UTC", intervals.AthleteWithSportSettings{ID: "12345"})
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

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345"}}
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
		ID:                    "12345",
		Name:                  "Safe Athlete",
		MeasurementPreference: "IMPERIAL",
		SportSettings: []intervals.SportSettings{{
			ID:        7,
			AthleteID: "12345",
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

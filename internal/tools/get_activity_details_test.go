package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeActivityReadClient struct {
	fakeProfileClient
	activity     intervals.Activity
	activityErr  error
	intervals    intervals.IntervalsDTO
	intervalErr  error
	streams      []intervals.ActivityStream
	streamErr    error
	streamCalls  int
	streamParams intervals.ActivityStreamsParams
	messages     []intervals.ActivityMessage
	messageErr   error
	gear         []intervals.Gear
	gearErr      error
	gearCalls    int
}

func (f *fakeActivityReadClient) ListGear(ctx context.Context) ([]intervals.Gear, error) {
	f.gearCalls++
	if f.gearErr != nil {
		return nil, f.gearErr
	}
	return f.gear, nil
}

func (f *fakeActivityReadClient) GetActivity(ctx context.Context, activityID string) (intervals.Activity, error) {
	return f.activity, f.activityErr
}

func (f *fakeActivityReadClient) GetActivityIntervals(ctx context.Context, activityID string) (intervals.IntervalsDTO, error) {
	return f.intervals, f.intervalErr
}

func TestActivityReadToolsRegistration(t *testing.T) {
	t.Parallel()

	registrar := &collectingRegistrar{}
	if err := NewRegistry(newNoNetworkIntervalsClient(t), "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	findTool(t, registrar.tools, getActivityDetailsName)
	findTool(t, registrar.tools, getActivityIntervalsName)
	findTool(t, registrar.tools, getActivityMessagesName)
}

func TestGetActivityDetailsCaloriesBurnedSemantics(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"a1","icu_athlete_id":"i12345","name":"Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00","calories":450}`)
	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity}
	tool := newGetActivityDetailsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	activityMap := payload["activity"].(map[string]any)
	if activityMap["calories_burned"] != float64(450) {
		t.Fatalf("activity = %#v, want calories_burned", activityMap)
	}
	if _, ok := activityMap["calories"]; ok {
		t.Fatalf("activity = %#v, want no ambiguous calories key", activityMap)
	}
	semantics := payload["_meta"].(map[string]any)["field_semantics"].(map[string]any)
	if !strings.Contains(semantics["calories_burned"].(string), "Active/exercise calories") {
		t.Fatalf("field_semantics = %#v, want active calories label", semantics)
	}
}

func TestGetActivityDetailsNutritionFieldsDisambiguated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		raw           string
		wantIngested  *int
		wantUsed      *int
		wantAbsent    []string
		wantSemantics []string
	}{
		{
			name:          "terse default with all three nutrition fields",
			raw:           `{"id":"a1","icu_athlete_id":"i12345","name":"Long Ride","type":"Ride","start_date_local":"2026-05-14T07:00:00","calories":1850,"carbs_ingested":210,"carbs_used":390}`,
			wantIngested:  intPtr(210),
			wantUsed:      intPtr(390),
			wantAbsent:    []string{"carbs", "carbohydrates", "carbs_ingested", "carbs_used"},
			wantSemantics: []string{"calories_burned", "carbs_ingested_g", "carbs_used_g"},
		},
		{
			name:         "null-stripping when nutrition absent",
			raw:          `{"id":"a2","icu_athlete_id":"i12345","name":"Easy Run","type":"Run","start_date_local":"2026-05-14T08:00:00","calories":400}`,
			wantIngested: nil,
			wantUsed:     nil,
			wantAbsent:   []string{"carbs_ingested_g", "carbs_used_g", "carbs", "carbohydrates"},
		},
		{
			name:          "only carbs_used present",
			raw:           `{"id":"a3","icu_athlete_id":"i12345","name":"Tempo","type":"Ride","start_date_local":"2026-05-14T09:00:00","calories":900,"carbs_used":200}`,
			wantIngested:  nil,
			wantUsed:      intPtr(200),
			wantAbsent:    []string{"carbs_ingested_g", "carbs", "carbohydrates"},
			wantSemantics: []string{"carbs_used_g"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			activity := decodeActivityFixture(t, tc.raw)
			client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity}
			tool := newGetActivityDetailsToolWithGear(client, client, nil, nil, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			activityMap := resultMap(t, result)["activity"].(map[string]any)

			if tc.wantIngested != nil {
				if activityMap["carbs_ingested_g"] != float64(*tc.wantIngested) {
					t.Fatalf("carbs_ingested_g = %v, want %d", activityMap["carbs_ingested_g"], *tc.wantIngested)
				}
			}
			if tc.wantUsed != nil {
				if activityMap["carbs_used_g"] != float64(*tc.wantUsed) {
					t.Fatalf("carbs_used_g = %v, want %d", activityMap["carbs_used_g"], *tc.wantUsed)
				}
			}
			for _, key := range tc.wantAbsent {
				if _, ok := activityMap[key]; ok {
					t.Fatalf("activity row emitted disallowed/ambiguous key %s: %#v", key, activityMap)
				}
			}
			if len(tc.wantSemantics) > 0 {
				semantics := resultMap(t, result)["_meta"].(map[string]any)["field_semantics"].(map[string]any)
				for _, key := range tc.wantSemantics {
					if label, ok := semantics[key].(string); !ok || label == "" {
						t.Fatalf("field_semantics missing %s: %#v", key, semantics)
					}
				}
			}
		})
	}
}

func TestGetActivityDetailsCarbsIngestedDistinctFromWellnessKcalConsumed(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"a1","icu_athlete_id":"i12345","name":"Ride","type":"Ride","start_date_local":"2026-05-14T07:00:00","calories":1850,"carbs_ingested":210,"carbs_used":390}`)
	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity}
	tool := newGetActivityDetailsToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	activityMap := resultMap(t, result)["activity"].(map[string]any)
	semantics := resultMap(t, result)["_meta"].(map[string]any)["field_semantics"].(map[string]any)

	if _, ok := activityMap["kcal_consumed"]; ok {
		t.Fatalf("activity row emitted kcal_consumed (wellness intake key) in activity: %#v", activityMap)
	}
	if _, ok := activityMap["calories_intake"]; ok {
		t.Fatalf("activity row emitted calories_intake (wellness intake key) in activity: %#v", activityMap)
	}
	caloriesLabel, ok := semantics["calories_burned"].(string)
	if !ok || !strings.Contains(caloriesLabel, "Distinct from wellness kcal_consumed") {
		t.Fatalf("calories_burned semantics = %q, want distinction from wellness kcal_consumed", caloriesLabel)
	}
}

func TestGetActivityDetailsNutritionIncludeFullPreservesRawUpstreamKeys(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"a1","icu_athlete_id":"i12345","name":"Ride","type":"Ride","start_date_local":"2026-05-14T07:00:00","calories":1850,"carbs_ingested":210,"carbs_used":390}`)
	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity}
	tool := newGetActivityDetailsToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	activityMap := resultMap(t, result)["activity"].(map[string]any)
	if activityMap["carbs_ingested_g"] != float64(210) || activityMap["carbs_used_g"] != float64(390) {
		t.Fatalf("include_full terse keys = %#v, want disambiguated nutrition keys", activityMap)
	}
	full, ok := activityMap["full"].(map[string]any)
	if !ok {
		t.Fatalf("include_full missing full payload: %#v", activityMap)
	}
	if full["carbs_ingested"] != float64(210) || full["carbs_used"] != float64(390) || full["calories"] != float64(1850) {
		t.Fatalf("full raw nutrition = %#v, want upstream key names preserved", full)
	}
}

func TestGetActivityDetailsShapesTerseFullAndStravaUnavailable(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"stub1","icu_athlete_id":"i12345","start_date_local":"2026-01-02T07:00:00","name":null}`)
	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "imperial", Timezone: "America/Sao_Paulo"}}, activity: activity}
	tool := newGetActivityDetailsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	activityMap := resultMap(t, result)["activity"].(map[string]any)
	if activityMap["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("timezone = %v, want profile timezone", activityMap["timezone"])
	}
	assertUnavailableReasonAndWorkaround(t, activityMap, "strava_tos", wantUnknownStravaWorkaround)
	full := activityMap["full"].(map[string]any)
	if value, ok := full["name"]; !ok || value != nil {
		t.Fatalf("full name = %#v present %v, want preserved nil", value, ok)
	}
}

func TestGetActivityDetailsMarksSyncChainStubsUnavailable(t *testing.T) {
	t.Parallel()

	wantWorkarounds := map[string]string{
		"1234567890": wantWahooStravaWorkaround,
		"2345678901": wantUnknownStravaWorkaround,
		"3456789012": wantUnknownStravaWorkaround,
	}
	for _, activity := range loadActivityFixtureFile(t, stravaSyncChainFixture) {
		t.Run(activity.ID, func(t *testing.T) {
			t.Parallel()
			client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity}
			tool := newGetActivityDetailsTool(client, client, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"` + activity.ID + `"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			activityMap := resultMap(t, result)["activity"].(map[string]any)
			if activityMap["strava_imported"] != true {
				t.Fatalf("activity = %#v, want strava_imported marker", activityMap)
			}
			assertUnavailableReasonAndWorkaround(t, activityMap, "strava_tos", wantWorkarounds[activity.ID])
		})
	}
}

func TestGetActivityDetailsResolvesGear(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"a1","icu_athlete_id":"i12345","name":"Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00","gear_id":"g-1"}`)
	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity, gear: decodeToolGear(t, `{"id":"g-1","name":"Race Bike"}`)}
	tool := newGetActivityDetailsToolWithGear(client, client, client, newGearListCache(), "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	activityMap := resultMap(t, result)["activity"].(map[string]any)
	if activityMap["gear_id"] != "g-1" || activityMap["gear_name"] != "Race Bike" || activityMap["gear_resolution"] != gearResolutionResolved {
		t.Fatalf("activity = %#v, want resolved gear", activityMap)
	}
}

func TestGetActivityDetailsMarksGearLookupUnavailable(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"a1","icu_athlete_id":"i12345","name":"Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00","gear_id":"g-1"}`)
	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity, gearErr: errors.New("gear upstream down")}
	tool := newGetActivityDetailsToolWithGear(client, client, client, newGearListCache(), "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	activityMap := resultMap(t, result)["activity"].(map[string]any)
	if activityMap["gear_id"] != "g-1" || activityMap["gear_resolution"] != gearResolutionLookupUnavailable {
		t.Fatalf("activity = %#v, want lookup_unavailable gear", activityMap)
	}
}

func TestGetActivityIntervalsCanonicalizesUnitsAndFullPayload(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{intervals: decodeIntervalsFixture(t, `{"id":"a123","analyzed":true,"top_null":null,"icu_intervals":[{"id":"i1","name":"Lap","unit":"MINS_KM","pace":4.2,"nullable":null},{"id":"i2","name":"Mystery","unit":"bananas"}],"icu_groups":[{"id":"g1","name":"Main"}]}`)}
	tool := newGetActivityIntervalsTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a123","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	resultPayload := resultMap(t, result)
	if value, ok := resultPayload["full"].(map[string]any)["top_null"]; !ok || value != nil {
		t.Fatalf("top-level full top_null = %#v present %v, want preserved nil", value, ok)
	}
	rows := resultPayload["intervals"].([]any)
	first := rows[0].(map[string]any)
	if first["unit"] != "MINS_KM" {
		t.Fatalf("first unit = %v, want canonical MINS_KM", first["unit"])
	}
	if value, ok := first["full"].(map[string]any)["nullable"]; !ok || value != nil {
		t.Fatalf("full nullable = %#v present %v, want preserved nil", value, ok)
	}
	second := rows[1].(map[string]any)
	if second["unit"] != "UNKNOWN" || !strings.Contains(second["unknown_unit"].(string), "bananas") {
		t.Fatalf("second row = %#v, want UNKNOWN with raw unit", second)
	}
}

func TestGetActivityIntervalsExposesCustomFieldsInTerseMode(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{intervals: decodeIntervalsFixture(t, `{"id":"a123","analyzed":true,"icu_intervals":[{"id":"i1","label":"Rep 1","lactate":3.8,"interval_note":"felt controlled","blood_sample_ok":true,"moving_time":180,"average_heartrate":165,"nullable":null,"custom_series":[1,2],"custom_object":{"value":1}}]}`)}
	tool := newGetActivityIntervalsTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a123"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["intervals"].([]any)
	first := rows[0].(map[string]any)
	if _, ok := first["full"]; ok {
		t.Fatalf("terse interval included full payload: %#v", first)
	}
	custom := first["custom_fields"].(map[string]any)
	if custom["lactate"] != float64(3.8) || custom["interval_note"] != "felt controlled" || custom["blood_sample_ok"] != true {
		t.Fatalf("custom_fields = %#v, want scalar upstream custom interval fields", custom)
	}
	for _, key := range []string{"moving_time", "average_heartrate", "nullable", "custom_series", "custom_object"} {
		if _, ok := custom[key]; ok {
			t.Fatalf("custom_fields included %q: %#v", key, custom)
		}
	}
}

func TestGetActivityIntervalsSourceMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fixture       string
		wantSource    string
		wantSuspected bool
	}{
		{name: "structured group", fixture: "structured.json", wantSource: "structured_workout"},
		{name: "one kilometer device laps", fixture: "auto_laps_1km.json", wantSource: "device_laps", wantSuspected: true},
		{name: "one mile device laps", fixture: "auto_laps_1mi.json", wantSource: "device_laps", wantSuspected: true},
		{name: "unknown source", fixture: "unknown.json", wantSource: "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeActivityReadClient{intervals: decodeActivityIntervalsFileFixture(t, tc.fixture)}
			tool := newGetActivityIntervalsTool(client, client, "test", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a123"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			meta := resultMap(t, result)["_meta"].(map[string]any)
			if meta["interval_source"] != tc.wantSource || meta["auto_lap_suspected"] != tc.wantSuspected {
				t.Fatalf("_meta = %#v, want source %q suspected %v", meta, tc.wantSource, tc.wantSuspected)
			}
		})
	}
}

func TestGetActivityIntervalsUnavailableReasons(t *testing.T) {
	t.Parallel()

	tests := activityReadUnavailableCases()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeActivityReadClient{activity: tc.fallbackActivity, activityErr: tc.fallbackErr, intervalErr: tc.upstreamErr}
			tool := newGetActivityIntervalsTool(client, client, "test", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v, want structured unavailable", err)
			}
			assertUnavailableReason(t, resultMap(t, result), tc.reason)
		})
	}
}

func TestGetActivityIntervalsUnavailableForHiddenSuccessPayload(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{intervals: decodeIntervalsFixture(t, `{"id":"stub1","icu_athlete_id":"i12345","start_date_local":"2026-01-02T07:00:00","name":null}`)}
	tool := newGetActivityIntervalsTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	assertUnavailableReasonAndWorkaround(t, payload, "strava_blocked", wantUnknownStravaWorkaround)
}

func TestGetActivityIntervalsFallbacksToDetailsForBlockedError(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{activity: decodeActivityFixture(t, `{"id":"stub1","source":"Strava","_note":"hidden"}`), intervalErr: intervals.ErrNotFound}
	tool := newGetActivityIntervalsTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	assertUnavailableReasonAndWorkaround(t, payload, "strava_blocked", wantUnknownStravaWorkaround)
}

type activityReadUnavailableCase struct {
	name             string
	upstreamErr      error
	fallbackActivity intervals.Activity
	fallbackErr      error
	reason           string
}

func activityReadUnavailableCases() []activityReadUnavailableCase {
	return []activityReadUnavailableCase{
		{name: "strava_blocked", upstreamErr: intervals.ErrUnauthorized, fallbackActivity: activityFixture(`{"id":"stub1","source":"Strava","_note":"hidden"}`), reason: "strava_blocked"},
		{name: "not_found", upstreamErr: intervals.ErrNotFound, fallbackErr: intervals.ErrNotFound, reason: "not_found"},
		{name: "unauthorized", upstreamErr: intervals.ErrUnauthorized, fallbackActivity: activityFixture(`{"id":"stub1","source":"Garmin"}`), reason: "unauthorized"},
		{name: "rate_limited", upstreamErr: intervals.ErrRateLimited, fallbackErr: intervals.ErrRateLimited, reason: "rate_limited"},
		{name: "upstream_unavailable_500", upstreamErr: &intervals.Error{StatusCode: 500, Kind: intervals.ErrUpstream}, fallbackErr: intervals.ErrUpstream, reason: "upstream_unavailable"},
		{name: "upstream_unavailable_400", upstreamErr: &intervals.Error{StatusCode: 400, Kind: intervals.ErrUpstream}, fallbackErr: intervals.ErrUpstream, reason: "upstream_unavailable"},
	}
}

func activityFixture(raw string) intervals.Activity {
	var activity intervals.Activity
	if err := json.Unmarshal([]byte(raw), &activity); err != nil {
		panic(err)
	}
	return activity
}

func assertUnavailableReason(t *testing.T, payload map[string]any, reason string) {
	t.Helper()
	if reason == "strava_blocked" {
		assertUnavailableReasonAndWorkaround(t, payload, reason, wantUnknownStravaWorkaround)
	} else {
		assertUnavailableReasonAndWorkaround(t, payload, reason, "")
	}
}

func assertUnavailableReasonAndWorkaround(t *testing.T, payload map[string]any, reason string, wantWorkaround string) {
	t.Helper()
	unavailable, ok := payload["unavailable"].(map[string]any)
	if !ok {
		t.Fatalf("payload = %#v, want unavailable object", payload)
	}
	if unavailable["reason"] != reason {
		t.Fatalf("unavailable = %#v, want reason %q", unavailable, reason)
	}
	if wantWorkaround != "" {
		if unavailable["workaround"] != wantWorkaround {
			t.Fatalf("unavailable = %#v, want workaround %q", unavailable, wantWorkaround)
		}
		if payload["strava_imported"] != true {
			t.Fatalf("payload = %#v, want strava_imported true for Strava-blocked activity", payload)
		}
	} else {
		if _, ok := unavailable["workaround"]; ok {
			t.Fatalf("unavailable = %#v, want no workaround for %s", unavailable, reason)
		}
		if payload["strava_imported"] == true {
			t.Fatalf("payload = %#v, want strava_imported only for Strava-blocked activity", payload)
		}
	}
	for _, key := range []string{"analyzed", "intervals", "groups", "streams", "splits", "metrics"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("payload = %#v, want no fabricated %s data on unavailable response", payload, key)
		}
	}
}

func decodeActivityFixture(t *testing.T, raw string) intervals.Activity {
	t.Helper()
	var activity intervals.Activity
	if err := json.Unmarshal([]byte(raw), &activity); err != nil {
		t.Fatalf("decode activity fixture: %v", err)
	}
	return activity
}

func decodeIntervalsFixture(t *testing.T, raw string) intervals.IntervalsDTO {
	t.Helper()
	var dto intervals.IntervalsDTO
	if err := json.Unmarshal([]byte(raw), &dto); err != nil {
		t.Fatalf("decode intervals fixture: %v", err)
	}
	return dto
}

func decodeActivityIntervalsFileFixture(t *testing.T, name string) intervals.IntervalsDTO {
	t.Helper()
	path := filepath.Join("testdata", "activity_intervals", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read intervals fixture %s: %v", path, err)
	}
	return decodeIntervalsFixture(t, string(data))
}

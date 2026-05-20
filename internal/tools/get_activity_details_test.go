package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeActivityReadClient struct {
	fakeProfileClient
	activity    intervals.Activity
	activityErr error
	intervals   intervals.IntervalsDTO
	intervalErr error
	streams     []intervals.ActivityStream
	streamErr   error
	messages    []intervals.ActivityMessage
	messageErr  error
	gear        []intervals.Gear
	gearErr     error
	gearCalls   int
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
	if unavailable := activityMap["unavailable"].(map[string]any); unavailable["reason"] != "strava_tos" {
		t.Fatalf("unavailable = %#v, want strava_tos", unavailable)
	}
	full := activityMap["full"].(map[string]any)
	if value, ok := full["name"]; !ok || value != nil {
		t.Fatalf("full name = %#v present %v, want preserved nil", value, ok)
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
	if payload["strava_imported"] != true || payload["unavailable"].(map[string]any)["reason"] != "strava_blocked" {
		t.Fatalf("payload = %#v, want Strava unavailable", payload)
	}
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
	if payload["strava_imported"] != true || payload["unavailable"].(map[string]any)["reason"] != "strava_blocked" {
		t.Fatalf("payload = %#v, want Strava unavailable fallback", payload)
	}
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
	unavailable, ok := payload["unavailable"].(map[string]any)
	if !ok {
		t.Fatalf("payload = %#v, want unavailable object", payload)
	}
	if unavailable["reason"] != reason {
		t.Fatalf("unavailable = %#v, want reason %q", unavailable, reason)
	}
	if reason == "strava_blocked" {
		workaround, ok := unavailable["workaround"].(string)
		if !ok || strings.TrimSpace(workaround) == "" {
			t.Fatalf("unavailable = %#v, want non-empty workaround for Strava-blocked activity", unavailable)
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

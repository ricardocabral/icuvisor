package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type progressionDetailsFake struct {
	activities map[string]intervals.Activity
	calls      []string
	err        error
}

func (f *progressionDetailsFake) GetActivity(_ context.Context, id string) (intervals.Activity, error) {
	f.calls = append(f.calls, id)
	if f.err != nil {
		return intervals.Activity{}, f.err
	}
	return f.activities[id], nil
}

type progressionIntervalsFake struct {
	rows  map[string]intervals.IntervalsDTO
	calls []string
	err   error
}

func (f *progressionIntervalsFake) GetActivityIntervals(_ context.Context, id string) (intervals.IntervalsDTO, error) {
	f.calls = append(f.calls, id)
	if f.err != nil {
		return intervals.IntervalsDTO{}, f.err
	}
	return f.rows[id], nil
}

type progressionExtendedFake struct {
	calls int
	value intervals.PowerVsHR
	err   error
}

func (f *progressionExtendedFake) GetActivityPowerVsHR(_ context.Context, _ string) (intervals.PowerVsHR, error) {
	f.calls++
	return f.value, f.err
}

type progressionProfileFake struct {
	calls int
	value intervals.AthleteWithSportSettings
	err   error
}

func (f *progressionProfileFake) GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error) {
	f.calls++
	return f.value, f.err
}

type progressionWellnessFake struct {
	calls int
	rows  []intervals.Wellness
	err   error
	param intervals.WellnessParams
}

func (f *progressionWellnessFake) ListWellness(_ context.Context, params intervals.WellnessParams) ([]intervals.Wellness, error) {
	f.calls++
	f.param = params
	return f.rows, f.err
}

type progressionEventFake struct {
	calls []string
	value intervals.Event
	err   error
}

func (f *progressionEventFake) GetEvent(_ context.Context, id string) (intervals.Event, error) {
	f.calls = append(f.calls, id)
	return f.value, f.err
}

func TestComputeWorkoutProgressionHandlerSuccessAndNoStreamCall(t *testing.T) {
	details := &progressionDetailsFake{activities: map[string]intervals.Activity{
		"a1": progressionActivity("a1", "2026-05-01T06:00:00", map[string]any{"steps": []any{map[string]any{"duration": 300, "power": map[string]any{"value": 200, "units": "W"}}, map[string]any{"description": "Recovery", "duration": 60}}}),
		"a2": progressionActivity("a2", "2026-05-02T06:00:00", map[string]any{"steps": []any{map[string]any{"duration": 300, "power": map[string]any{"value": 200, "units": "W"}}, map[string]any{"description": "Recovery", "duration": 60}}}),
	}}
	intervalsClient := &progressionIntervalsFake{rows: map[string]intervals.IntervalsDTO{
		"a1": structuredProgressionIntervals(), "a2": structuredProgressionIntervals(),
	}}
	extended := &progressionExtendedFake{value: intervals.PowerVsHR{PowerHR: float64Pointer(1.5), Decoupling: float64Pointer(2)}}
	tool := newComputeWorkoutProgressionTool(details, intervalsClient, extended, nil, nil, nil, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: computeWorkoutProgressionName, Arguments: json.RawMessage(`{"activities":[{"activity_id":"a1","label":"first"},{"activity_id":"a2","label":"second"}]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := result.StructuredContent.(map[string]any)
	body := payload["result"].(map[string]any)
	if body["status"] != "ok" {
		t.Fatalf("result status = %v, want ok: %#v", body["status"], body)
	}
	rows := body["rows"].([]any)
	if rows[0].(map[string]any)["activity_id"] != "a1" || rows[1].(map[string]any)["activity_id"] != "a2" {
		t.Fatalf("rows = %#v, want caller order", rows)
	}
	meta := payload["_meta"].(map[string]any)
	if !progressionAnyStringSliceContains(t, meta["source_tools"], "get_extended_metrics") || progressionAnyStringSliceContains(t, meta["source_tools"], "get_activity_streams") {
		t.Fatalf("source_tools = %#v, want narrow extended source and no streams", meta["source_tools"])
	}
	if extended.calls != 2 || len(details.calls) != 2 || len(intervalsClient.calls) != 2 {
		t.Fatalf("calls details=%d intervals=%d extended=%d, want two each", len(details.calls), len(intervalsClient.calls), extended.calls)
	}
}

func TestDecodeComputeWorkoutProgressionPresenceAndDuplicateValidation(t *testing.T) {
	for _, raw := range []string{
		`{"activities":[{"activity_id":"a1"},{"activity_id":"a2"}],"include_full":null}`,
		`{"activities":[{"activity_id":"a1"},{"activity_id":"a2"}],"include_readiness":"true"}`,
		`{"activities":[{"activity_id":"a1"},{"activity_id":" a1 "}]}`,
		`{"activities":[{"activity_id":"a1","event_id":""},{"activity_id":"a2"}]}`,
		`{"activities":[{"activity_id":"a1","label":null},{"activity_id":"a2"}]}`,
		`{"activities":[{"activity_id":"a1"},{"activity_id":"a2"}],"unknown":true}`,
	} {
		if _, err := decodeComputeWorkoutProgressionRequest(json.RawMessage(raw)); err == nil {
			t.Fatalf("decode(%s) error = nil, want validation error", raw)
		}
	}
}

func TestComputeWorkoutProgressionValidatesBeforeReads(t *testing.T) {
	details := &progressionDetailsFake{activities: map[string]intervals.Activity{}}
	tool := newComputeWorkoutProgressionTool(details, nil, nil, nil, nil, nil, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activities":[{"activity_id":"a1"},{"activity_id":" a1 "}]}`)})
	if err == nil || len(details.calls) != 0 {
		t.Fatalf("error=%v detail calls=%v, want validation error before reads", err, details.calls)
	}
}

func TestComputeWorkoutProgressionReadinessProfileAndBoundedWellness(t *testing.T) {
	details := &progressionDetailsFake{activities: map[string]intervals.Activity{
		"a1": progressionActivity("a1", "2026-05-01T06:00:00", nil),
		"a2": progressionActivity("a2", "2026-05-02T06:00:00", nil),
	}}
	intervalsClient := &progressionIntervalsFake{rows: map[string]intervals.IntervalsDTO{"a1": structuredProgressionIntervals(), "a2": structuredProgressionIntervals()}}
	profile := &progressionProfileFake{value: intervals.AthleteWithSportSettings{Timezone: "UTC"}}
	wellness := &progressionWellnessFake{rows: []intervals.Wellness{{ID: stringPointer("2026-05-01"), Raw: map[string]any{"id": "2026-05-01", "feel": float64(4)}}}}
	tool := newComputeWorkoutProgressionTool(details, intervalsClient, nil, nil, wellness, profile, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activities":[{"activity_id":"a1"},{"activity_id":"a2"}],"include_readiness":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if profile.calls != 1 || wellness.calls != 1 || wellness.param.Oldest != "2026-05-01" || wellness.param.Newest != "2026-05-02" {
		t.Fatalf("profile calls=%d wellness calls=%d params=%#v", profile.calls, wellness.calls, wellness.param)
	}
	meta := result.StructuredContent.(map[string]any)["_meta"].(map[string]any)
	if !progressionAnyStringSliceContains(t, meta["source_tools"], "get_athlete_profile") || meta["missing_days"] != float64(1) {
		t.Fatalf("meta = %#v, want profile source and one missing wellness day", meta)
	}
}

func TestComputeWorkoutProgressionEventFailureDoesNotFallback(t *testing.T) {
	details := &progressionDetailsFake{activities: map[string]intervals.Activity{"a1": progressionActivityWithRaw("a1", "2026-05-01T06:00:00", map[string]any{"workout_doc": map[string]any{"steps": []any{map[string]any{"duration": 300}}}, "paired_event_id": "e1"}), "a2": progressionActivity("a2", "2026-05-02T06:00:00", nil)}}
	events := &progressionEventFake{err: errors.New("event unavailable")}
	tool := newComputeWorkoutProgressionTool(details, &progressionIntervalsFake{rows: map[string]intervals.IntervalsDTO{"a1": structuredProgressionIntervals(), "a2": structuredProgressionIntervals()}}, nil, events, nil, nil, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activities":[{"activity_id":"a1"},{"activity_id":"a2"}]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := result.StructuredContent.(map[string]any)["result"].(map[string]any)["rows"].([]any)
	if !progressionContainsAnyString(rows[0].(map[string]any)["reasons"], "missing_prescription") || len(events.calls) != 1 {
		t.Fatalf("row/events = %#v/%#v, want no fallback and exact event read", rows[0], events.calls)
	}
}

func progressionActivity(id, date string, doc map[string]any) intervals.Activity {
	return progressionActivityWithRaw(id, date, map[string]any{"workout_doc": doc})
}

func progressionActivityWithRaw(id, date string, raw map[string]any) intervals.Activity {
	typ := "Ride"
	return intervals.Activity{ID: id, Type: &typ, StartDateLocal: stringPointer(date), MovingTime: progressionIntPointer(360), Feel: progressionIntPointer(3), RPE: progressionIntPointer(6), Raw: raw}
}

func structuredProgressionIntervals() intervals.IntervalsDTO {
	work := "work"
	recovery := "recovery"
	return intervals.IntervalsDTO{Raw: map[string]any{"source": "structured_workout"}, ICUIntervals: []intervals.ActivityInterval{{Type: &work, Duration: float64Pointer(300), AveragePower: float64Pointer(200), AverageHR: float64Pointer(140)}, {Type: &recovery, Duration: float64Pointer(60), AveragePower: float64Pointer(100), AverageHR: float64Pointer(120)}}}
}

func float64Pointer(value float64) *float64 { return &value }
func stringPointer(value string) *string    { return &value }
func progressionIntPointer(value int) *int  { return &value }

func progressionAnyStringSliceContains(t *testing.T, value any, want string) bool {
	t.Helper()
	rows, ok := value.([]any)
	if !ok {
		return false
	}
	for _, row := range rows {
		if row == want {
			return true
		}
	}
	return false
}

func progressionContainsAnyString(value any, want string) bool {
	rows, ok := value.([]any)
	if !ok {
		return false
	}
	for _, row := range rows {
		if row == want {
			return true
		}
	}
	return false
}

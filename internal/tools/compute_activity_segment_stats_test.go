package tools

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type segmentStatsStreamsClient struct {
	called bool
	params intervals.ActivityStreamsParams
	rows   []intervals.ActivityStream
}

func (c *segmentStatsStreamsClient) GetActivityStreams(_ context.Context, params intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error) {
	c.called = true
	c.params = params
	return c.rows, nil
}

func TestComputeActivitySegmentStatsDescriptionRoutesCollapsedLaps(t *testing.T) {
	t.Parallel()

	tool := newComputeActivitySegmentStatsTool(&segmentStatsStreamsClient{}, "test", false)
	description := strings.ToLower(tool.Description)
	for _, want := range []string{"single", "collapsed", "lap", "stream-based", "interval execution"} {
		if !strings.Contains(description, want) {
			t.Fatalf("description = %q, want routing hint %q", tool.Description, want)
		}
	}
}

func TestComputeActivitySegmentStatsHandlerScalarDistanceOmitsTimeFetch(t *testing.T) {
	client := &segmentStatsStreamsClient{rows: []intervals.ActivityStream{
		{Type: "distance", Data: []float64{0, 100, 200, 300}},
		{Type: "watts", Data: []float64{100, 200, 300, 400}},
	}}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	result, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"mean","metric":"watts","start_distance_m":100,"end_distance_m":300}`)})
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}
	payload := result.StructuredContent.(map[string]any)
	body := payload["result"].(map[string]any)
	if body["value"] != float64(300) || body["unit"] != "W" {
		t.Fatalf("result body = %#v, want mean 300 W", body)
	}
	if _, ok := payload["series"]; ok {
		t.Fatalf("terse payload has series: %#v", payload["series"])
	}
	if got := strings.Join(client.params.Types, ","); got != "distance,watts" {
		t.Fatalf("requested types = %q, want distance,watts", got)
	}
}

func TestComputeActivitySegmentStatsHandlerFirstAndLastDistanceSegments(t *testing.T) {
	rows := []intervals.ActivityStream{
		{Type: "distance", Data: []float64{0, 5000, 10000, 20000, 32195, 42195}},
		{Type: "velocity_smooth", Data: []float64{3.0, 3.1, 3.2, 3.3, 3.4, 3.5}},
		{Type: "watts", Data: []float64{180, 200, 220, 240, 260, 280}},
		{Type: "heartrate", Data: []float64{130, 135, 140, 145, 155, 160}},
	}
	cases := []struct {
		name       string
		metric     string
		start      float64
		end        float64
		wantValue  float64
		wantUnit   string
		wantTypes  string
		wantMetric string
	}{
		{name: "first 10 km pace velocity", metric: "velocity_smooth", start: 0, end: 10000, wantValue: 3.1, wantUnit: "m/s", wantTypes: "distance,velocity_smooth", wantMetric: "velocity_smooth"},
		{name: "last 10 km pace velocity", metric: "velocity_smooth", start: 32195, end: 42195, wantValue: 3.45, wantUnit: "m/s", wantTypes: "distance,velocity_smooth", wantMetric: "velocity_smooth"},
		{name: "first 10 km power", metric: "watts", start: 0, end: 10000, wantValue: 200, wantUnit: "W", wantTypes: "distance,watts", wantMetric: "watts"},
		{name: "last 10 km power", metric: "watts", start: 32195, end: 42195, wantValue: 270, wantUnit: "W", wantTypes: "distance,watts", wantMetric: "watts"},
		{name: "first 10 km heart rate", metric: "heart_rate", start: 0, end: 10000, wantValue: 135, wantUnit: "bpm", wantTypes: "distance,heartrate", wantMetric: "heart_rate"},
		{name: "last 10 km heart rate", metric: "heart_rate", start: 32195, end: 42195, wantValue: 157.5, wantUnit: "bpm", wantTypes: "distance,heartrate", wantMetric: "heart_rate"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &segmentStatsStreamsClient{rows: rows}
			handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
			args := map[string]any{"activity_id": "marathon1", "stat": "mean", "metric": tc.metric, "start_distance_m": tc.start, "end_distance_m": tc.end}
			encoded, err := json.Marshal(args)
			if err != nil {
				t.Fatalf("json.Marshal args error = %v", err)
			}
			result, err := handler(context.Background(), Request{Arguments: encoded})
			if err != nil {
				t.Fatalf("handler error = %v", err)
			}
			payload := result.StructuredContent.(map[string]any)
			if _, ok := payload["series"]; ok {
				t.Fatalf("terse payload has raw series: %#v", payload["series"])
			}
			body := payload["result"].(map[string]any)
			if body["value"] != tc.wantValue || body["unit"] != tc.wantUnit || body["metric"] != tc.wantMetric {
				t.Fatalf("result body = %#v, want %v %s for %s", body, tc.wantValue, tc.wantUnit, tc.wantMetric)
			}
			segment := body["segment"].(map[string]any)
			if segment["axis"] != "distance" || segment["start"] != tc.start || segment["end"] != tc.end {
				t.Fatalf("segment = %#v, want distance %.0f..%.0f", segment, tc.start, tc.end)
			}
			if got := strings.Join(client.params.Types, ","); got != tc.wantTypes {
				t.Fatalf("requested types = %q, want %s", got, tc.wantTypes)
			}
		})
	}
}

func TestComputeActivitySegmentStatsHandlerDistanceInsufficientStaysTerse(t *testing.T) {
	client := &segmentStatsStreamsClient{rows: []intervals.ActivityStream{
		{Type: "distance", Data: []float64{0, 5000, 10000}},
		{Type: "watts", Data: []float64{math.NaN(), math.NaN(), math.NaN()}},
	}}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	result, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"mean","metric":"watts","start_distance_m":0,"end_distance_m":10000}`)})
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}
	payload := result.StructuredContent.(map[string]any)
	if _, ok := payload["series"]; ok {
		t.Fatalf("terse insufficient payload has raw series: %#v", payload["series"])
	}
	meta := payload["_meta"].(map[string]any)
	if meta["insufficient_sample"] != true || meta["n"] != float64(0) {
		t.Fatalf("_meta = %#v, want explicit insufficient n=0", meta)
	}
	body := payload["result"].(map[string]any)
	if body["insufficient_sample"] != true || body["value"] != nil {
		t.Fatalf("result body = %#v, want insufficient without value", body)
	}
	if got := strings.Join(client.params.Types, ","); got != "distance,watts" {
		t.Fatalf("requested types = %q, want distance,watts", got)
	}
}

func TestComputeActivitySegmentStatsHandlerDistanceMissingStreamMessageIncludesKey(t *testing.T) {
	client := &segmentStatsStreamsClient{rows: []intervals.ActivityStream{{Type: "distance", Data: []float64{0, 5000, 10000}}}}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	_, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"mean","metric":"heart_rate","start_distance_m":0,"end_distance_m":10000}`)})
	if err == nil || !strings.Contains(err.Error(), "missing required activity stream for compute_activity_segment_stats: heart_rate") {
		t.Fatalf("handler error = %v, want missing heart_rate public message", err)
	}
	if got := strings.Join(client.params.Types, ","); got != "distance,heartrate" {
		t.Fatalf("requested types = %q, want distance,heartrate", got)
	}
}

func TestComputeActivitySegmentStatsHandlerRejectsFTPForNonIFWithoutFetch(t *testing.T) {
	client := &segmentStatsStreamsClient{}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	_, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"mean","metric":"watts","start_seconds":0,"end_seconds":10,"ftp_watts":0}`)})
	if err == nil || !strings.Contains(err.Error(), invalidActivitySegmentStatsMessage) {
		t.Fatalf("handler error = %v, want invalid arguments", err)
	}
	if client.called {
		t.Fatalf("GetActivityStreams called for ftp_watts on non-IF stat")
	}
}

func TestComputeActivitySegmentStatsHandlerDoesNotFetchInvalidArgs(t *testing.T) {
	client := &segmentStatsStreamsClient{}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	_, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"mean","metric":"watts","start_seconds":30,"end_seconds":10}`)})
	if err == nil || !strings.Contains(err.Error(), invalidActivitySegmentStatsMessage) {
		t.Fatalf("handler error = %v, want invalid arguments", err)
	}
	if client.called {
		t.Fatalf("GetActivityStreams called for locally invalid arguments")
	}
}

func TestComputeActivitySegmentStatsHandlerOutOfRangeMessage(t *testing.T) {
	client := &segmentStatsStreamsClient{rows: []intervals.ActivityStream{
		{Type: "time", Data: []float64{0, 10, 20}},
		{Type: "watts", Data: []float64{100, 200, 300}},
	}}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	_, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"mean","metric":"watts","start_seconds":0,"end_seconds":30}`)})
	if err == nil || !strings.Contains(err.Error(), "activity segment range is outside available stream coverage") {
		t.Fatalf("handler error = %v, want out-of-coverage message", err)
	}
}

func TestComputeActivitySegmentStatsHandlerMissingStreamMessageIncludesKey(t *testing.T) {
	client := &segmentStatsStreamsClient{rows: []intervals.ActivityStream{{Type: "time", Data: []float64{0, 10, 20}}}}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	_, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"mean","metric":"watts","start_seconds":0,"end_seconds":20}`)})
	if err == nil || !strings.Contains(err.Error(), "missing required activity stream for compute_activity_segment_stats: watts") {
		t.Fatalf("handler error = %v, want missing watts public message", err)
	}
}

func TestComputeActivitySegmentStatsHandlerAlignsInsufficientMeta(t *testing.T) {
	client := &segmentStatsStreamsClient{rows: []intervals.ActivityStream{
		{Type: "time", Data: []float64{0, 30, 35, 40}},
		{Type: "heartrate", Data: []float64{120, 130, 131, 132}},
	}}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	result, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"drift","start_seconds":0,"end_seconds":40}`)})
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}
	payload := result.StructuredContent.(map[string]any)
	meta := payload["_meta"].(map[string]any)
	if meta["source_tools"].([]any)[0] != getActivityStreamsName || meta["missing_days"] != float64(0) || meta["missing_action"] != "skip" || meta["method"] == "" {
		t.Fatalf("_meta mandatory fields = %#v", meta)
	}
	if meta["insufficient_sample"] != true {
		t.Fatalf("_meta.insufficient_sample = %#v, want true", meta["insufficient_sample"])
	}
	body := payload["result"].(map[string]any)
	if body["insufficient_sample"] != true {
		t.Fatalf("result.insufficient_sample = %#v, want true", body["insufficient_sample"])
	}
	segment := body["segment"].(map[string]any)
	if _, ok := segment["axis"]; !ok {
		t.Fatalf("segment = %#v, want snake_case json fields", segment)
	}
	if _, ok := segment["Axis"]; ok {
		t.Fatalf("segment = %#v, must not expose Go field names", segment)
	}
}

func TestComputeActivitySegmentStatsHandlerFullDecouplingAudit(t *testing.T) {
	client := &segmentStatsStreamsClient{rows: []intervals.ActivityStream{
		{Type: "time", Data: []float64{0, 10, 20, 30, 40, 50}},
		{Type: "heartrate", Data: []float64{100, 100, 100, 100, 100, 100}},
		{Type: "watts", Data: []float64{200, 200, 200, 180, 180, 180}},
	}}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	result, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"decoupling","start_seconds":0,"end_seconds":50,"include_full":true}`)})
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}
	payload := result.StructuredContent.(map[string]any)
	meta := payload["_meta"].(map[string]any)
	if meta["source_tools"].([]any)[0] != getActivityStreamsName || meta["formula_ref"] == "" {
		t.Fatalf("_meta = %#v, want source_tools and decoupling formula_ref", meta)
	}
	series := payload["series"].(map[string]any)
	if _, ok := series["watts_first_half"]; !ok {
		t.Fatalf("series = %#v, want watts_first_half audit", series)
	}
	if _, ok := series["heart_rate_second_half"]; !ok {
		t.Fatalf("series = %#v, want heart_rate_second_half audit", series)
	}
	if client.params.IncludeDefaults {
		t.Fatalf("IncludeDefaults = true, want false")
	}
	if got := strings.Join(client.params.Types, ","); got != "heartrate,time,watts" {
		t.Fatalf("requested types = %q, want heartrate,time,watts", got)
	}
}

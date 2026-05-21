package tools

import (
	"context"
	"encoding/json"
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

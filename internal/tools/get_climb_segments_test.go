package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type climbSegmentsClient struct {
	calls  int
	params intervals.ActivityStreamsParams
	rows   []intervals.ActivityStream
	err    error
}

func (c *climbSegmentsClient) GetActivityStreams(_ context.Context, params intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error) {
	c.calls++
	c.params = params
	return c.rows, c.err
}

func rawClimbRow(kind string, data any, allNull bool) intervals.ActivityStream {
	return intervals.ActivityStream{Type: kind, Raw: map[string]any{"type": kind, "data": data}, AllNull: allNull}
}

func TestGetClimbSegmentsRegistrationAndSchema(t *testing.T) {
	t.Parallel()

	tool := newGetClimbSegmentsTool(nil, "test", false)
	if tool.Name != getClimbSegmentsName || tool.Requirement.effective() != RequirementRead || tool.EffectiveToolset().String() != "full" {
		t.Fatalf("tool registration = %#v, want full/read get_climb_segments", tool)
	}
	if !strings.Contains(tool.Description, "do not fetch raw streams") || !strings.Contains(tool.Description, "full-toolset") {
		t.Fatalf("description = %q, want routing guidance", tool.Description)
	}
	schema := tool.InputSchema.(map[string]any)
	if schema["additionalProperties"] != false {
		t.Fatalf("schema additionalProperties = %#v, want false", schema["additionalProperties"])
	}
	properties := schema["properties"].(map[string]any)
	for _, name := range []string{"activity_id", "min_grade_percent", "min_elevation_gain_m", "max_gap_distance_m", "max_bridged_elevation_loss_m", "include_full"} {
		if _, ok := properties[name]; !ok {
			t.Fatalf("input schema missing %s: %#v", name, schema)
		}
	}
	for _, name := range []string{"min_grade_percent", "min_elevation_gain_m", "max_gap_distance_m", "max_bridged_elevation_loss_m"} {
		property := properties[name].(map[string]any)
		if property["type"] != "number" || !strings.Contains(property["description"].(string), "bounded") {
			t.Fatalf("parameter schema %s = %#v, want bounded number", name, property)
		}
	}
}

func TestGetClimbSegmentsRequestsExactCanonicalSourceStreams(t *testing.T) {
	client := &climbSegmentsClient{rows: []intervals.ActivityStream{
		rawClimbRow("distance", []any{0.0, 10.0, 20.0}, false),
		rawClimbRow("altitude", []any{100.0, 101.0, 102.0}, false),
		rawClimbRow("time", []any{0.0, 10.0, 20.0}, false),
		rawClimbRow("heartrate", []any{140.0, 145.0, 150.0}, false),
		rawClimbRow("watts", []any{200.0, 210.0, 220.0}, false),
	}}
	tool := newGetClimbSegmentsTool(client, t.Name(), t.Name() == "debug", responseShaping{})
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","min_elevation_gain_m":1,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if got := strings.Join(client.params.Types, ","); got != "distance,altitude,time,heartrate,watts" {
		t.Fatalf("requested streams = %q, want exact upstream order", got)
	}
	payload := result.StructuredContent.(map[string]any)
	body := payload["result"].(map[string]any)
	segments := body["segments"].([]any)
	if len(segments) != 1 {
		t.Fatalf("segments = %#v, want one concise segment", segments)
	}
	segment := segments[0].(map[string]any)
	if segment["average_heart_rate_bpm"] != float64(145) || segment["average_power_watts"] != float64(210) || segment["duration_seconds"] != float64(20) {
		t.Fatalf("segment = %#v, want optional metrics", segment)
	}
	if _, ok := payload["series"]; ok {
		t.Fatalf("include_full payload exposes series: %#v", payload)
	}
	meta := payload["_meta"].(map[string]any)
	if meta["n"] != float64(3) || meta["min_samples"] != float64(2) {
		t.Fatalf("meta = %#v, want normalized n=3 and min_samples=2", meta)
	}
}

func TestGetClimbSegmentsValidatesBeforeFetching(t *testing.T) {
	cases := []string{
		`{"activity_id":"a1","min_grade_percent":null}`,
		`{"activity_id":"a1","min_grade_percent":101}`,
		`{"activity_id":"a1","min_grade_percent":0}`,
		`{"activity_id":"a1","max_gap_distance_m":-1}`,
		`{"activity_id":"a1","unknown":true}`,
		`{"activity_id":"a1","include_full":null}`,
		`{"activity_id":""}`,
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			client := &climbSegmentsClient{}
			tool := newGetClimbSegmentsTool(client, t.Name(), t.Name() == "debug")
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)})
			if err == nil {
				t.Fatal("Handler() error = nil, want validation error")
			}
			if client.calls != 0 {
				t.Fatalf("stream calls = %d, want zero for invalid arguments", client.calls)
			}
		})
	}
}

func TestGetClimbSegmentsPreservesNullEvidenceInQuality(t *testing.T) {
	cases := []struct {
		name       string
		altitude   intervals.ActivityStream
		wantStatus string
	}{
		{name: "data null", altitude: rawClimbRow("altitude", nil, false), wantStatus: "null"},
		{name: "all null marker", altitude: intervals.ActivityStream{Type: "altitude", Raw: map[string]any{"type": "altitude", "allNull": true}, AllNull: true}, wantStatus: "null"},
		{name: "null element", altitude: rawClimbRow("altitude", []any{nil, nil}, false), wantStatus: "null"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &climbSegmentsClient{rows: []intervals.ActivityStream{
				rawClimbRow("distance", []any{0.0, 10.0}, false),
				tc.altitude,
			}}
			tool := newGetClimbSegmentsTool(client, t.Name(), t.Name() == "debug")
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			body := result.StructuredContent.(map[string]any)["result"].(map[string]any)
			if body["data_quality"].(map[string]any)["status"] != tc.wantStatus {
				t.Fatalf("body = %#v, want status %q", body, tc.wantStatus)
			}
			segments, ok := body["segments"].([]any)
			if !ok || len(segments) != 0 {
				t.Fatalf("segments = %#v, want stable empty array", body["segments"])
			}
		})
	}
}

func TestGetClimbSegmentsCountsMixedRawNullAndInvalidAltitude(t *testing.T) {
	client := &climbSegmentsClient{rows: []intervals.ActivityStream{
		rawClimbRow("distance", []any{0.0, 10.0, 20.0}, false),
		rawClimbRow("altitude", []any{nil, "bad", 2.0}, false),
	}}
	tool := newGetClimbSegmentsTool(client, t.Name(), t.Name() == "debug")
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	quality := result.StructuredContent.(map[string]any)["result"].(map[string]any)["data_quality"].(map[string]any)
	if quality["status"] != "null" || quality["null_altitude_samples"] != float64(1) || quality["invalid_altitude_samples"] != float64(1) {
		t.Fatalf("quality = %#v, want mixed null/non-finite counts", quality)
	}
}

func TestGetClimbSegmentsUnavailableIsShortUserError(t *testing.T) {
	client := &climbSegmentsClient{err: errors.New("authorization detail")}
	tool := newGetClimbSegmentsTool(client, t.Name(), t.Name() == "debug")
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err == nil || !strings.Contains(err.Error(), getClimbSegmentsFetchMsg) || strings.Contains(err.Error(), "authorization detail") {
		t.Fatalf("error = %v, want short user error without transport detail", err)
	}
}

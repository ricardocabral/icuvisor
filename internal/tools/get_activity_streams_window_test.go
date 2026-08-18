package tools

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestGetActivityStreamsTimeWindowSlicesCommonIndicesAndFiltersHelper(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{streams: decodeStreamFixtures(t,
		`{"type":"Time","data":[0,10,20,30,40]}`,
		`{"type":"Distance","data":[0,100,200,300,400]}`,
		`{"type":"Power","data":[250,260,270,280,290],"data2":[1,2,3,4,5],"extra":"kept"}`,
	)}
	tool := newGetActivityStreamsTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","keys":["Power"],"include_full":true,"time_window":{"start":10,"end":20}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	streamMap := payload["streams"].(map[string]any)
	if len(streamMap) != 1 {
		t.Fatalf("streams = %#v, want only requested watts row", streamMap)
	}
	watts := streamMap["watts"].(map[string]any)
	if got := watts["samples"]; !equalFloatSlices(got.([]any), []float64{260, 270}) {
		t.Fatalf("samples = %#v, want [260 270]", got)
	}
	if got := watts["data2"]; !equalFloatSlices(got.([]any), []float64{2, 3}) {
		t.Fatalf("data2 = %#v, want [2 3]", got)
	}
	if got := watts["full"].(map[string]any)["data"]; !equalFloatSlices(got.([]any), []float64{260, 270}) {
		t.Fatalf("full.data = %#v, want bounded data", got)
	}
	if got := watts["full"].(map[string]any)["data2"]; !equalFloatSlices(got.([]any), []float64{2, 3}) {
		t.Fatalf("full.data2 = %#v, want bounded data2", got)
	}
	if got := watts["source_sample_count"]; got != float64(5) {
		t.Fatalf("source_sample_count = %#v, want 5", got)
	}
	if got := watts["selected_sample_count"]; got != float64(2) {
		t.Fatalf("selected_sample_count = %#v, want 2", got)
	}
	if got := watts["returned_sample_count"]; got != float64(2) {
		t.Fatalf("returned_sample_count = %#v, want 2", got)
	}
	if got := watts["sampling_method"]; got != "window" {
		t.Fatalf("sampling_method = %#v, want window", got)
	}
	window := watts["window"].(map[string]any)
	timeWindow := window["time"].(map[string]any)
	if got := timeWindow["boundary_key"]; got != "time" || timeWindow["boundary_unit"] != "seconds" {
		t.Fatalf("time provenance = %#v", timeWindow)
	}
	if got := timeWindow["requested"].(map[string]any); got["start"] != float64(10) || got["end"] != float64(20) {
		t.Fatalf("requested time bounds = %#v", got)
	}
	if _, ok := streamMap["time"]; ok {
		t.Fatalf("streams = %#v, helper time must be filtered", streamMap)
	}
	if got, want := client.streamParams.Types, []string{"Power", "time"}; !equalStringSlices(got, want) {
		t.Fatalf("upstream types = %#v, want %#v", got, want)
	}
}

func TestGetActivityStreamsDistanceAndIntersectionWindows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      string
		wantPower []float64
		wantTime  []float64
		wantDist  []float64
	}{
		{name: "distance", args: `{"activity_id":"a1","include_full":true,"distance_window":{"start":100,"end":300}}`, wantPower: []float64{260, 270, 280}, wantTime: []float64{10, 20, 30}, wantDist: []float64{100, 200, 300}},
		{name: "intersection", args: `{"activity_id":"a1","keys":["Power"],"include_full":true,"time_window":{"start":15,"end":25},"distance_window":{"start":150,"end":250}}`, wantPower: []float64{270}},
		{name: "equal bound", args: `{"activity_id":"a1","include_full":true,"time_window":{"start":20,"end":20}}`, wantPower: []float64{270}, wantTime: []float64{20}, wantDist: []float64{200}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{streams: decodeStreamFixtures(t,
				`{"type":"time","data":[0,10,20,30,40]}`,
				`{"type":"distance","data":[0,100,200,300,400]}`,
				`{"type":"power","data":[250,260,270,280,290]}`,
			)}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			streamsMap := resultMap(t, result)["streams"].(map[string]any)
			assertWindowSamples(t, streamsMap, "watts", tc.wantPower)
			if len(tc.wantTime) > 0 {
				assertWindowSamples(t, streamsMap, "time", tc.wantTime)
			}
			if len(tc.wantDist) > 0 {
				assertWindowSamples(t, streamsMap, "distance", tc.wantDist)
			}
		})
	}
}

func TestGetActivityStreamsWindowEmptyAndMissingBoundaryDoNotLeakRaw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		streams    []intervals.ActivityStream
		args       string
		wantStatus string
		wantReason string
	}{
		{name: "empty", streams: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,2,3],"data2":[4,5,6]}`), args: `{"activity_id":"a1","include_full":true,"time_window":{"start":100,"end":200}}`, wantStatus: "empty"},
		{name: "missing boundary", streams: decodeStreamFixtures(t, `{"type":"power","data":[1,2,3],"data2":[4,5,6],"secret":"kept"}`), args: `{"activity_id":"a1","include_full":true,"time_window":{"start":0,"end":10}}`, wantStatus: "invalid", wantReason: "window_boundary_unavailable"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{streams: tc.streams}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			stream := payload["streams"].(map[string]any)["watts"].(map[string]any)
			if got := stream["source_sample_count"]; got != float64(3) {
				t.Fatalf("source_sample_count = %#v, want 3", got)
			}
			if got := stream["selected_sample_count"]; got != float64(0) {
				t.Fatalf("selected_sample_count = %#v, want 0", got)
			}
			if got := stream["returned_sample_count"]; got != float64(0) {
				t.Fatalf("returned_sample_count = %#v, want 0", got)
			}
			if _, ok := stream["samples"]; ok {
				t.Fatalf("stream = %#v, want no samples", stream)
			}
			window := stream["window"].(map[string]any)
			if tc.wantStatus == "empty" {
				full, ok := stream["full"].(map[string]any)
				if !ok || len(full["data"].([]any)) != 0 || len(full["data2"].([]any)) != 0 {
					t.Fatalf("full = %#v, want bounded empty arrays", stream["full"])
				}
			} else if _, ok := stream["full"]; ok {
				t.Fatalf("stream = %#v, want no full payload", stream)
			}
			if got := window["status"]; got != tc.wantStatus {
				t.Fatalf("window.status = %#v, want %q", got, tc.wantStatus)
			}
			wantMethod := "window"
			if tc.wantStatus == "invalid" {
				wantMethod = "unavailable"
			}
			if got := stream["sampling_method"]; got != wantMethod {
				t.Fatalf("sampling_method = %#v, want %q", got, wantMethod)
			}
			if got := window["empty"]; got != true {
				t.Fatalf("window.empty = %#v, want true", got)
			}
			if tc.wantStatus == "empty" {
				if _, ok := window["time"].(map[string]any)["effective"]; ok {
					t.Fatalf("window = %#v, wholly outside window should omit effective bounds", window)
				}
			} else {
				diagnostics := payload["_meta"].(map[string]any)["data_availability"].([]any)
				if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["reason"] != tc.wantReason {
					t.Fatalf("data_availability = %#v, want %q", diagnostics, tc.wantReason)
				}
			}
		})
	}
}

func TestGetActivityStreamsWindowMaxPointsKeepsData2Aligned(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{streams: decodeStreamFixtures(t,
		`{"type":"time","data":[0,10,20,30,40]}`,
		`{"type":"power","data":[250,260,270,280,290],"data2":[1,2,3,4,5]}`,
	)}
	tool := newGetActivityStreamsTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","keys":["Power"],"include_full":true,"time_window":{"start":0,"end":40},"max_points":3}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	stream := resultMap(t, result)["streams"].(map[string]any)["watts"].(map[string]any)
	if got := stream["samples"]; !equalFloatSlices(got.([]any), []float64{250, 270, 290}) {
		t.Fatalf("samples = %#v, want [250 270 290]", got)
	}
	if got := stream["data2"]; !equalFloatSlices(got.([]any), []float64{1, 3, 5}) {
		t.Fatalf("data2 = %#v, want [1 3 5]", got)
	}
	if got := stream["full"].(map[string]any)["data"]; !equalFloatSlices(got.([]any), []float64{250, 270, 290}) {
		t.Fatalf("full.data = %#v, want bounded samples", got)
	}
	if got := stream["selected_sample_count"]; got != float64(5) {
		t.Fatalf("selected_sample_count = %#v, want 5", got)
	}
	if got := stream["returned_sample_count"]; got != float64(3) || stream["sampling_method"] != "uniform_index" {
		t.Fatalf("sampling provenance = %#v, want returned 3 uniform_index", stream)
	}
}

func TestGetActivityStreamsWindowRejectsMalformedArgumentsBeforeFetch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args string
	}{
		{name: "reversed", args: `{"activity_id":"a1","time_window":{"start":20,"end":10}}`},
		{name: "negative", args: `{"activity_id":"a1","time_window":{"start":-1,"end":10}}`},
		{name: "too wide", args: `{"activity_id":"a1","time_window":{"start":0,"end":86401}}`},
		{name: "null window", args: `{"activity_id":"a1","time_window":null}`},
		{name: "missing start", args: `{"activity_id":"a1","time_window":{"end":10}}`},
		{name: "null end", args: `{"activity_id":"a1","time_window":{"start":0,"end":null}}`},
		{name: "wrong type", args: `{"activity_id":"a1","time_window":{"start":"0","end":10}}`},
		{name: "nested extra", args: `{"activity_id":"a1","time_window":{"start":0,"end":10,"unit":"s"}}`},
		{name: "top-level extra", args: `{"activity_id":"a1","time_window":{"start":0,"end":10},"unexpected":true}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if _, ok := PublicErrorMessage(err); !ok {
				t.Fatalf("PublicErrorMessage(%v) = _, false, want malformed argument error", err)
			}
			if client.streamCalls != 0 {
				t.Fatalf("GetActivityStreams calls = %d, want 0", client.streamCalls)
			}
		})
	}
}

func TestGetActivityStreamsWindowData2MismatchWithholdsChannel(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{streams: decodeStreamFixtures(t,
		`{"type":"time","data":[0,10,20]}`,
		`{"type":"power","data":[1,2,3],"data2":[4,5]}`,
	)}
	tool := newGetActivityStreamsTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","include_full":true,"time_window":{"start":0,"end":20}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	stream := payload["streams"].(map[string]any)["watts"].(map[string]any)
	if _, ok := stream["samples"]; ok {
		t.Fatalf("stream = %#v, want no samples for mismatched data2", stream)
	}
	if _, ok := stream["full"]; ok {
		t.Fatalf("stream = %#v, want no full payload for mismatched data2", stream)
	}
	if got := stream["sampling_method"]; got != "unavailable" {
		t.Fatalf("sampling_method = %#v, want unavailable", got)
	}
	diagnostics := payload["_meta"].(map[string]any)["data_availability"].([]any)
	if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["reason"] != "window_channel_length_mismatch" {
		t.Fatalf("data_availability = %#v, want data2 mismatch", diagnostics)
	}
}

func TestGetActivityStreamsWindowWithholdsNullSamples(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		power string
	}{
		{name: "data element", power: `{"type":"power","data":[1,null,3],"data2":[4,5,6]}`},
		{name: "data2 null", power: `{"type":"power","data":[1,2,3],"data2":null}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{streams: decodeStreamFixtures(t,
				`{"type":"time","data":[0,10,20]}`,
				tc.power,
			)}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","include_full":true,"time_window":{"start":0,"end":20}}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			stream := payload["streams"].(map[string]any)["watts"].(map[string]any)
			if _, ok := stream["samples"]; ok {
				t.Fatalf("stream = %#v, want null-containing samples withheld", stream)
			}
			if _, ok := stream["full"]; ok {
				t.Fatalf("stream = %#v, want null-containing full payload withheld", stream)
			}
			if got := stream["sampling_method"]; got != "unavailable" {
				t.Fatalf("sampling_method = %#v, want unavailable", got)
			}
			diagnostics := payload["_meta"].(map[string]any)["data_availability"].([]any)
			if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["reason"] != "window_channel_null" {
				t.Fatalf("data_availability = %#v, want window_channel_null", diagnostics)
			}
		})
	}
}

func TestGetActivityStreamsCombinedWindowKeepsBothProvenanceDimensionsOnFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		streams []intervals.ActivityStream
		args    string
		missing string
		other   string
	}{
		{name: "time fails first", streams: decodeStreamFixtures(t, `{"type":"distance","data":[0,100,200]}`, `{"type":"power","data":[1,2,3]}`), args: `{"activity_id":"a1","include_full":true,"time_window":{"start":0,"end":20},"distance_window":{"start":50,"end":150}}`, missing: "time", other: "distance"},
		{name: "distance fails after time", streams: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,2,3]}`), args: `{"activity_id":"a1","include_full":true,"time_window":{"start":0,"end":20},"distance_window":{"start":50,"end":150}}`, missing: "distance", other: "time"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{streams: tc.streams}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			stream := payload["streams"].(map[string]any)["watts"].(map[string]any)
			window := stream["window"].(map[string]any)
			if _, ok := window[tc.missing]; !ok {
				t.Fatalf("window = %#v, want missing %s provenance", window, tc.missing)
			}
			other, ok := window[tc.other].(map[string]any)
			if !ok || other["effective"] == nil {
				t.Fatalf("window = %#v, want effective %s provenance", window, tc.other)
			}
			if got := stream["sampling_method"]; got != "unavailable" {
				t.Fatalf("sampling_method = %#v, want unavailable", got)
			}
		})
	}
}

func TestGetActivityStreamsAdversarialBoundsAndMetadataOnlyWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       string
		want       []float64
		wantStart  float64
		wantEnd    float64
		wantStatus string
		full       bool
	}{
		{name: "clamps before source", args: `{"activity_id":"a1","include_full":true,"time_window":{"start":0,"end":25}}`, want: []float64{1, 2}, wantStart: 10, wantEnd: 25, wantStatus: "selected", full: true},
		{name: "clamps after source", args: `{"activity_id":"a1","include_full":true,"time_window":{"start":25,"end":100}}`, want: []float64{3}, wantStart: 25, wantEnd: 30, wantStatus: "selected", full: true},
		{name: "no overlap intersection", args: `{"activity_id":"a1","include_full":true,"time_window":{"start":0,"end":10},"distance_window":{"start":250,"end":350}}`, want: []float64{}, wantStart: 10, wantEnd: 10, wantStatus: "empty", full: true},
		{name: "metadata only", args: `{"activity_id":"a1","time_window":{"start":10,"end":25}}`, want: []float64{1, 2}, wantStart: 10, wantEnd: 25, wantStatus: "selected", full: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{streams: decodeStreamFixtures(t,
				`{"type":"time","data":[10,20,30]}`,
				`{"type":"distance","data":[0,100,200]}`,
				`{"type":"power","data":[1,2,3]}`,
			)}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			stream := resultMap(t, result)["streams"].(map[string]any)["watts"].(map[string]any)
			if tc.full {
				if len(tc.want) > 0 {
					assertWindowSamples(t, resultMap(t, result)["streams"].(map[string]any), "watts", tc.want)
				}
				full := stream["full"].(map[string]any)
				if got := full["data"].([]any); !equalFloatSlices(got, tc.want) {
					t.Fatalf("full.data = %#v, want %#v", got, tc.want)
				}
			} else if _, ok := stream["samples"]; ok {
				t.Fatalf("stream = %#v, want metadata-only row", stream)
			} else if _, ok := stream["full"]; ok {
				t.Fatalf("stream = %#v, want full omitted when include_full is false", stream)
			}
			if got := stream["sampling_method"]; got != "window" {
				t.Fatalf("sampling_method = %#v, want window", got)
			}
			window := stream["window"].(map[string]any)
			if got := window["status"]; got != tc.wantStatus {
				t.Fatalf("window.status = %#v, want %q", got, tc.wantStatus)
			}
			if got := stream["source_sample_count"]; got != float64(3) {
				t.Fatalf("source_sample_count = %#v, want 3", got)
			}
			selected := len(tc.want)
			if got := stream["selected_sample_count"]; got != float64(selected) {
				t.Fatalf("selected_sample_count = %#v, want %d", got, selected)
			}
			returned := selected
			if !tc.full {
				returned = 0
			}
			if got := stream["returned_sample_count"]; got != float64(returned) {
				t.Fatalf("returned_sample_count = %#v, want %d", got, returned)
			}
			timeWindow := window["time"].(map[string]any)
			effective := timeWindow["effective"].(map[string]any)
			if effective["start"] != tc.wantStart || effective["end"] != tc.wantEnd {
				t.Fatalf("effective bounds = %#v, want [%v %v]", effective, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

func TestGetActivityStreamsDistanceClampAndNonReducingCap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       string
		want       []float64
		wantStart  float64
		wantEnd    float64
		wantMethod string
	}{
		{name: "distance clamps before source", args: `{"activity_id":"a1","include_full":true,"distance_window":{"start":0,"end":250}}`, want: []float64{1, 2}, wantStart: 100, wantEnd: 250, wantMethod: "window"},
		{name: "cap larger than short selection", args: `{"activity_id":"a1","include_full":true,"max_points":4,"time_window":{"start":10,"end":20}}`, want: []float64{2, 3}, wantStart: 10, wantEnd: 20, wantMethod: "window"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{streams: decodeStreamFixtures(t,
				`{"type":"time","data":[0,10,20]}`,
				`{"type":"distance","data":[100,200,300]}`,
				`{"type":"power","data":[1,2,3]}`,
			)}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			stream := payload["streams"].(map[string]any)["watts"].(map[string]any)
			assertWindowSamples(t, payload["streams"].(map[string]any), "watts", tc.want)
			if got := stream["source_sample_count"]; got != float64(3) {
				t.Fatalf("source_sample_count = %#v, want 3", got)
			}
			if got := stream["selected_sample_count"]; got != float64(len(tc.want)) {
				t.Fatalf("selected_sample_count = %#v, want %d", got, len(tc.want))
			}
			if got := stream["returned_sample_count"]; got != float64(len(tc.want)) {
				t.Fatalf("returned_sample_count = %#v, want %d", got, len(tc.want))
			}
			if got := stream["sampling_method"]; got != tc.wantMethod {
				t.Fatalf("sampling_method = %#v, want %q", got, tc.wantMethod)
			}
			window := stream["window"].(map[string]any)
			var dimension map[string]any
			if _, ok := window["distance"]; ok {
				dimension = window["distance"].(map[string]any)
			} else {
				dimension = window["time"].(map[string]any)
			}
			effective := dimension["effective"].(map[string]any)
			if effective["start"] != tc.wantStart || effective["end"] != tc.wantEnd {
				t.Fatalf("effective bounds = %#v, want [%v %v]", effective, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

func TestGetActivityStreamsOmittedWindowRetainsLegacyMetadataShape(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{streams: decodeStreamFixtures(t, `{"type":"power","data":[1,2,3]}`)}
	tool := newGetActivityStreamsTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	stream := resultMap(t, result)["streams"].(map[string]any)["watts"].(map[string]any)
	for _, key := range []string{"source_sample_count", "selected_sample_count", "returned_sample_count", "sampling_method", "window", "samples", "full"} {
		if _, ok := stream[key]; ok {
			t.Fatalf("legacy metadata stream = %#v, want no %s", stream, key)
		}
	}
}

func TestGetActivityStreamsUnusableBoundaryMatrix(t *testing.T) {
	t.Parallel()

	validDistance := `{"type":"distance","data":[0,100,200]}`
	tests := []struct {
		name   string
		rows   []intervals.ActivityStream
		reason string
	}{
		{name: "missing", rows: decodeStreamFixtures(t, validDistance, `{"type":"power","data":[1,2,3]}`), reason: "window_boundary_unavailable"},
		{name: "null", rows: decodeStreamFixtures(t, `{"type":"time","data":null}`, validDistance, `{"type":"power","data":[1,2,3]}`), reason: "window_boundary_invalid"},
		{name: "empty", rows: decodeStreamFixtures(t, `{"type":"time","data":[]}`, validDistance, `{"type":"power","data":[1,2,3]}`), reason: "window_boundary_invalid"},
		{name: "all null", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,0,0],"allNull":true}`, validDistance, `{"type":"power","data":[1,2,3]}`), reason: "window_boundary_invalid"},
		{name: "null element", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,null,20]}`, validDistance, `{"type":"power","data":[1,2,3]}`), reason: "window_boundary_invalid"},
		{name: "decreasing", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,20,10]}`, validDistance, `{"type":"power","data":[1,2,3]}`), reason: "window_boundary_invalid"},
		{name: "non finite", rows: []intervals.ActivityStream{{Type: "time", Data: []float64{0, math.NaN(), 20}}, {Type: "distance", Data: []float64{0, 100, 200}}, {Type: "power", Data: []float64{1, 2, 3}}}, reason: "window_boundary_invalid"},
		{name: "boundary length mismatch", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"distance","data":[0,100]}`, `{"type":"power","data":[1,2,3]}`), reason: "window_boundary_length_mismatch"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{streams: tc.rows}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","include_full":true,"time_window":{"start":0,"end":20},"distance_window":{"start":0,"end":200},"keys":["Power"]}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			stream := payload["streams"].(map[string]any)["watts"].(map[string]any)
			if got := stream["source_sample_count"]; got != float64(3) {
				t.Fatalf("source_sample_count = %#v, want 3", got)
			}
			for _, key := range []string{"selected_sample_count", "returned_sample_count"} {
				if got := stream[key]; got != float64(0) {
					t.Fatalf("%s = %#v, want zero", key, got)
				}
			}
			if got := stream["sampling_method"]; got != "unavailable" {
				t.Fatalf("sampling_method = %#v, want unavailable", got)
			}
			window := stream["window"].(map[string]any)
			if window["status"] != "invalid" || window["empty"] != true {
				t.Fatalf("window = %#v, want invalid empty provenance", window)
			}
			if _, ok := window["time"]; !ok {
				t.Fatalf("window = %#v, want time provenance", window)
			}
			if _, ok := window["distance"]; !ok {
				t.Fatalf("window = %#v, want distance provenance", window)
			}
			if _, ok := stream["full"]; ok {
				t.Fatalf("stream = %#v, want full withheld", stream)
			}
			diagnostics := payload["_meta"].(map[string]any)["data_availability"].([]any)
			if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["reason"] != tc.reason {
				t.Fatalf("data_availability = %#v, want %q", diagnostics, tc.reason)
			}
		})
	}
}

func TestGetActivityStreamsRequestedChannelFailureMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		rows   []intervals.ActivityStream
		reason string
		source int
	}{
		{name: "data null", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":null}`), reason: "window_channel_null", source: 0},
		{name: "selected null", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,null,3]}`), reason: "window_channel_null", source: 3},
		{name: "all null", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,2,3],"allNull":true}`), reason: "window_channel_all_null", source: 3},
		{name: "data mismatch", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,2]}`), reason: "window_channel_length_mismatch", source: 2},
		{name: "data2 null", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,2,3],"data2":null}`), reason: "window_channel_null", source: 3},
		{name: "data2 empty", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,2,3],"data2":[]}`), reason: "window_channel_length_mismatch", source: 3},
		{name: "data2 selected null", rows: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,2,3],"data2":[4,null,6]}`), reason: "window_channel_null", source: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{streams: tc.rows}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","include_full":true,"keys":["Power"],"time_window":{"start":0,"end":20}}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			stream := payload["streams"].(map[string]any)["watts"].(map[string]any)
			if _, ok := stream["samples"]; ok {
				t.Fatalf("stream = %#v, want samples withheld", stream)
			}
			if _, ok := stream["full"]; ok {
				t.Fatalf("stream = %#v, want full withheld", stream)
			}
			if got := stream["source_sample_count"]; got != float64(tc.source) {
				t.Fatalf("source_sample_count = %#v, want %d", got, tc.source)
			}
			if got := stream["selected_sample_count"]; got != float64(0) || stream["returned_sample_count"] != float64(0) {
				t.Fatalf("zero channel provenance = %#v", stream)
			}
			if got := stream["sampling_method"]; got != "unavailable" {
				t.Fatalf("sampling_method = %#v, want unavailable", got)
			}
			diagnostics := payload["_meta"].(map[string]any)["data_availability"].([]any)
			if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["reason"] != tc.reason {
				t.Fatalf("data_availability = %#v, want %q", diagnostics, tc.reason)
			}
		})
	}
}

func TestGetActivityStreamsMissingRequestedChannelAndAbsentData2(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{streams: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,2,3]}`)}
	tool := newGetActivityStreamsTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","include_full":true,"keys":["heart_rate"],"time_window":{"start":0,"end":20}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	if len(payload["streams"].(map[string]any)) != 0 {
		t.Fatalf("streams = %#v, want absent requested channel", payload["streams"])
	}
	diagnostics := payload["_meta"].(map[string]any)["data_availability"].([]any)
	if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["reason"] != "missing_stream" {
		t.Fatalf("data_availability = %#v, want missing_stream", diagnostics)
	}

	client = &fakeActivityReadClient{streams: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","data":[1,2,3]}`)}
	tool = newGetActivityStreamsTool(client, client, "test", false)
	result, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","include_full":true,"time_window":{"start":0,"end":20}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	stream := resultMap(t, result)["streams"].(map[string]any)["watts"].(map[string]any)
	if _, ok := stream["data2"]; ok {
		t.Fatalf("stream = %#v, want absent data2 preserved as absent", stream)
	}
}

func TestGetActivityStreamsStravaUnavailableWindowUsesUnavailableShape(t *testing.T) {
	t.Parallel()

	for _, includeFull := range []bool{false, true} {
		t.Run(map[bool]string{false: "terse", true: "full"}[includeFull], func(t *testing.T) {
			client := &fakeActivityReadClient{activity: decodeExtendedMetricsActivity(t, `{"id":"stub1","source":"Strava","_note":"sanitized"}`), streamErr: intervals.ErrNotFound}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			args := `{"activity_id":"stub1","time_window":{"start":0,"end":20}}`
			if includeFull {
				args = `{"activity_id":"stub1","include_full":true,"time_window":{"start":0,"end":20}}`
			}
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			if payload["strava_imported"] != true || payload["unavailable"].(map[string]any)["reason"] != "strava_blocked" {
				t.Fatalf("payload = %#v, want Strava unavailable shape", payload)
			}
			if _, ok := payload["streams"]; ok {
				t.Fatalf("payload = %#v, want no fabricated streams", payload)
			}
			availability := payload["_meta"].(map[string]any)["data_availability"].([]any)
			if len(availability) != 1 || availability[0].(map[string]any)["reason"] != "restricted_source" {
				t.Fatalf("data_availability = %#v, want restricted_source", availability)
			}
			if includeFull {
				if _, ok := payload["full"]; !ok {
					t.Fatalf("payload = %#v, want gated full fallback", payload)
				}
			} else if _, ok := payload["full"]; ok {
				t.Fatalf("payload = %#v, want no full fallback", payload)
			}
		})
	}
}

func assertWindowSamples(t *testing.T, streamsMap map[string]any, key string, want []float64) {
	t.Helper()
	stream, ok := streamsMap[key].(map[string]any)
	if !ok {
		t.Fatalf("streams[%q] = %#v, want row", key, streamsMap[key])
	}
	got, ok := stream["samples"].([]any)
	if !ok || !equalFloatSlices(got, want) {
		t.Fatalf("streams[%q].samples = %#v, want %#v", key, stream["samples"], want)
	}
}

func equalStringSlices(got, want []string) bool {
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

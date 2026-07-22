package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func (f *fakeActivityReadClient) GetActivityStreams(ctx context.Context, params intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error) {
	f.streamCalls++
	f.streamParams = params
	return f.streams, f.streamErr
}

func TestGetActivityStreamsUnavailableReasons(t *testing.T) {
	t.Parallel()

	tests := activityReadUnavailableCases()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeActivityReadClient{activity: tc.fallbackActivity, activityErr: tc.fallbackErr, streamErr: tc.upstreamErr}
			tool := newGetActivityStreamsTool(client, client, "test", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v, want structured unavailable", err)
			}
			assertUnavailableReason(t, resultMap(t, result), tc.reason)
		})
	}
}

func TestGetActivityStreamsCanonicalizesKeysAndRequiresSamplesOptIn(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{streams: decodeStreamFixtures(t,
		`{"type":"Power","name":"Power","data":[250,260]}`,
		`{"type":"CustomThing","name":"CustomThing","data":[1]}`,
	)}
	tool := newGetActivityStreamsTool(client, client, "test", false)

	defaultResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err != nil {
		t.Fatalf("default Handler() error = %v", err)
	}
	defaultStream := resultMap(t, defaultResult)["streams"].(map[string]any)["watts"].(map[string]any)
	if _, ok := defaultStream["samples"]; ok {
		t.Fatalf("default stream = %#v, want no samples without keys/include_full", defaultStream)
	}

	keyedResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","keys":["Power","unknownThing"]}`)})
	if err != nil {
		t.Fatalf("keyed Handler() error = %v", err)
	}
	payload := resultMap(t, keyedResult)
	streamsMap := payload["streams"].(map[string]any)
	if _, ok := streamsMap["watts"].(map[string]any)["samples"]; ok {
		t.Fatalf("streams = %#v, want watts metadata without samples for explicit key", streamsMap)
	}
	if got := payload["_meta"].(map[string]any)["samples_included"]; got != false {
		t.Fatalf("_meta.samples_included = %#v, want false", got)
	}
	unknown := payload["_meta"].(map[string]any)["unknown_stream_keys"].([]any)
	if len(unknown) == 0 {
		t.Fatalf("_meta = %#v, want unknown stream keys", payload["_meta"])
	}

	fullResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","keys":["Power"],"include_full":true}`)})
	if err != nil {
		t.Fatalf("full Handler() error = %v", err)
	}
	fullPayload := resultMap(t, fullResult)
	fullStreams := fullPayload["streams"].(map[string]any)
	if got := fullStreams["watts"].(map[string]any)["samples"]; !equalFloatSlices(got.([]any), []float64{250, 260}) {
		t.Fatalf("streams = %#v, want watts samples [250 260]", fullStreams)
	}
	if got := fullPayload["_meta"].(map[string]any)["samples_included"]; got != true {
		t.Fatalf("_meta.samples_included = %#v, want true", got)
	}

	sampledClient := &fakeActivityReadClient{streams: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20,30,40]}`)}
	sampledTool := newGetActivityStreamsTool(sampledClient, sampledClient, "test", false)
	sampledResult, err := sampledTool.Handler(context.Background(), Request{Name: sampledTool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","include_full":true,"max_points":3}`)})
	if err != nil {
		t.Fatalf("sampled Handler() error = %v", err)
	}
	sampledStreams := resultMap(t, sampledResult)["streams"].(map[string]any)
	sampled := sampledStreams["time"].(map[string]any)
	if got := sampled["samples"]; !equalFloatSlices(got.([]any), []float64{0, 20, 40}) {
		t.Fatalf("samples = %#v, want [0 20 40]", got)
	}
	if got := sampled["sample_count"]; got != float64(5) {
		t.Fatalf("sample_count = %#v, want 5", got)
	}
	if got := sampled["returned_sample_count"]; got != float64(3) {
		t.Fatalf("returned_sample_count = %#v, want 3", got)
	}
	if got := sampled["sampling_method"]; got != "uniform_index" {
		t.Fatalf("sampling_method = %#v, want uniform_index", got)
	}
	if got := sampled["full"].(map[string]any)["data"]; !equalFloatSlices(got.([]any), []float64{0, 20, 40}) {
		t.Fatalf("full.data = %#v, want [0 20 40]", got)
	}
}

func TestGetActivityStreamsRejectsInvalidMaxPoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args string
	}{
		{name: "explicit zero", args: `{"activity_id":"a1","include_full":true,"max_points":0}`},
		{name: "below minimum", args: `{"activity_id":"a1","include_full":true,"max_points":1}`},
		{name: "above maximum", args: `{"activity_id":"a1","include_full":true,"max_points":5001}`},
		{name: "without include full", args: `{"activity_id":"a1","max_points":3}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeActivityReadClient{}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if _, ok := PublicErrorMessage(err); !ok {
				t.Fatalf("PublicErrorMessage(%v) = _, false, want NewUserError", err)
			}
			if client.streamCalls != 0 {
				t.Fatalf("GetActivityStreams calls = %d, want 0", client.streamCalls)
			}
		})
	}
}

func TestGetActivityStreamsReportsMissingHeartRateGuidance(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{streams: decodeStreamFixtures(t, `{"type":"time","data":[0,1]}`)}
	tool := newGetActivityStreamsTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","keys":["heart_rate"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	diagnostics := meta["data_availability"].([]any)
	if len(diagnostics) != 1 {
		t.Fatalf("data_availability = %#v, want one missing-stream diagnostic", diagnostics)
	}
	diagnostic := diagnostics[0].(map[string]any)
	if diagnostic["reason"] != "missing_stream" || !strings.Contains(diagnostic["message"].(string), "max-heart-rate") || !strings.Contains(diagnostic["workaround"].(string), "re-import") {
		t.Fatalf("diagnostic = %#v, want max-HR stream guidance", diagnostic)
	}
}

func TestGetActivityStreamsUnavailableIncludesRestrictedSourceDiagnostic(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{activity: decodeExtendedMetricsActivity(t, `{"id":"stub1","source":"Strava","_note":"hidden"}`), streamErr: intervals.ErrNotFound}
	tool := newGetActivityStreamsTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v, want structured unavailable", err)
	}
	payload := resultMap(t, result)
	assertUnavailableReason(t, payload, "strava_blocked")
	diagnostics := payload["_meta"].(map[string]any)["data_availability"].([]any)
	diagnostic := diagnostics[0].(map[string]any)
	if diagnostic["reason"] != "restricted_source" || !strings.Contains(diagnostic["message"].(string), "max-heart-rate") {
		t.Fatalf("diagnostic = %#v, want restricted source data guidance", diagnostic)
	}
}

func TestGetActivitySplitsUnavailableReasons(t *testing.T) {
	t.Parallel()

	tests := activityReadUnavailableCases()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}}, activity: tc.fallbackActivity, activityErr: tc.fallbackErr, intervalErr: intervals.ErrNotFound, streamErr: tc.upstreamErr}
			tool := newGetActivitySplitsTool(client, client, client, client, "test", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v, want structured unavailable", err)
			}
			assertUnavailableReason(t, resultMap(t, result), tc.reason)
		})
	}
}

func TestGetActivitySplitsComputesVirtualMetricAndImperial(t *testing.T) {
	t.Parallel()

	metricClient := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}}, streams: decodeStreamFixtures(t,
		`{"type":"distance","data":[0,1000,2000]}`,
		`{"type":"time","data":[0,300,620]}`,
	)}
	metricTool := newGetActivitySplitsTool(metricClient, metricClient, metricClient, metricClient, "test", false)
	metricResult, err := metricTool.Handler(context.Background(), Request{Name: metricTool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err != nil {
		t.Fatalf("metric Handler() error = %v", err)
	}
	metricPayload := resultMap(t, metricResult)
	if metricPayload["split_unit"] != "km" || len(metricPayload["splits"].([]any)) != 2 {
		t.Fatalf("metric payload = %#v, want two km splits", metricPayload)
	}

	imperialClient := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "miles"}}, streams: decodeStreamFixtures(t,
		`{"type":"distance","data":[0,1609.344]}`,
		`{"type":"time","data":[0,480]}`,
	)}
	imperialTool := newGetActivitySplitsTool(imperialClient, imperialClient, imperialClient, imperialClient, "test", false)
	imperialResult, err := imperialTool.Handler(context.Background(), Request{Name: imperialTool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if err != nil {
		t.Fatalf("imperial Handler() error = %v", err)
	}
	imperialPayload := resultMap(t, imperialResult)
	if imperialPayload["split_unit"] != "mi" || len(imperialPayload["splits"].([]any)) != 1 {
		t.Fatalf("imperial payload = %#v, want one mile split", imperialPayload)
	}
}

func decodeStreamFixtures(t *testing.T, raws ...string) []intervals.ActivityStream {
	t.Helper()
	out := make([]intervals.ActivityStream, 0, len(raws))
	for _, raw := range raws {
		var stream intervals.ActivityStream
		if err := json.Unmarshal([]byte(raw), &stream); err != nil {
			t.Fatalf("decode stream fixture: %v", err)
		}
		out = append(out, stream)
	}
	return out
}

func equalFloatSlices(got []any, want []float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i, wantValue := range want {
		if gotValue, ok := got[i].(float64); !ok || gotValue != wantValue {
			return false
		}
	}
	return true
}

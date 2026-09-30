package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func (f *fakeActivityReadClient) GetActivityStreams(ctx context.Context, params intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error) {
	f.streamCalls++
	f.streamParams = params
	f.streamParamHistory = append(f.streamParamHistory, params)
	key := strings.Join(params.Types, ",")
	if f.streamErrors != nil {
		if err, ok := f.streamErrors[key]; ok {
			return nil, err
		}
	}
	if f.streamResponses != nil {
		if rows, ok := f.streamResponses[key]; ok {
			return rows, nil
		}
	}
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

func TestActivityStreamToolsWithUpstreamAnomaliesAndHeartRateNames(t *testing.T) {
	t.Parallel()
	client := newActivityStreamsFixtureClient(t, "testdata/activity_streams_anomalies.json")
	tests := []struct {
		name string
		tool Tool
		args string
	}{
		{"segment control", newComputeActivitySegmentStatsTool(client, "test", false), `{"activity_id":"run-1","stat":"mean","metric":"heart_rate","start_seconds":0,"end_seconds":240}`},
		{"streams", newGetActivityStreamsTool(client, client, "test", false), `{"activity_id":"run-1"}`},
		{"streams full", newGetActivityStreamsTool(client, client, "test", false), `{"activity_id":"run-1","include_full":true}`},
		{"streams window", newGetActivityStreamsTool(client, client, "test", false), `{"activity_id":"run-1","include_full":true,"time_window":{"start":60,"end":180}}`},
		{"filtered streams", newGetActivityStreamsTool(client, client, "test", false), `{"activity_id":"run-1","keys":["heart_rate"],"include_full":true}`},
		{"histogram", newGetActivityHistogramTool(client, client, client, "test", false), `{"activity_id":"run-1","metric":"heart_rate_bpm"}`},
		{"splits", newGetActivitySplitsTool(client, client, client, client, "test", false), `{"activity_id":"run-1","split_unit":"km"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.tool.Handler(context.Background(), Request{Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatal(err)
			}
			payload := resultMap(t, result)
			if payload["unavailable"] != nil {
				t.Fatalf("unexpected unavailable: %s", resultText(t, result))
			}
			switch tc.name {
			case "segment control":
				if payload["result"].(map[string]any)["value"] != float64(120) {
					t.Fatalf("mean = %#v", payload)
				}
			case "streams", "filtered streams", "streams full", "streams window":
				hr := payload["streams"].(map[string]any)["heart_rate"].(map[string]any)
				if hr["type"] != "heartrate" {
					t.Fatalf("HR metadata = %#v", hr)
				}
				if tc.name == "filtered streams" || tc.name == "streams full" {
					if !equalFloatSlices(hr["samples"].([]any), []float64{100, 110, 130, 140}) {
						t.Fatalf("HR samples = %#v", hr)
					}
				} else if tc.name == "streams window" {
					if !equalFloatSlices(hr["samples"].([]any), []float64{110, 130}) {
						t.Fatalf("windowed HR = %#v", hr)
					}
				} else if _, ok := hr["samples"]; ok {
					t.Fatalf("terse response includes samples: %#v", hr)
				}
				if tc.name != "filtered streams" {
					meta := payload["_meta"].(map[string]any)
					diagnostics := meta["data_availability"].([]any)
					found := false
					for _, item := range diagnostics {
						d := item.(map[string]any)
						if d["reason"] == "channel_decode_failed" && reflect.DeepEqual(d["requested"], []any{"moving"}) {
							found = true
						}
					}
					if !found {
						t.Fatalf("missing moving-channel diagnostic: %#v", meta)
					}
					moving := payload["streams"].(map[string]any)["moving"].(map[string]any)
					if moving["samples"] != nil || moving["full"] != nil {
						t.Fatalf("invalid raw samples leaked: %#v", moving)
					}
				}
			case "histogram":
				if len(payload["buckets"].([]any)) == 0 {
					t.Fatalf("empty histogram: %#v", payload)
				}
			case "splits":
				splits := payload["splits"].([]any)
				if len(splits) != 2 || splits[0].(map[string]any)["average_heart_rate_bpm"] != float64(110) {
					t.Fatalf("splits = %#v", splits)
				}
			}
		})
	}
}

func TestGetActivityStreamsFetchDiagnosticsAreSafe(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(previous)
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"HTTP", &intervals.Error{StatusCode: 500, Kind: intervals.ErrUpstream}, "stream_http_error"},
		{"decode", &json.UnmarshalTypeError{Value: "string PRIVATE_SAMPLE", Type: reflect.TypeFor[bool](), Field: "custom"}, "stream_response_decode_failed"},
		{"syntax", &json.SyntaxError{Offset: 15}, "stream_response_decode_failed"},
		{"size", intervals.ErrResponseTooLarge, "stream_response_too_large"},
		{"transport", &url.Error{Op: "Get", URL: "https://example.invalid/PRIVATE_SAMPLE", Err: errors.New("PRIVATE_SAMPLE")}, "stream_transport_failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{activity: decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`), streamErr: fmt.Errorf("PRIVATE_SAMPLE: %w", tc.err)}
			result, err := newGetActivityStreamsTool(client, client, "test", false).Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
			if err != nil {
				t.Fatal(err)
			}
			payload := resultMap(t, result)
			meta := payload["_meta"].(map[string]any)
			diagnostics := meta["data_availability"].([]any)
			if diagnostics[0].(map[string]any)["reason"] != tc.want {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if strings.Contains(resultText(t, result), "PRIVATE_SAMPLE") || strings.Contains(logs.String(), "PRIVATE_SAMPLE") {
				t.Fatal("raw error leaked into response or logs")
			}
		})
	}
}

func newActivityStreamsFixtureClient(t *testing.T, fixturePath string) *intervals.Client {
	t.Helper()
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(fixture, &rows); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/athlete/i12345":
			_, _ = w.Write([]byte(`{"id":"i12345","preferred_units":"metric"}`))
		case "/activity/run-1":
			_, _ = w.Write([]byte(`{"id":"run-1","type":"Run"}`))
		case "/activity/run-1/intervals":
			_, _ = w.Write([]byte(`{"icu_intervals":[]}`))
		case "/activity/run-1/streams":
			types := r.URL.Query().Get("types")
			if types != "" && r.URL.Query().Get("includeDefaults") == "true" {
				t.Errorf("filtered request unexpectedly included defaults: %s", types)
			}
			selected := strings.Split(types, ",")
			for _, key := range selected {
				if key == "heart_rate" {
					http.Error(w, "unknown stream type", http.StatusBadRequest)
					return
				}
			}
			var out []json.RawMessage
			for _, row := range rows {
				var channel struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal(row, &channel); err != nil {
					t.Error(err)
					return
				}
				include := types == "" || r.URL.Query().Get("includeDefaults") == "true"
				for _, key := range selected {
					if key == channel.Type {
						include = true
					}
				}
				if include {
					out = append(out, row)
				}
			}
			_ = json.NewEncoder(w).Encode(out)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	client, err := intervals.NewClient(intervals.Options{Config: config.Config{APIKey: "test-key", AthleteID: "i12345", APIBaseURL: server.URL}, HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestActivityStreamToolsIssue64UpstreamScalarNulls(t *testing.T) {
	t.Parallel()
	client := newActivityStreamsFixtureClient(t, "testdata/activity_streams/issue64_nonzero_paused.json")
	cases := []struct {
		name  string
		tool  Tool
		args  string
		count int
		n     int
	}{
		{name: "unfiltered", tool: newGetActivityStreamsTool(client, client, "test", false), args: `{"activity_id":"run-1"}`},
		{name: "filtered HR", tool: newGetActivityStreamsTool(client, client, "test", false), args: `{"activity_id":"run-1","keys":["heart_rate","time"]}`},
		{name: "filtered distance", tool: newGetActivityStreamsTool(client, client, "test", false), args: `{"activity_id":"run-1","keys":["distance","time"]}`},
		{name: "full", tool: newGetActivityStreamsTool(client, client, "test", false), args: `{"activity_id":"run-1","include_full":true}`, count: 2858},
		{name: "sampled", tool: newGetActivityStreamsTool(client, client, "test", false), args: `{"activity_id":"run-1","include_full":true,"max_points":5}`, count: 5},
		{name: "window", tool: newGetActivityStreamsTool(client, client, "test", false), args: `{"activity_id":"run-1","include_full":true,"time_window":{"start":0,"end":120}}`, count: 121},
		{name: "sampled window", tool: newGetActivityStreamsTool(client, client, "test", false), args: `{"activity_id":"run-1","include_full":true,"max_points":5,"time_window":{"start":0,"end":120}}`, count: 5},
		{name: "HR histogram control", tool: newGetActivityHistogramTool(client, client, client, "test", false), args: `{"activity_id":"run-1","metric":"heart_rate_bpm"}`, n: 2857},
		{name: "pace histogram control", tool: newGetActivityHistogramTool(client, client, client, "test", false), args: `{"activity_id":"run-1","metric":"pace_seconds_per_km"}`, n: 2833},
		{name: "segment control", tool: newComputeActivitySegmentStatsTool(client, "test", false), args: `{"activity_id":"run-1","metric":"heart_rate","stat":"mean","start_seconds":0,"end_seconds":120}`, n: 121},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.tool.Handler(context.Background(), Request{Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatal(err)
			}
			p := resultMap(t, result)
			meta := p["_meta"].(map[string]any)
			if p["unavailable"] != nil || meta["data_availability"] != nil {
				t.Fatalf("valid scalar samples withheld: %s", resultText(t, result))
			}
			if tc.n > 0 {
				if meta["n"] != float64(tc.n) {
					t.Fatalf("n=%v, want %d", meta["n"], tc.n)
				}
				if tc.name == "segment control" && p["result"].(map[string]any)["value"] != float64(114) {
					t.Fatalf("wrong control mean: %#v", p)
				}
				return
			}
			rows := p["streams"].(map[string]any)
			for key, value := range rows {
				row := value.(map[string]any)
				if row["sampling_method"] == "unavailable" {
					t.Fatalf("%s unavailable", key)
				}
				if row["data2"] != nil {
					t.Fatalf("scalar %s has invented paired channel", key)
				}
				if tc.count == 0 {
					if row["samples"] != nil || row["full"] != nil {
						t.Fatalf("terse %s leaked samples", key)
					}
					continue
				}
				samples, ok := row["samples"].([]any)
				if !ok || len(samples) != tc.count {
					t.Fatalf("%s sample count=%d, want %d", key, len(samples), tc.count)
				}
				full := row["full"].(map[string]any)
				if len(full["data"].([]any)) != tc.count || full["data2"] != nil {
					t.Fatalf("invalid scalar full payload for %s", key)
				}
				if key == "time" {
					if samples[0] != float64(0) {
						t.Fatalf("first timestamp=%v", samples[0])
					}
					if tc.count == 5 && tc.name == "sampled" && samples[4] != float64(2857) {
						t.Fatalf("last timestamp=%v", samples[4])
					}
				}
			}
		})
	}
}

func TestGetActivitySplitsIssue64UpstreamNonzeroOrigin(t *testing.T) {
	t.Parallel()
	client := newActivityStreamsFixtureClient(t, "testdata/activity_streams/issue64_nonzero_paused.json")
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"run-1","split_unit":"km"}`)})
	if err != nil {
		t.Fatal(err)
	}
	p := resultMap(t, result)
	rows := p["splits"].([]any)
	if len(rows) != 6 {
		t.Fatalf("split count=%d, want six covered km splits; %s", len(rows), resultText(t, result))
	}
	for i, v := range rows {
		row := v.(map[string]any)
		if row["index"] != float64(i+2) || row["distance_km"] != float64(1) || row["duration_seconds"].(float64) <= 0 {
			t.Fatalf("invalid covered split: %#v", row)
		}
	}
	diagnostics := p["_meta"].(map[string]any)["data_availability"].([]any)
	if !hasSplitDiagnostic(diagnostics, "paused_samples_present") || !hasSplitDiagnostic(diagnostics, "initial_split_unavailable") || hasSplitDiagnostic(diagnostics, "insufficient_split_coverage") {
		t.Fatalf("wrong coverage diagnostics: %#v", diagnostics)
	}
}

func TestGetActivityStreamsDeclaredPairNullIsUnavailable(t *testing.T) {
	t.Parallel()
	for _, windowed := range []bool{false, true} {
		name, args, want := "unwindowed", `{"activity_id":"run-1","include_full":true}`, "channel_null"
		if windowed {
			name, args, want = "windowed", `{"activity_id":"run-1","include_full":true,"time_window":{"start":0,"end":20}}`, "window_channel_null"
		}
		t.Run(name, func(t *testing.T) {
			client := &fakeActivityReadClient{streams: decodeStreamFixtures(t, `{"type":"time","data":[0,10,20]}`, `{"type":"power","valueTypeIsArray":true,"data":[1,2,3],"data2":null}`)}
			tool := newGetActivityStreamsTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(args)})
			if err != nil {
				t.Fatal(err)
			}
			p := resultMap(t, result)
			row := p["streams"].(map[string]any)["watts"].(map[string]any)
			if row["samples"] != nil || row["full"] != nil || row["sampling_method"] != "unavailable" {
				t.Fatalf("invalid declared pair exposed: %#v", row)
			}
			if !hasSplitDiagnostic(p["_meta"].(map[string]any)["data_availability"].([]any), want) {
				t.Fatalf("missing %s diagnostic", want)
			}
		})
	}
}

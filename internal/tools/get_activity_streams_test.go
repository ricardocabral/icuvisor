package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func (f *fakeActivityReadClient) GetActivityStreams(ctx context.Context, params intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error) {
	return f.streams, nil
}

func TestGetActivityStreamsCanonicalizesKeysAndRequiresSamplesOptIn(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{streams: decodeStreamFixtures(t,
		`{"type":"Power","name":"Power","data":[250,260]}`,
		`{"type":"CustomThing","name":"CustomThing","data":[1]}`,
	)}
	tool := newGetActivityStreamsTool(client, "test", false)

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
	if _, ok := streamsMap["watts"].(map[string]any)["samples"]; !ok {
		t.Fatalf("streams = %#v, want watts samples for explicit key", streamsMap)
	}
	unknown := payload["_meta"].(map[string]any)["unknown_stream_keys"].([]any)
	if len(unknown) == 0 {
		t.Fatalf("_meta = %#v, want unknown stream keys", payload["_meta"])
	}
}

func TestGetActivitySplitsComputesVirtualMetricAndImperial(t *testing.T) {
	t.Parallel()

	metricClient := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}}, streams: decodeStreamFixtures(t,
		`{"type":"distance","data":[0,1000,2000]}`,
		`{"type":"time","data":[0,300,620]}`,
	)}
	metricTool := newGetActivitySplitsTool(metricClient, metricClient, metricClient, "test", false)
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
	imperialTool := newGetActivitySplitsTool(imperialClient, imperialClient, imperialClient, "test", false)
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

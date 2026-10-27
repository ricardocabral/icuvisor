package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type seasonPowerClient struct {
	sets  map[string]intervals.DataCurveSet
	calls []intervals.CurveParams
	err   error
}

func (c *seasonPowerClient) ListAthletePowerCurves(_ context.Context, p intervals.CurveParams) (intervals.DataCurveSet, error) {
	c.calls = append(c.calls, p)
	return c.sets[p.CurveSpec], c.err
}

func TestComparePowerCurvesSeasonsAndLocalBoundaries(t *testing.T) {
	client := &seasonPowerClient{sets: map[string]intervals.DataCurveSet{
		"s0,s0-kj0,s0-kj1": decodeCurveSet(t, []byte(`{"list":[{"id":"s0","label":"Current","start_date_local":"2026-03-01","end_date_local":"2026-10-26","secs":[7200],"values":[250],"activity_id":["watts"],"watts_per_kg":[4.25],"wkg_activity_id":["wkg"]},{"id":"s0-kj0","after_kj":1000,"secs":[7200],"values":[230]}],"activities":{"watts":{"start_date_local":"2026-03-01T00:01:00"},"wkg":{"start_date_local":"2026-10-26T23:59:00"}}}`)),
	}}
	tool := newComparePowerCurvesTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"as_of_date":"2026-10-26","season_offsets":[0,1],"duration_seconds":[7200,18000]}`)})
	if err != nil {
		t.Fatal(err)
	}
	got := resultMap(t, result)
	seasons := got["seasons"].([]any)
	if len(seasons) != 2 {
		t.Fatalf("seasons=%#v", seasons)
	}
	current := seasons[0].(map[string]any)
	for k, want := range map[string]any{"curve_spec": "s0", "season_offset": float64(0), "status": "available", "oldest": "2026-03-01", "newest": "2026-10-26"} {
		if current[k] != want {
			t.Errorf("current %s=%#v, want %#v", k, current[k], want)
		}
	}
	point := current["points"].([]any)[0].(map[string]any)
	if point["effort_date"] != "2026-03-01" || point["wkg_effort_date"] != "2026-10-26" {
		t.Errorf("local effort boundaries=%#v", point)
	}
	if len(current["durability_curves"].([]any)) != 1 {
		t.Fatalf("durability=%#v", current)
	}
	absent := seasons[1].(map[string]any)
	if absent["status"] != "unavailable" || absent["curve_spec"] != "s1" || len(absent["points"].([]any)) != 0 {
		t.Fatalf("absent season=%#v", absent)
	}
	if len(absent["missing_buckets"].([]any)) != 2 {
		t.Fatalf("missing=%#v", absent)
	}
	if _, present := absent["oldest"]; present {
		t.Fatalf("invented absent season boundary: %#v", absent)
	}
	if _, present := got["full"]; present {
		t.Fatal("raw arrays exposed without opt-in")
	}
	if len(client.calls) != 2 {
		t.Fatalf("calls=%#v", client.calls)
	}
	for _, p := range client.calls {
		if p.Now != "2026-10-26" {
			t.Errorf("local date anchor=%#v", p)
		}
	}
	meta := got["_meta"].(map[string]any)
	if meta["as_of_date"] != "2026-10-26" || meta["season_source"] != "upstream_season_selectors" {
		t.Fatalf("meta=%#v", meta)
	}
}

func TestComparePowerCurvesRejectsUnboundedOrAmbiguousInputs(t *testing.T) {
	cases := []struct{ name, args string }{
		{"missing date", `{}`},
		{"timestamp not local date", `{"as_of_date":"2026-10-27T00:00:00Z"}`},
		{"duplicate offsets", `{"as_of_date":"2026-10-27","season_offsets":[0,0]}`},
		{"negative offset", `{"as_of_date":"2026-10-27","season_offsets":[-1]}`},
		{"too old", `{"as_of_date":"2026-10-27","season_offsets":[11]}`},
		{"empty offsets", `{"as_of_date":"2026-10-27","season_offsets":[]}`},
		{"too many offsets", `{"as_of_date":"2026-10-27","season_offsets":[0,1,2,3,4]}`},
		{"invalid bucket", `{"as_of_date":"2026-10-27","duration_seconds":[0]}`},
		{"too many buckets", `{"as_of_date":"2026-10-27","duration_seconds":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &seasonPowerClient{}
			tool := newComparePowerCurvesTool(client, "test", false)
			_, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(tc.args)})
			if err == nil {
				t.Fatal("invalid request accepted")
			}
			if len(client.calls) != 0 {
				t.Fatalf("invalid input fetched data: %#v", client.calls)
			}
		})
	}
}

func TestComparePowerCurvesFullAndCancellation(t *testing.T) {
	t.Run("full defaults", func(t *testing.T) {
		client := &seasonPowerClient{sets: map[string]intervals.DataCurveSet{"s0,s0-kj0,s0-kj1": decodeCurveSet(t, []byte(`{"list":[],"activities":{}}`))}}
		tool := newComparePowerCurvesTool(client, "test", false)
		result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"as_of_date":"2026-10-27","include_full":true}`)})
		if err != nil {
			t.Fatal(err)
		}
		if len(client.calls) != 2 || client.calls[0].Sport != "Ride" {
			t.Fatalf("defaults=%#v", client.calls)
		}
		if _, ok := resultMap(t, result)["full"].(map[string]any); !ok {
			t.Fatal("requested full payload absent")
		}
	})
	t.Run("cancellation", func(t *testing.T) {
		client := &seasonPowerClient{err: context.Canceled}
		tool := newComparePowerCurvesTool(client, "test", false)
		_, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"as_of_date":"2026-10-27"}`)})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error=%v", err)
		}
		if len(client.calls) != 1 {
			t.Fatalf("continued after cancellation: %#v", client.calls)
		}
	})
}

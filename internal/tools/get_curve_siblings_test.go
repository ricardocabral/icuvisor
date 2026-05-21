package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCurveSiblingToolShapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tool       func(*fakeFitnessMetricsClient) Tool
		arguments  string
		bucketKey  string
		valueKey   string
		wantValue  float64
		wantMiss   float64
		metaBucket string
	}{
		{
			name:       "power duration curve",
			tool:       func(client *fakeFitnessMetricsClient) Tool { return newGetPowerCurvesTool(client, "test", false) },
			arguments:  `{"oldest":"2026-05-01","newest":"2026-05-07","sport":"Ride","duration_seconds":[60,999]}`,
			bucketKey:  "duration_seconds",
			valueKey:   "watts",
			wantValue:  420,
			wantMiss:   999,
			metaBucket: "duration_seconds",
		},
		{
			name:       "hr duration curve",
			tool:       func(client *fakeFitnessMetricsClient) Tool { return newGetHRCurvesTool(client, "test", false) },
			arguments:  `{"oldest":"2026-05-01","newest":"2026-05-07","sport":"Run","duration_seconds":[60,999]}`,
			bucketKey:  "duration_seconds",
			valueKey:   "heart_rate_bpm",
			wantValue:  182,
			wantMiss:   999,
			metaBucket: "duration_seconds",
		},
		{
			name: "pace distance curve",
			tool: func(client *fakeFitnessMetricsClient) Tool {
				return newGetPaceCurvesTool(client, client, "test", false)
			},
			arguments:  `{"oldest":"2026-05-01","newest":"2026-05-07","sport":"Run","distance_meters":[1000,9999]}`,
			bucketKey:  "distance_meters",
			valueKey:   "pace_seconds_per_km",
			wantValue:  230,
			wantMiss:   9999,
			metaBucket: "distance_meters",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newFakeFitnessMetricsClient(t)
			tool := tc.tool(client)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.arguments)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			got := resultMap(t, result)
			if _, ok := got["full"]; ok {
				t.Fatalf("full payload present in terse %s response", tool.Name)
			}
			points := got["points"].([]any)
			if len(points) != 1 {
				t.Fatalf("points = %#v, want one present bucket", points)
			}
			point := points[0].(map[string]any)
			if point[tc.bucketKey] == nil || point[tc.valueKey] != tc.wantValue || point["activity_id"] != "a1" {
				t.Fatalf("point = %#v, want %s=%v and activity a1", point, tc.valueKey, tc.wantValue)
			}
			meta := got["_meta"].(map[string]any)
			if len(meta[tc.metaBucket].([]any)) != 2 {
				t.Fatalf("meta buckets = %#v", meta[tc.metaBucket])
			}
			missing := meta["missing_buckets"].([]any)
			if len(missing) != 1 || missing[0] != tc.wantMiss {
				t.Fatalf("missing_buckets = %#v, want %v", missing, tc.wantMiss)
			}
		})
	}
}

func TestCurveSiblingIncludeFull(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tool      func(*fakeFitnessMetricsClient) Tool
		arguments string
	}{
		{name: "power", tool: func(client *fakeFitnessMetricsClient) Tool { return newGetPowerCurvesTool(client, "test", false) }, arguments: `{"oldest":"2026-05-01","newest":"2026-05-07","include_full":true,"duration_seconds":[60]}`},
		{name: "hr", tool: func(client *fakeFitnessMetricsClient) Tool { return newGetHRCurvesTool(client, "test", false) }, arguments: `{"oldest":"2026-05-01","newest":"2026-05-07","include_full":true,"sport":"Ride","duration_seconds":[60]}`},
		{name: "pace", tool: func(client *fakeFitnessMetricsClient) Tool {
			return newGetPaceCurvesTool(client, client, "test", false)
		}, arguments: `{"oldest":"2026-05-01","newest":"2026-05-07","include_full":true,"sport":"Run","distance_meters":[1000]}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newFakeFitnessMetricsClient(t)
			tool := tc.tool(client)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.arguments)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			got := resultMap(t, result)
			if _, ok := got["full"].(map[string]any); !ok {
				t.Fatalf("full payload missing from %s response: %#v", tool.Name, got)
			}
		})
	}
}

func TestPaceCurvesPreferredUnits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		preferredUnits string
		wantField      string
		wantValue      float64
		wantSystem     string
	}{
		{name: "metric", preferredUnits: "metric", wantField: "pace_seconds_per_km", wantValue: 230, wantSystem: "metric"},
		{name: "imperial", preferredUnits: "miles", wantField: "pace_seconds_per_mile", wantValue: 370.1, wantSystem: "imperial"},
		{name: "unknown falls back to metric", preferredUnits: "furlongs", wantField: "pace_seconds_per_km", wantValue: 230, wantSystem: "metric"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newFakeFitnessMetricsClient(t)
			client.profile.PreferredUnits = tc.preferredUnits
			tool := newGetPaceCurvesTool(client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-07","sport":"Run","distance_meters":[1000]}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			got := resultMap(t, result)
			point := got["points"].([]any)[0].(map[string]any)
			if point[tc.wantField] != tc.wantValue {
				t.Fatalf("point = %#v, want %s=%v", point, tc.wantField, tc.wantValue)
			}
			meta := got["_meta"].(map[string]any)
			units := meta["units"].(map[string]any)
			if units["system"] != tc.wantSystem {
				t.Fatalf("units = %#v, want system %s", units, tc.wantSystem)
			}
		})
	}
}

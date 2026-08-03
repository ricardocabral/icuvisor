package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	runningDynamicsActivityFixture  = "../../testdata/extended-metrics/activity-running-dynamics.json"
	runningDynamicsIntervalsFixture = "../../testdata/extended-metrics/activity-intervals-running-dynamics.json"
)

func TestExtendedMetricsDefersUnitUnverifiedRunningCadence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		includeFull bool
		cadenceMode string
	}{
		{name: "numeric terse", cadenceMode: "numeric"},
		{name: "numeric full", includeFull: true, cadenceMode: "numeric"},
		{name: "null full", includeFull: true, cadenceMode: "null"},
		{name: "absent full", includeFull: true, cadenceMode: "absent"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newFakeExtendedMetricsClient(t)
			client.activity = decodeActivityFileFixture(t, runningDynamicsActivityFixture)
			client.intervals = decodeIntervalsFileFixture(t, runningDynamicsIntervalsFixture)
			setRunningCadenceFixtureMode(t, &client.activity.Raw, &client.intervals, tc.cadenceMode)
			tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

			arguments := `{"activity_id":"activity-running-dynamics-fixture"}`
			if tc.includeFull {
				arguments = `{"activity_id":"activity-running-dynamics-fixture","include_full":true}`
			}
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(arguments)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			assertRunningCadenceIsTerseAbsent(t, payload)

			if !tc.includeFull {
				if _, ok := payload["full"]; ok {
					t.Fatalf("full present in terse response: %#v", payload)
				}
				return
			}
			assertRunningCadenceRawMode(t, payload, tc.cadenceMode)
		})
	}
}

func setRunningCadenceFixtureMode(t *testing.T, activity *map[string]any, dto *intervals.IntervalsDTO, mode string) {
	t.Helper()
	intervalsRaw, ok := dto.Raw["icu_intervals"].([]any)
	if !ok || len(intervalsRaw) != 1 {
		t.Fatalf("fixture intervals = %#v, want one raw interval", dto.Raw)
	}
	intervalRaw, ok := intervalsRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("fixture interval = %#v, want object", intervalsRaw[0])
	}
	if len(dto.ICUIntervals) != 1 {
		t.Fatalf("decoded intervals = %#v, want one interval", dto.ICUIntervals)
	}

	applyCadenceMode(*activity, []string{"average_cadence"}, mode)
	applyCadenceMode(intervalRaw, runningCadenceKeys, mode)
	applyCadenceMode(dto.ICUIntervals[0].Raw, runningCadenceKeys, mode)
}

var runningCadenceKeys = []string{"average_cadence", "min_cadence", "max_cadence"}

func applyCadenceMode(raw map[string]any, keys []string, mode string) {
	for _, key := range keys {
		switch mode {
		case "numeric":
		case "null":
			raw[key] = nil
		case "absent":
			delete(raw, key)
		default:
			panic("unknown cadence fixture mode")
		}
	}
}

func assertRunningCadenceIsTerseAbsent(t *testing.T, payload map[string]any) {
	t.Helper()
	metrics := payload["metrics"].(map[string]any)
	for _, key := range runningCadenceKeys {
		if _, ok := metrics[key]; ok {
			t.Fatalf("terse metrics included unit-unverified %s: %#v", key, metrics)
		}
	}
	if _, ok := payload["intervals"]; ok {
		t.Fatalf("cadence-only interval was surfaced in terse response: %#v", payload["intervals"])
	}

	meta := payload["_meta"].(map[string]any)
	units := meta["extended_metric_units"].(map[string]any)
	for _, key := range runningCadenceKeys {
		if _, ok := units[key]; ok {
			t.Fatalf("extended_metric_units included unit-unverified %s: %#v", key, units)
		}
	}
	dropped := meta["dropped_fields"].([]any)
	for _, key := range droppedExtendedMetricFields {
		if !containsAnyString(dropped, key) {
			t.Fatalf("dropped_fields missing %s: %#v", key, dropped)
		}
		if _, ok := metrics[key]; ok {
			t.Fatalf("dropped field %s was surfaced: %#v", key, metrics)
		}
	}
}

func assertRunningCadenceRawMode(t *testing.T, payload map[string]any, mode string) {
	t.Helper()
	full := payload["full"].(map[string]any)
	activity := full["activity"].(map[string]any)
	intervalsRaw := full["intervals"].(map[string]any)
	rows := intervalsRaw["icu_intervals"].([]any)
	interval := rows[0].(map[string]any)

	assertCadenceRawValue(t, activity, "average_cadence", mode, float64(176))
	assertCadenceRawValue(t, interval, "average_cadence", mode, float64(176))
	assertCadenceRawValue(t, interval, "min_cadence", mode, float64(164))
	assertCadenceRawValue(t, interval, "max_cadence", mode, float64(188))
}

func assertCadenceRawValue(t *testing.T, raw map[string]any, key, mode string, want float64) {
	t.Helper()
	value, present := raw[key]
	switch mode {
	case "numeric":
		if !present || value != want {
			t.Fatalf("raw %s = (%v, %t), want (%v, true)", key, value, present, want)
		}
	case "null":
		if !present || value != nil {
			t.Fatalf("raw %s = (%v, %t), want (nil, true)", key, value, present)
		}
	case "absent":
		if present {
			t.Fatalf("raw %s = %v, want absent", key, value)
		}
	default:
		t.Fatalf("unknown cadence fixture mode %q", mode)
	}
}

func containsAnyString(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

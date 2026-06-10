package analysis

import "testing"

func TestInferIntervalSource(t *testing.T) {
	km := 1000.0
	mile := 1609.344
	tests := []struct {
		name string
		in   IntervalSourceInput
		want IntervalSourceResult
	}{
		{
			name: "structured groups take precedence over uniform laps",
			in: IntervalSourceInput{
				Groups:    []IntervalSourceGroup{{Name: "Main"}},
				Intervals: genericDistanceIntervals(6, km),
			},
			want: IntervalSourceResult{Source: IntervalSourceStructuredWorkout},
		},
		{
			name: "structured interval names take precedence over uniform distances",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				distanceInterval("Warmup", 0, 1000),
				distanceInterval("Work 1", 1000, 2000),
				distanceInterval("Recovery", 2000, 3000),
				distanceInterval("Work 2", 3000, 4000),
				distanceInterval("Cooldown", 4000, 5000),
			}},
			want: IntervalSourceResult{Source: IntervalSourceStructuredWorkout},
		},
		{
			name: "all evidence-bearing raw intervals without group markers are manual added",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				{Name: "Lap", Raw: map[string]any{"id": "interval-1", "start_index": 0, "end_index": 100}},
				{Name: "Lap", Raw: map[string]any{"id": "interval-2", "start_index": 100, "end_index": 200}},
			}},
			want: IntervalSourceResult{Source: IntervalSourceManualAdded},
		},
		{
			name: "raw intervals with and without group markers are mixed",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				{Name: "Lap", Raw: map[string]any{"id": "interval-1", "group_id": "group-a"}},
				{Name: "Lap", Raw: map[string]any{"id": "interval-2", "start_index": 100, "end_index": 200}},
			}},
			want: IntervalSourceResult{Source: IntervalSourceMixed},
		},
		{
			name: "all raw intervals with group markers preserve uniform device fallback",
			in:   IntervalSourceInput{Intervals: groupedRawDistanceIntervals(6, km)},
			want: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true},
		},
		{
			name: "one kilometer generic auto laps",
			in:   IntervalSourceInput{Intervals: genericDistanceIntervals(6, km)},
			want: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true},
		},
		{
			name: "missing raw evidence does not classify as manual added",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				{Name: "Lap", Raw: nil},
				{Name: "Lap", Raw: map[string]any{}},
			}},
			want: IntervalSourceResult{Source: IntervalSourceUnknown},
		},
		{
			name: "one mile generic auto laps",
			in:   IntervalSourceInput{Intervals: genericDistanceIntervals(6, mile)},
			want: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true},
		},
		{
			name: "edge partials can be excluded",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				distanceInterval("Lap", 0, 400),
				distanceInterval("Lap 1", 400, 1402),
				distanceInterval("Lap 2", 1402, 2401),
				distanceInterval("Lap 3", 2401, 3400),
				distanceInterval("Lap 4", 3400, 4403),
				distanceInterval("Lap", 4403, 4700),
			}},
			want: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true},
		},
		{
			name: "insufficient rows remain unknown",
			in:   IntervalSourceInput{Intervals: genericDistanceIntervals(3, km)},
			want: IntervalSourceResult{Source: IntervalSourceUnknown},
		},
		{
			name: "mixed target distances remain unknown",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				distanceInterval("Lap 1", 0, 1000),
				distanceInterval("Lap 2", 1000, 2300),
				distanceInterval("Lap 3", 2300, 4050),
				distanceInterval("Lap 4", 4050, 4550),
				distanceInterval("Lap 5", 4550, 5450),
			}},
			want: IntervalSourceResult{Source: IntervalSourceUnknown},
		},
		{
			name: "non generic custom names remain unknown despite uniform distances",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				distanceInterval("Set A", 0, 1000),
				distanceInterval("Set B", 1000, 2000),
				distanceInterval("Set C", 2000, 3000),
				distanceInterval("Set D", 3000, 4000),
			}},
			want: IntervalSourceResult{Source: IntervalSourceUnknown},
		},
		{
			name: "explicit workout step marker is structured",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				{Name: "Lap", Raw: map[string]any{"workout_step_id": "step-1"}},
			}},
			want: IntervalSourceResult{Source: IntervalSourceStructuredWorkout},
		},
		{
			name: "explicit workout step marker takes precedence over group-id heuristic",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				{Name: "Lap", Raw: map[string]any{"id": "interval-1", "workout_step_id": "step-1"}},
				{Name: "Lap", Raw: map[string]any{"id": "interval-2", "group_id": "group-a"}},
			}},
			want: IntervalSourceResult{Source: IntervalSourceStructuredWorkout},
		},
		{
			name: "explicit auto lap marker is device laps",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				{Name: "Lap", Raw: map[string]any{"lap_source": "device auto lap"}},
			}},
			want: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true},
		},
		{
			name: "explicit boolean auto lap marker is device laps",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				{Name: "Lap", Raw: map[string]any{"auto_lap": true}},
			}},
			want: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true},
		},
		{
			name: "explicit auto lap type marker is device laps",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				{Name: "Lap", Raw: map[string]any{"lap_type": "auto"}},
			}},
			want: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true},
		},
		{
			name: "explicit auto lap marker takes precedence over group-id heuristic",
			in: IntervalSourceInput{Intervals: []IntervalSourceInterval{
				{Name: "Lap", Raw: map[string]any{"id": "interval-1", "lap_source": "device auto lap"}},
				{Name: "Lap", Raw: map[string]any{"id": "interval-2"}},
			}},
			want: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true},
		},
		{
			name: "uniform duration laps",
			in:   IntervalSourceInput{Intervals: genericDurationIntervals(5, 300)},
			want: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InferIntervalSource(tc.in)
			if got != tc.want {
				t.Fatalf("InferIntervalSource() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func genericDistanceIntervals(count int, distance float64) []IntervalSourceInterval {
	out := make([]IntervalSourceInterval, 0, count)
	for i := range count {
		start := float64(i) * distance
		out = append(out, distanceInterval("Lap", start, start+distance))
	}
	return out
}

func genericDurationIntervals(count int, duration float64) []IntervalSourceInterval {
	out := make([]IntervalSourceInterval, 0, count)
	for i := range count {
		startIndex := i * 100
		endIndex := startIndex + 100
		d := duration
		out = append(out, IntervalSourceInterval{Name: "Lap", Duration: &d, StartIndex: &startIndex, EndIndex: &endIndex})
	}
	return out
}

func groupedRawDistanceIntervals(count int, distance float64) []IntervalSourceInterval {
	out := genericDistanceIntervals(count, distance)
	for i := range out {
		out[i].Raw = map[string]any{"id": i + 1, "group_id": "auto-group"}
	}
	return out
}

func distanceInterval(name string, start float64, end float64) IntervalSourceInterval {
	distance := end - start
	return IntervalSourceInterval{Name: name, StartDistance: &start, EndDistance: &end, Distance: &distance}
}

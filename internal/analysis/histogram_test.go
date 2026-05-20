package analysis

import (
	"math"
	"testing"
)

func TestBuildHistogramConfiguredZones(t *testing.T) {
	t.Parallel()

	got := BuildHistogram([]HistogramSample{
		{Value: 90, Seconds: 30},
		{Value: 100, Seconds: 30},
		{Value: 149.9, Seconds: 15},
		{Value: 150, Seconds: 20},
		{Value: 220, Seconds: 5},
	}, "W", &HistogramZoneConfig{Sport: "Ride", SportSettingID: 7, Metric: string(HistogramMetricPowerWatts), Boundaries: []float64{150, 100}, Names: []string{"Tempo", "Endurance"}, Unit: "W"})

	if got.BucketMethod != BucketMethodConfiguredZones {
		t.Fatalf("BucketMethod = %q, want configured_zones", got.BucketMethod)
	}
	if got.N != 5 {
		t.Fatalf("N = %d, want 5", got.N)
	}
	if len(got.Buckets) != 3 {
		t.Fatalf("buckets = %d, want 3: %#v", len(got.Buckets), got.Buckets)
	}
	assertBucket(t, got.Buckets[0], "Below Endurance", nil, ptr(100), 30, 30)
	assertBucket(t, got.Buckets[1], "Endurance", ptr(100), ptr(150), 45, 45)
	assertBucket(t, got.Buckets[2], "Tempo", ptr(150), nil, 25, 25)
	if got.ZoneSource == nil || got.ZoneSource.Boundaries[0] != 100 || got.ZoneSource.Boundaries[1] != 150 {
		t.Fatalf("ZoneSource = %#v, want sorted boundaries preserving names", got.ZoneSource)
	}
}

func TestBuildHistogramFixedWidth(t *testing.T) {
	t.Parallel()

	got := BuildHistogram([]HistogramSample{{Value: 0, Seconds: 1}, {Value: 5, Seconds: 1}, {Value: 10, Seconds: 1}, {Value: math.Inf(1), Seconds: 1}, {Value: 3, Seconds: -1}}, "bpm", nil)
	if got.BucketMethod != BucketMethodFixedWidth {
		t.Fatalf("BucketMethod = %q, want fixed_width", got.BucketMethod)
	}
	if got.N != 3 {
		t.Fatalf("N = %d, want valid sample count 3", got.N)
	}
	if len(got.Buckets) != 10 {
		t.Fatalf("buckets = %d, want 10", len(got.Buckets))
	}
	if got.Buckets[0].Label != "0.0-1.0 bpm" || got.Buckets[9].Label != ">= 9.0 bpm" {
		t.Fatalf("labels = %q / %q", got.Buckets[0].Label, got.Buckets[9].Label)
	}
	if got.Buckets[9].Seconds != 1 || got.Buckets[9].Percentage != 33.3 {
		t.Fatalf("final bucket = %#v, want max value included with rounded percent", got.Buckets[9])
	}
	if got.FixedWidth == nil || got.FixedWidth.BucketCount != 10 || got.FixedWidth.Width != 1 {
		t.Fatalf("FixedWidth = %#v", got.FixedWidth)
	}
}

func TestBuildHistogramIdenticalValuesUsesOneFixedBucket(t *testing.T) {
	t.Parallel()

	got := BuildHistogram([]HistogramSample{{Value: 250, Seconds: 2}, {Value: 250, Seconds: 3}}, "W", nil)
	if len(got.Buckets) != 1 {
		t.Fatalf("buckets = %d, want 1", len(got.Buckets))
	}
	assertBucket(t, got.Buckets[0], "250.0 W", ptr(250), nil, 5, 100)
	if got.FixedWidth == nil || got.FixedWidth.BucketCount != 1 || got.FixedWidth.Width != 0 {
		t.Fatalf("FixedWidth = %#v", got.FixedWidth)
	}
}

func TestPaceZoneBoundaryConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value float64
		unit  string
		emit  string
		want  float64
		ok    bool
	}{
		{name: "mins km stored as seconds per km to mile", value: 300, unit: "MINS_KM", emit: "seconds_per_mile", want: 482.8032, ok: true},
		{name: "mins mile stored as seconds per mile to km", value: 480, unit: "MINS_MILE", emit: "seconds_per_km", want: 298.2581722739203, ok: true},
		{name: "secs 100m to km", value: 35, unit: "SECS_100M", emit: "seconds_per_km", want: 350, ok: true},
		{name: "secs 500m to mile", value: 120, unit: "SECS_500M", emit: "seconds_per_mile", want: 386.24256, ok: true},
		{name: "unknown", value: 1, unit: "", emit: "seconds_per_km", ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ConvertPaceZoneBoundary(tc.value, tc.unit, tc.emit)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && !near(got, tc.want) {
				t.Fatalf("value = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHistogramMetricSchemaAndAliases(t *testing.T) {
	t.Parallel()

	schema := HistogramMetricSchemaProperty()
	enum := schema["enum"].([]string)
	if len(enum) != 3 {
		t.Fatalf("enum = %#v, want 3 histogram metrics", enum)
	}
	for _, input := range []string{"power", "watts", "hr", "heart_rate", "pace"} {
		if _, err := ParseHistogramMetric(input); err != nil {
			t.Fatalf("ParseHistogramMetric(%q) error = %v", input, err)
		}
	}
	if _, err := ParseHistogramMetric("ctl"); err == nil {
		t.Fatal("ParseHistogramMetric(ctl) error = nil, want rejection")
	}
}

func assertBucket(t *testing.T, got HistogramBucket, label string, lower *float64, upper *float64, seconds float64, percentage float64) {
	t.Helper()
	if got.Label != label || !ptrEqual(got.Lower, lower) || !ptrEqual(got.Upper, upper) || got.Seconds != seconds || got.Percentage != percentage {
		t.Fatalf("bucket = %#v, want label=%q lower=%v upper=%v seconds=%v percentage=%v", got, label, lower, upper, seconds, percentage)
	}
}

func ptr(v float64) *float64 { return &v }

func ptrEqual(a *float64, b *float64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func near(a float64, b float64) bool {
	if a > b {
		return a-b < 0.000001
	}
	return b-a < 0.000001
}

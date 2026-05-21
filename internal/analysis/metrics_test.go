package analysis

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseMetricValidCanonicalAndAliases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  Metric
	}{
		{name: "canonical ctl", input: "ctl", want: "ctl"},
		{name: "canonical per km pace", input: "pace_seconds_per_km", want: "pace_seconds_per_km"},
		{name: "canonical per mile pace", input: "pace_seconds_per_mile", want: "pace_seconds_per_mile"},
		{name: "resting hr alias", input: "resting_hr", want: "rhr"},
		{name: "camel resting hr alias", input: "restingHR", want: "rhr"},
		{name: "sleep secs alias", input: "sleepSecs", want: "sleep_secs"},
		{name: "sleep score alias", input: "sleepScore", want: "sleep_score"},
		{name: "sleep quality alias", input: "sleepQuality", want: "sleep_quality"},
		{name: "intensity factor alias", input: "intensity_factor", want: "if"},
		{name: "variability index alias", input: "variability_index", want: "vi"},
		{name: "compliance percent alias", input: "compliance_percent", want: "compliance_pct"},
		{name: "ramp rate alias", input: "rampRate", want: "ramp"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseMetric(tc.input)
			if err != nil {
				t.Fatalf("ParseMetric(%q) error = %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("ParseMetric(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestMetricValuesRoundTripThroughParser(t *testing.T) {
	t.Parallel()

	for _, value := range MetricValues() {
		t.Run(value, func(t *testing.T) {
			got, err := ParseMetric(value)
			if err != nil {
				t.Fatalf("ParseMetric(%q) error = %v", value, err)
			}
			if string(got) != value {
				t.Fatalf("ParseMetric(%q) = %q", value, got)
			}
		})
	}
}

func TestParseMetricUnknownHints(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		contains string
	}{
		{name: "best effort duration", input: "5min_power", contains: "try analyze_efforts_delta"},
		{name: "best effort distance", input: "5k_pace", contains: "try analyze_efforts_delta"},
		{name: "zone distribution", input: "power_zone_distribution_seconds", contains: "try compute_zone_time or compute_load_balance"},
		{name: "load balance", input: "load_balance", contains: "try compute_zone_time or compute_load_balance"},
		{name: "segment stat", input: "mean_power_segment", contains: "try compute_activity_segment_stats"},
		{name: "stream stat", input: "hr_drift_stream", contains: "try compute_activity_segment_stats"},
		{name: "compliance", input: "completed_vs_scheduled", contains: "try compute_compliance_rate"},
		{name: "generic ftp", input: "ftp", contains: "use one of: ctl, atl, tsb"},
		{name: "generic distance", input: "distance", contains: "use one of: ctl, atl, tsb"},
		{name: "generic tss", input: "tss", contains: "use one of: ctl, atl, tsb"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseMetric(tc.input)
			assertInvalidMetricError(t, err, tc.contains)
		})
	}
}

func TestParseMetricRejectsExpressions(t *testing.T) {
	t.Parallel()

	cases := []string{
		"ctl/atl",
		"ctl - atl",
		"weekly_tss/weekly_hours",
		"(ctl+atl)/2",
		"ctl,atl",
		"ctl|atl",
		"tss_per_hour",
		"power:weight",
		"np/ftp",
	}

	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			_, err := ParseMetric(input)
			assertInvalidMetricError(t, err, "expressions are not supported")
		})
	}
}

func TestMetricSchemaProperty(t *testing.T) {
	t.Parallel()

	schema := MetricSchemaProperty()
	if got := schema["type"]; got != "string" {
		t.Fatalf("schema type = %v, want string", got)
	}
	enumValues, ok := schema["enum"].([]string)
	if !ok {
		t.Fatalf("schema enum type = %T, want []string", schema["enum"])
	}
	if !reflect.DeepEqual(enumValues, MetricValues()) {
		t.Fatalf("schema enum = %v, want MetricValues()", enumValues)
	}
	for _, alias := range []string{"resting_hr", "sleepSecs", "intensity_factor", "compliance_percent"} {
		if containsString(enumValues, alias) {
			t.Fatalf("schema enum includes alias %q", alias)
		}
	}
	description, ok := schema["description"].(string)
	if !ok {
		t.Fatalf("schema description type = %T, want string", schema["description"])
	}
	for _, want := range []string{"aliases", "Expressions", "ctl"} {
		if !strings.Contains(description, want) {
			t.Fatalf("schema description %q does not contain %q", description, want)
		}
	}
	for _, forbidden := range []string{"MetricSource", "SourceFamily", "get_wellness_data", "internal/analysis"} {
		if strings.Contains(description, forbidden) {
			t.Fatalf("schema description exposes internal detail %q: %q", forbidden, description)
		}
	}
	if len(description) > 260 {
		t.Fatalf("schema description length = %d, want concise", len(description))
	}
}

func TestMetricMetadataHelpers(t *testing.T) {
	t.Parallel()

	if err := ValidateMetricCatalog(); err != nil {
		t.Fatalf("ValidateMetricCatalog() error = %v", err)
	}
	for _, value := range MetricValues() {
		t.Run("sources_"+value, func(t *testing.T) {
			if sources := MetricSources(Metric(value)); len(sources) == 0 {
				t.Fatalf("MetricSources(%q) is empty", value)
			}
		})
	}

	assertSource(t, "ctl", SourceFitnessDaily, GrainDaily, KindScalar)
	assertSource(t, "ctl", SourceWellnessDaily, GrainDaily, KindScalar)
	assertSource(t, "training_load", SourceActivityRow, GrainActivity, KindScalar)
	assertSource(t, "training_load", SourceTrainingSummary, GrainSummaryWindow, KindScalar)
	assertSource(t, "training_load", SourceExtendedActivity, GrainActivity, KindScalar)
	assertSource(t, "training_load", SourceExtendedInterval, GrainInterval, KindScalar)
	assertSource(t, "feel", SourceWellnessDaily, GrainDaily, KindSubjectiveScale)
	assertSource(t, "feel", SourceExtendedActivity, GrainActivity, KindSubjectiveScale)

	for _, metric := range []Metric{"weekly_tss", "weekly_hours"} {
		sources := MetricSources(metric)
		if len(sources) != 1 {
			t.Fatalf("MetricSources(%q) length = %d, want 1", metric, len(sources))
		}
		source := sources[0]
		if source.Kind != KindDerived || source.Family != SourceDerivedWeekly || source.UnitLabel == "" || source.Method == "" {
			t.Fatalf("derived source for %q = %+v", metric, source)
		}
	}

	for _, metric := range []Metric{"sleep_quality", "feel", "fatigue", "mood"} {
		if !hasScaleMetadata(metric) {
			t.Fatalf("%q missing subjective scale metadata", metric)
		}
	}

	sources := MetricSources("ctl")
	if len(sources) == 0 {
		t.Fatal("MetricSources(ctl) returned no sources")
	}
	sources[0].Field = "mutated"
	if got := MetricSources("ctl")[0].Field; got == "mutated" {
		t.Fatal("MetricSources returned a mutable backing slice")
	}
}

func TestInvalidMetricErrorContract(t *testing.T) {
	t.Parallel()

	if _, err := ParseMetric("ctl"); err != nil {
		t.Fatalf("ParseMetric(ctl) error = %v", err)
	}
	_, err := ParseMetric("ctl/atl")
	assertInvalidMetricError(t, err, "expressions are not supported")
	if !IsInvalidMetric(fmt.Errorf("wrapping: %w", err)) {
		t.Fatalf("IsInvalidMetric(wrapped err) = false")
	}
	message := err.Error()
	for _, forbidden := range []string{"analysis.InvalidMetricError", "internal/analysis", "MetricSource", "goroutine", "panic"} {
		if strings.Contains(message, forbidden) {
			t.Fatalf("error message exposes internal detail %q: %q", forbidden, message)
		}
	}
	if len(message) > 120 {
		t.Fatalf("error message length = %d, want concise: %q", len(message), message)
	}
}

func assertInvalidMetricError(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatal("error = nil, want invalid metric error")
	}
	if !IsInvalidMetric(err) {
		t.Fatalf("IsInvalidMetric(%T) = false", err)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("error = %q, want to contain %q", err.Error(), contains)
	}
}

func assertSource(t *testing.T, metric Metric, family SourceFamily, grain Grain, kind MetricKind) {
	t.Helper()
	for _, source := range MetricSources(metric) {
		if source.Family == family && source.Grain == grain && source.Kind == kind {
			return
		}
	}
	t.Fatalf("MetricSources(%q) missing family=%q grain=%q kind=%q; got %+v", metric, family, grain, kind, MetricSources(metric))
}

func hasScaleMetadata(metric Metric) bool {
	for _, source := range MetricSources(metric) {
		if source.Kind == KindSubjectiveScale && source.ScaleLabel != "" {
			return true
		}
	}
	return false
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

package tools

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/response"
)

// TestMetricCatalogSourcesResolveToWorkingExtractors asserts every advertised
// analysis_metric source resolves to a field the extraction layer can read, so
// a metric cannot enter the public enum without a working upstream mapping.
func TestMetricCatalogSourcesResolveToWorkingExtractors(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"a1","type":"Ride","start_date_local":"2026-05-01T07:00:00","distance":10000,"icu_distance":10000,"moving_time":1800,"elapsed_time":1900,"average_speed":5.5,"max_speed":9.1,"total_elevation_gain":120,"total_elevation_loss":118,"icu_training_load":75,"average_heartrate":141,"max_heartrate":171,"average_cadence":85,"icu_average_watts":166,"calories":600}`)
	summary := decodeSummaries(t, `[{"date":"2026-05-01","time":3600,"moving_time":3400,"elapsed_time":3700,"calories":600,"total_elevation_gain":120,"training_load":75,"srpe":5,"distance":10000,"fitness":60,"fatigue":55,"form":5,"timeInZones":[100,200],"timeInZonesTot":300}]`)[0]
	extendedActivityFields := jsonFieldSet(reflect.TypeOf(extendedActivityMetrics{}))
	extendedIntervalFields := jsonFieldSet(reflect.TypeOf(extendedIntervalMetrics{}))
	activityIntervalFields := jsonFieldSet(reflect.TypeOf(activityIntervalRow{}))

	for _, name := range analysis.MetricValues() {
		metric := analysis.Metric(name)
		sources := analysis.MetricSources(metric)
		if len(sources) == 0 {
			t.Fatalf("metric %s has no sources", name)
		}
		for _, source := range sources {
			t.Run(fmt.Sprintf("%s/%s", name, source.Family), func(t *testing.T) {
				switch source.Family {
				case analysis.SourceWellnessDaily:
					row := decodeWellnessRow(t, fmt.Sprintf(`{"id":"2026-05-01","%s":2}`, source.Field))
					if _, ok := wellnessMetricValue(row, metric); !ok {
						t.Fatalf("wellnessMetricValue cannot read %s via upstream field %q; catalog field and Wellness struct tag must match", name, source.Field)
					}
				case analysis.SourceActivityRow:
					if _, ok := activityMetricValue(activity, metric, response.UnitSystemMetric); !ok {
						t.Fatalf("activityMetricValue cannot read %s from a fully populated activity row; add an extractor case and upstream field", name)
					}
				case analysis.SourceFitnessDaily, analysis.SourceTrainingSummary, analysis.SourceDerivedWeekly:
					if _, ok := summaryMetricValue(summary, metric, response.UnitSystemMetric); !ok {
						t.Fatalf("summaryMetricValue cannot read %s from a fully populated summary row", name)
					}
				case analysis.SourceExtendedActivity:
					if !extendedActivityFields[source.Field] {
						t.Fatalf("extended activity field %q for %s is not an extendedActivityMetrics output field; compute_baseline resolves it by that json tag", source.Field, name)
					}
				case analysis.SourceExtendedInterval:
					if !extendedIntervalFields[source.Field] {
						t.Fatalf("extended interval field %q for %s is not an extendedIntervalMetrics output field", source.Field, name)
					}
				case analysis.SourceActivityInterval:
					if !activityIntervalFields[source.Field] {
						t.Fatalf("activity interval field %q for %s is not an activityIntervalRow output field", source.Field, name)
					}
				default:
					t.Fatalf("metric %s declares unknown source family %q", name, source.Family)
				}
			})
		}
	}
}

// TestMetricCatalogAnalyzerReachability pins which enum metrics no analyzer can
// serve at daily or activity grain. Interval-only metrics stay in the enum for
// schema stability and fail with a specific error naming the interval tools;
// adding a new one must be a conscious decision recorded here.
func TestMetricCatalogAnalyzerReachability(t *testing.T) {
	t.Parallel()

	intervalOnly := map[string]bool{
		"dfa_alpha1":               true,
		"distance_m":               true,
		"duration_seconds":         true,
		"w_prime_balance_end_kj":   true,
		"w_prime_balance_start_kj": true,
	}
	for _, name := range analysis.MetricValues() {
		reachable := false
		for _, source := range analysis.MetricSources(analysis.Metric(name)) {
			switch source.Family {
			case analysis.SourceFitnessDaily, analysis.SourceWellnessDaily, analysis.SourceTrainingSummary, analysis.SourceActivityRow, analysis.SourceDerivedWeekly, analysis.SourceExtendedActivity:
				reachable = true
			}
		}
		if !reachable && !intervalOnly[name] {
			t.Errorf("metric %s is unreachable by every analyzer; give it a daily/activity source or record it as interval-only", name)
		}
		if reachable && intervalOnly[name] {
			t.Errorf("metric %s is now analyzer-reachable; remove it from the interval-only list", name)
		}
	}
}

func jsonFieldSet(structType reflect.Type) map[string]bool {
	fields := map[string]bool{}
	for i := 0; i < structType.NumField(); i++ {
		tag := structType.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name != "" {
			fields[name] = true
		}
	}
	return fields
}

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	maxAnalyzerWindowDays = 366
)

type analyzerWindowRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (w analyzerWindowRequest) analysisWindow() analysis.Window {
	return analysis.Window{StartDate: strings.TrimSpace(w.StartDate), EndDate: strings.TrimSpace(w.EndDate)}
}

type analyzerClients struct {
	fitness    FitnessClient
	wellness   WellnessClient
	activities ActivitiesClient
	efforts    BestEffortsClient
}

func decodeAnalyzerStrict[T any](raw json.RawMessage) (T, error) {
	var zero T
	if strings.TrimSpace(string(raw)) == "" {
		return zero, errors.New("arguments must be a JSON object")
	}
	return DecodeStrict[T](raw)
}

func loadAnalyzerSeries(ctx context.Context, clients analyzerClients, metric analysis.Metric, window analysis.ParsedWindow, grain analysis.SampleGrain, sport string, unitSystem response.UnitSystem, allowWeekly bool) (analyzerSampleSeries, error) {
	selection, err := selectAnalyzerMetricSource(metric, grain, allowWeekly)
	if err != nil {
		return analyzerSampleSeries{}, err
	}
	series := analyzerSampleSeries{Metric: metric, Unit: selection.Source.UnitLabel, ScaleLabel: selection.Source.ScaleLabel, SourceTools: []string{selection.Source.Tool}, Assumptions: map[string]any{"sample_grain": string(grain), "unit": selection.Source.UnitLabel}}
	if selection.Source.ScaleLabel != "" {
		series.Assumptions["scale_label"] = selection.Source.ScaleLabel
	}
	switch selection.Source.Family {
	case analysis.SourceFitnessDaily, analysis.SourceTrainingSummary, analysis.SourceDerivedWeekly:
		if clients.fitness == nil {
			return series, errors.New("missing fitness client")
		}
		rows, err := clients.fitness.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: window.StartDate, End: window.EndDate})
		if err != nil {
			return series, err
		}
		if metric == "weekly_tss" || metric == "weekly_hours" {
			series.Samples, series.MissingDays = weeklySummarySamples(rows, metric, window, unitSystem)
			series.Assumptions["sample_grain"] = string(analysis.SampleGrainWeekly)
			series.Assumptions["aggregation"] = map[bool]string{true: "weekly_hours", false: "weekly_sum"}[metric == "weekly_hours"]
			series.Assumptions["expected_weekly_buckets"] = expectedWeeklyBuckets(window.Days)
			return series, nil
		}
		seen := map[string]bool{}
		for _, row := range rows {
			if value, ok := summaryMetricValue(row, metric, unitSystem); ok {
				series.Samples = append(series.Samples, analysis.NumericSample{Key: row.Date, Date: row.Date, Value: value})
				seen[row.Date] = true
			}
		}
		series.Samples = sortedSamples(series.Samples)
		series.MissingDays = analysis.MissingSamples(window.Days, len(seen))
		series.Assumptions["aggregation"] = "native_daily"
	case analysis.SourceWellnessDaily:
		if clients.wellness == nil {
			return series, errors.New("missing wellness client")
		}
		rows, err := clients.wellness.ListWellness(ctx, intervals.WellnessParams{Oldest: window.StartDate, Newest: window.EndDate, Fields: []string{selection.Source.Field}})
		if err != nil {
			return series, err
		}
		seen := map[string]bool{}
		for _, row := range rows {
			date := strings.TrimSpace(stringValue(row.ID))
			if value, ok := wellnessMetricValue(row, metric); ok && date != "" {
				series.Samples = append(series.Samples, analysis.NumericSample{Key: date, Date: date, Value: value})
				seen[date] = true
			}
		}
		series.Samples = sortedSamples(series.Samples)
		series.MissingDays = analysis.MissingSamples(window.Days, len(seen))
		series.Assumptions["aggregation"] = "native_daily"
	case analysis.SourceActivityRow:
		if clients.activities == nil {
			return series, errors.New("missing activities client")
		}
		rows, err := loadAllAnalyzerActivities(ctx, clients.activities, window.StartDate, window.EndDate)
		if err != nil {
			return series, err
		}
		rows = filterAnalyzerSport(rows, sport)
		if grain == analysis.SampleGrainActivity {
			for _, row := range rows {
				if value, ok := activityMetricValue(row, metric, unitSystem); ok {
					date := localActivityDate(row)
					series.Samples = append(series.Samples, analysis.NumericSample{Key: row.ID, Date: date, ActivityID: row.ID, Value: value})
				}
			}
			series.MissingDays = 0
			series.Assumptions["missing_days_applicable"] = false
			series.Assumptions["sample_grain"] = string(analysis.SampleGrainActivity)
			series.Samples = sortedSamples(series.Samples)
			return series, nil
		}
		byDate := groupActivitiesByLocalDate(rows)
		for date, dateRows := range byDate {
			if sample, ok, assumptions := aggregateActivityDay(date, dateRows, metric, unitSystem); ok {
				series.Samples = append(series.Samples, sample)
				for key, value := range assumptions {
					series.Assumptions[key] = value
				}
			}
		}
		series.Samples = sortedSamples(series.Samples)
		series.MissingDays = analysis.MissingSamples(window.Days, len(series.Samples))
	}
	return series, nil
}

func filterAnalyzerSport(rows []intervals.Activity, sport string) []intervals.Activity {
	trimmed := strings.ToLower(strings.TrimSpace(sport))
	if trimmed == "" {
		return rows
	}
	out := make([]intervals.Activity, 0, len(rows))
	for _, row := range rows {
		if strings.ToLower(strings.TrimSpace(stringValue(row.Type))) == trimmed || strings.ToLower(strings.TrimSpace(stringValue(row.SubType))) == trimmed {
			out = append(out, row)
		}
	}
	return out
}

func sampleMapByDate(samples []analysis.NumericSample) map[string]analysis.NumericSample {
	out := map[string]analysis.NumericSample{}
	for _, sample := range samples {
		out[sample.Date] = sample
	}
	return out
}

func pairDailySamples(xSeries analyzerSampleSeries, ySeries analyzerSampleSeries, window analysis.ParsedWindow, lagDays int) []analysis.PairedSample {
	yByDate := sampleMapByDate(ySeries.Samples)
	pairs := []analysis.PairedSample{}
	for _, x := range xSeries.Samples {
		date, err := time.Parse(time.DateOnly, x.Date)
		if err != nil || date.Before(window.Start) || date.After(window.End) {
			continue
		}
		yDate := date.AddDate(0, 0, lagDays).Format(time.DateOnly)
		if y, ok := yByDate[yDate]; ok {
			pairs = append(pairs, analysis.PairedSample{Key: x.Date, Date: x.Date, X: x.Value, Y: y.Value})
		}
	}
	return pairs
}

func pairActivitySamples(xSeries analyzerSampleSeries, ySeries analyzerSampleSeries) []analysis.PairedSample {
	yByID := map[string]analysis.NumericSample{}
	for _, y := range ySeries.Samples {
		yByID[y.ActivityID] = y
	}
	pairs := []analysis.PairedSample{}
	for _, x := range xSeries.Samples {
		if y, ok := yByID[x.ActivityID]; ok {
			pairs = append(pairs, analysis.PairedSample{Key: x.ActivityID, Date: x.Date, X: x.Value, Y: y.Value})
		}
	}
	return pairs
}

func shiftedLookupWindow(window analysis.ParsedWindow, lagDays int) analysis.Window {
	if lagDays > 0 {
		return analysis.Window{StartDate: window.StartDate, EndDate: window.End.AddDate(0, 0, lagDays).Format(time.DateOnly)}
	}
	if lagDays < 0 {
		return analysis.Window{StartDate: window.Start.AddDate(0, 0, lagDays).Format(time.DateOnly), EndDate: window.EndDate}
	}
	return window.Window
}

func weeklySummarySamples(rows []intervals.SummaryWithCats, metric analysis.Metric, window analysis.ParsedWindow, unitSystem response.UnitSystem) ([]analysis.NumericSample, int) {
	byDate := map[string]intervals.SummaryWithCats{}
	for _, row := range rows {
		byDate[row.Date] = row
	}
	samples := []analysis.NumericSample{}
	missingDays := 0
	for bucketIndex, start := 0, window.Start; !start.After(window.End); bucketIndex, start = bucketIndex+1, start.AddDate(0, 0, 7) {
		end := start.AddDate(0, 0, 6)
		if end.After(window.End) {
			end = window.End
		}
		var sum float64
		var seen bool
		for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
			row, ok := byDate[day.Format(time.DateOnly)]
			if !ok {
				missingDays++
				continue
			}
			if value, ok := summaryMetricValue(row, metric, unitSystem); ok {
				sum += value
				seen = true
			}
		}
		if seen {
			key := start.Format(time.DateOnly) + "/" + end.Format(time.DateOnly)
			samples = append(samples, analysis.NumericSample{Key: key, Date: start.Format(time.DateOnly), Bucket: bucketIndex, Value: sum})
		}
	}
	return samples, missingDays
}

func expectedWeeklyBuckets(days int) int {
	if days <= 0 {
		return 0
	}
	return (days + 6) / 7
}

func analyzerMetaAssumptions(base map[string]any, window analysis.Window, includeFull bool) map[string]any {
	out := map[string]any{"window": window, "include_full": includeFull}
	for key, value := range base {
		out[key] = value
	}
	return out
}

func mergeSourceTools(series ...analyzerSampleSeries) []string {
	var tools []string
	for _, item := range series {
		tools = append(tools, item.SourceTools...)
	}
	return analysis.NormalizeSourceTools(tools)
}

func parseMetricArgument(value string) (analysis.Metric, error) {
	metric, err := analysis.ParseMetric(value)
	if err != nil {
		return "", err
	}
	return metric, nil
}

func analyzerMetricProperty(description string) map[string]any {
	property := analysis.MetricSchemaProperty()
	if description != "" {
		property["description"] = description + " " + fmt.Sprint(property["description"])
	}
	return property
}

func sortedFloatCopy(values []float64) []float64 {
	out := append([]float64(nil), values...)
	sort.Float64s(out)
	return out
}

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	computeTrainingMonotonyName        = "compute_training_monotony"
	computeTrainingMonotonyDescription = "Use when the prompt asks for transparent daily training-load monotony across an inclusive athlete-local date window; returns the Foster-style descriptive statistic only when every expected get_training_summary date has exactly one valid raw training_load number. It is not a readiness, risk, medical, adaptation, or sport-variety score."
	maxTrainingMonotonyDays            = 31
)

// TrainingMonotonyClient is the raw daily-summary source required by the monotony analyzer.
type TrainingMonotonyClient interface {
	ListAthleteSummaryRaw(context.Context, intervals.AthleteSummaryParams) ([]intervals.RawSummaryRow, error)
}

type trainingMonotonyRequest struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type trainingMonotonyResult struct {
	Status            string   `json:"status"`
	Reason            string   `json:"reason,omitempty"`
	StartDate         string   `json:"start_date"`
	EndDate           string   `json:"end_date"`
	MeanDailyLoad     *float64 `json:"mean_daily_load,omitempty"`
	StandardDeviation *float64 `json:"standard_deviation,omitempty"`
	Monotony          *float64 `json:"monotony,omitempty"`
}

type trainingMonotonyCoverage struct {
	SourceTool        string   `json:"source_tool"`
	SourceField       string   `json:"source_field"`
	DateBasis         string   `json:"date_basis"`
	ExpectedDays      int      `json:"expected_days"`
	ExpectedDates     []string `json:"expected_dates"`
	ReceivedRows      int      `json:"received_rows"`
	UniqueDates       int      `json:"unique_dates"`
	ValidRows         int      `json:"valid_rows"`
	ValidDates        []string `json:"valid_dates"`
	MissingDates      []string `json:"missing_dates"`
	DuplicateDates    []string `json:"duplicate_dates"`
	InvalidDates      []string `json:"invalid_dates"`
	OutOfWindowDates  []string `json:"out_of_window_dates"`
	MalformedLoadRows []string `json:"malformed_load_rows"`
	NegativeLoadRows  []string `json:"negative_load_rows"`
	NonObjectRows     []string `json:"non_object_rows"`
	DecodeErrorRows   []string `json:"decode_error_rows"`
	RejectedRows      []string `json:"rejected_rows"`
}

type trainingMonotonyDailyEvidence struct {
	Date         string  `json:"date"`
	TrainingLoad float64 `json:"training_load"`
}

func newComputeTrainingMonotonyTool(client TrainingMonotonyClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{
		Name:         computeTrainingMonotonyName,
		Description:  computeTrainingMonotonyDescription,
		InputSchema:  trainingMonotonyInputSchema(),
		OutputSchema: genericOutputSchema("Foster-style daily training-load monotony with explicit source and coverage evidence."),
		Handler:      computeTrainingMonotonyHandler(client, version, debugMetadata, shapeCfg),
	})
}

func computeTrainingMonotonyHandler(client TrainingMonotonyClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, start, end, err := decodeTrainingMonotonyRequest(req.Arguments)
		if err != nil {
			return Result{}, err
		}
		if client == nil {
			return Result{}, NewUserError("unavailable", errors.New("training summary source is not configured"))
		}
		rows, err := client.ListAthleteSummaryRaw(ctx, intervals.AthleteSummaryParams{Start: args.StartDate, End: args.EndDate})
		if err != nil {
			return Result{}, NewUserError("could not compute training monotony; check intervals.icu credentials, athlete ID, and date range", err)
		}
		coverage, loads, daily := inspectTrainingMonotonyRows(rows, start, end)
		meta := monotonyMeta(coverage)
		result := trainingMonotonyResult{Status: "unavailable", StartDate: args.StartDate, EndDate: args.EndDate}
		if coverage.ExpectedDays == 1 {
			result.Reason = "insufficient_days"
			meta.N = 0
			meta.MissingDays = coverage.ExpectedDays
			meta.InsufficientSample = monotonyBoolPointer(true)
		} else if len(rows) == 0 {
			result.Reason = "no_daily_rows"
			meta.N = 0
			meta.MissingDays = coverage.ExpectedDays
			meta.InsufficientSample = monotonyBoolPointer(true)
		} else if reason := coverageDefectReason(coverage); reason != "" {
			result.Reason = reason
			meta.N = 0
			meta.MissingDays = coverage.ExpectedDays
			meta.InsufficientSample = monotonyBoolPointer(true)
		} else {
			computed, computeErr := analysis.ComputeTrainingLoadMonotony(loads)
			if computeErr != nil {
				return Result{}, fmt.Errorf("computing training monotony: %w", computeErr)
			}
			meta.N = coverage.ExpectedDays
			meta.MissingDays = 0
			meta.InsufficientSample = monotonyBoolPointer(false)
			if computed.ZeroVariance {
				result.Status = "undefined"
				result.Reason = "zero_variance"
			} else {
				result.Status = "ok"
				mean := roundMonotony(computed.Mean)
				sd := roundMonotony(computed.PopulationStandardDev)
				monotony := roundMonotony(computed.Monotony)
				result.MeanDailyLoad = &mean
				result.StandardDeviation = &sd
				result.Monotony = &monotony
			}
		}
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: daily, Meta: meta}, args.IncludeFull, version, debugMetadata, computeTrainingMonotonyName, response.UnitSystemMetric, shapeCfg)
	}
}

func decodeTrainingMonotonyRequest(raw json.RawMessage) (trainingMonotonyRequest, time.Time, time.Time, error) {
	var args trainingMonotonyRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, time.Time{}, time.Time{}, NewUserError("invalid_date_range", errors.New("arguments must be a JSON object"))
	}
	decoded, err := DecodeStrict[trainingMonotonyRequest](raw)
	if err != nil {
		return args, time.Time{}, time.Time{}, NewUserError("invalid_date_range", err)
	}
	args = decoded
	start, startErr := time.Parse(time.DateOnly, args.StartDate)
	end, endErr := time.Parse(time.DateOnly, args.EndDate)
	if startErr != nil || endErr != nil || end.Before(start) {
		return args, time.Time{}, time.Time{}, NewUserError("invalid_date_range", errors.New("dates must be valid YYYY-MM-DD with end on or after start"))
	}
	if int(end.Sub(start)/(24*time.Hour))+1 > maxTrainingMonotonyDays {
		return args, time.Time{}, time.Time{}, NewUserError("date_range_too_large", errors.New("date range exceeds 31 expected dates"))
	}
	return args, start, end, nil
}

func inspectTrainingMonotonyRows(rows []intervals.RawSummaryRow, start, end time.Time) (trainingMonotonyCoverage, []float64, []trainingMonotonyDailyEvidence) {
	expected := expectedDateStrings(start, end)
	coverage := trainingMonotonyCoverage{
		SourceTool:    "get_training_summary",
		SourceField:   "training_load",
		DateBasis:     "athlete_local_summary_date",
		ExpectedDays:  len(expected),
		ExpectedDates: expected,
		ReceivedRows:  len(rows),
		ValidDates:    []string{}, MissingDates: []string{}, DuplicateDates: []string{}, InvalidDates: []string{},
		OutOfWindowDates: []string{}, MalformedLoadRows: []string{}, NegativeLoadRows: []string{}, NonObjectRows: []string{},
		DecodeErrorRows: []string{}, RejectedRows: []string{},
	}
	dateCounts := map[string]int{}
	dateRowIndices := map[string][]int{}
	validLoadByDate := map[string]float64{}
	validDateSet := map[string]bool{}
	rowReasons := make([][]string, len(rows))
	for index, row := range rows {
		if row.Raw == nil {
			token := rawRowToken(index, row.RawJSON)
			coverage.NonObjectRows = append(coverage.NonObjectRows, token)
			rowReasons[index] = append(rowReasons[index], "non_object_daily_row")
			continue
		}
		dateRaw, loadRaw := rawFields(row.RawJSON)
		date, dateValid := parseSummaryDate(dateRaw)
		if !dateValid {
			coverage.InvalidDates = append(coverage.InvalidDates, rawFieldToken(index, dateRaw))
			rowReasons[index] = append(rowReasons[index], "invalid_daily_date")
		} else {
			dateCounts[date]++
			if date >= expected[0] && date <= expected[len(expected)-1] {
				dateRowIndices[date] = append(dateRowIndices[date], index)
			}
			if date < expected[0] || date > expected[len(expected)-1] {
				coverage.OutOfWindowDates = append(coverage.OutOfWindowDates, rawFieldToken(index, dateRaw))
				rowReasons[index] = append(rowReasons[index], "out_of_window_daily_row")
			}
		}
		load, loadState := parseSummaryLoad(loadRaw)
		switch loadState {
		case "malformed":
			coverage.MalformedLoadRows = append(coverage.MalformedLoadRows, rawFieldToken(index, loadRaw))
			rowReasons[index] = append(rowReasons[index], "malformed_daily_load")
		case "negative":
			coverage.NegativeLoadRows = append(coverage.NegativeLoadRows, rawFieldToken(index, loadRaw))
			rowReasons[index] = append(rowReasons[index], "negative_daily_load")
		}
		if dateValid && date >= expected[0] && date <= expected[len(expected)-1] && loadState == "valid" {
			coverage.ValidRows++
			validDateSet[date] = true
			validLoadByDate[date] = load
		}
	}
	for date, count := range dateCounts {
		if date >= expected[0] && date <= expected[len(expected)-1] && count > 1 {
			coverage.DuplicateDates = append(coverage.DuplicateDates, date)
			for _, index := range dateRowIndices[date] {
				rowReasons[index] = append(rowReasons[index], "duplicate_daily_date")
			}
		}
	}
	for _, date := range expected {
		if !validDateSet[date] {
			coverage.MissingDates = append(coverage.MissingDates, date)
		} else {
			coverage.ValidDates = append(coverage.ValidDates, date)
		}
	}
	for index, reasons := range rowReasons {
		if len(reasons) == 0 {
			continue
		}
		coverage.RejectedRows = append(coverage.RejectedRows, fmt.Sprintf("row:%d", index))
	}
	coverage.UniqueDates = len(dateCounts)
	sort.Strings(coverage.ValidDates)
	sort.Strings(coverage.DuplicateDates)
	sort.Strings(coverage.MissingDates)
	sort.Strings(coverage.InvalidDates)
	sort.Strings(coverage.OutOfWindowDates)
	sort.Strings(coverage.MalformedLoadRows)
	sort.Strings(coverage.NegativeLoadRows)
	sort.Strings(coverage.NonObjectRows)
	sort.Strings(coverage.DecodeErrorRows)
	sort.Strings(coverage.RejectedRows)
	loads := make([]float64, 0, len(expected))
	daily := make([]trainingMonotonyDailyEvidence, 0, len(expected))
	for _, date := range expected {
		if load, ok := validLoadByDate[date]; ok {
			loads = append(loads, load)
			daily = append(daily, trainingMonotonyDailyEvidence{Date: date, TrainingLoad: load})
		}
	}
	return coverage, loads, daily
}

func rawFields(raw json.RawMessage) (json.RawMessage, json.RawMessage) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return json.RawMessage("null"), json.RawMessage("null")
	}
	date := fields["date"]
	if len(date) == 0 {
		date = json.RawMessage("null")
	}
	load := fields["training_load"]
	if len(load) == 0 {
		load = json.RawMessage("null")
	}
	return date, load
}

func parseSummaryDate(raw json.RawMessage) (string, bool) {
	var date string
	if len(raw) == 0 || json.Unmarshal(raw, &date) != nil || len(date) != len("2006-01-02") || string(raw) != strconv.Quote(date) {
		return "", false
	}
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return "", false
	}
	return date, true
}

func parseSummaryLoad(raw json.RawMessage) (float64, string) {
	text := string(raw)
	value, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, "malformed"
	}
	rational, ok := new(big.Rat).SetString(text)
	if !ok || (value == 0 && rational.Sign() != 0) {
		return 0, "malformed"
	}
	if value < 0 {
		return value, "negative"
	}
	return value, "valid"
}

func rawFieldToken(index int, raw json.RawMessage) string {
	return fmt.Sprintf("row:%d:%s", index, compactRaw(raw))
}

func rawRowToken(index int, raw json.RawMessage) string {
	return fmt.Sprintf("row:%d:%s", index, compactRaw(raw))
}

func compactRaw(raw json.RawMessage) string {
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "null"
	}
	return compact.String()
}

func expectedDateStrings(start, end time.Time) []string {
	dates := make([]string, 0, int(end.Sub(start)/(24*time.Hour))+1)
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		dates = append(dates, date.Format(time.DateOnly))
	}
	return dates
}

func coverageDefectReason(coverage trainingMonotonyCoverage) string {
	switch {
	case len(coverage.InvalidDates) > 0:
		return "invalid_daily_date"
	case len(coverage.OutOfWindowDates) > 0:
		return "out_of_window_daily_row"
	case len(coverage.DuplicateDates) > 0:
		return "duplicate_daily_date"
	case len(coverage.NonObjectRows) > 0:
		return "non_object_daily_row"
	case len(coverage.MalformedLoadRows) > 0:
		return "malformed_daily_load"
	case len(coverage.NegativeLoadRows) > 0:
		return "negative_daily_load"
	case len(coverage.MissingDates) > 0:
		return "missing_daily_rows"
	default:
		return ""
	}
}

func monotonyBoolPointer(value bool) *bool { return &value }

func monotonyMeta(coverage trainingMonotonyCoverage) analysis.AnalyzerMetaInput {
	return analysis.AnalyzerMetaInput{Method: "foster_training_load_monotony", SourceTools: []string{"get_training_summary"}, MissingAction: "refuse", FormulaRef: resources.AnalysisFormulaRefTrainingLoadMonotony, Assumptions: map[string]any{"coverage": coverage}}
}

func roundMonotony(value float64) float64 {
	if value > math.MaxFloat64/10000 {
		return value
	}
	return math.Round(value*10000) / 10000
}

func trainingMonotonyInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start_date", "end_date"}, "properties": map[string]any{
		"start_date":   map[string]any{"type": "string", "description": "Inclusive athlete-local start date, YYYY-MM-DD."},
		"end_date":     map[string]any{"type": "string", "description": "Inclusive athlete-local end date, YYYY-MM-DD; at most 31 expected dates."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include validated daily evidence in series."},
	}}
}

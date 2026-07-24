package tools

import (
	"context"
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type monotonyTestClient struct {
	rows  []intervals.RawSummaryRow
	calls int
}

func (c *monotonyTestClient) ListAthleteSummaryRaw(context.Context, intervals.AthleteSummaryParams) ([]intervals.RawSummaryRow, error) {
	c.calls++
	return c.rows, nil
}

func monotonyRows(t *testing.T, values ...string) []intervals.RawSummaryRow {
	t.Helper()
	rows := make([]intervals.RawSummaryRow, 0, len(values))
	for _, value := range values {
		raw := json.RawMessage(value)
		var object map[string]any
		if err := json.Unmarshal(raw, &object); err == nil {
			rows = append(rows, intervals.RawSummaryRow{Raw: object, RawJSON: raw})
		} else {
			rows = append(rows, intervals.RawSummaryRow{RawJSON: raw, DecodeError: err.Error()})
		}
	}
	return rows
}

func monotonyPayload(t *testing.T, result Result) map[string]any {
	t.Helper()
	return resultMap(t, result)
}

func TestDecodeTrainingMonotonyRequestUsesInclusiveBoundedDates(t *testing.T) {
	_, start, end, err := decodeTrainingMonotonyRequest(json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-31"}`))
	if err != nil || start.Format("2006-01-02") != "2026-01-01" || end.Format("2006-01-02") != "2026-01-31" {
		t.Fatalf("31-day request = %v, %v, %v; want inclusive valid window", start, end, err)
	}
	for _, raw := range []string{
		`{"start_date":"2026-01-02","end_date":"2026-01-01"}`,
		`{"start_date":"2026-02-01","end_date":"2026-03-04"}`,
		`{"start_date":"2026-01-01","end_date":"2026-02-01"}`,
	} {
		if _, _, _, err := decodeTrainingMonotonyRequest(json.RawMessage(raw)); err == nil {
			t.Fatalf("decodeTrainingMonotonyRequest(%s) error = nil, want bounded date rejection", raw)
		}
	}
}

func TestComputeTrainingMonotonyOneDayFetchPrecedesInsufficientDays(t *testing.T) {
	client := &monotonyTestClient{}
	tool := newComputeTrainingMonotonyTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := monotonyPayload(t, result)
	got := payload["result"].(map[string]any)
	if got["reason"] != "insufficient_days" || client.calls != 1 {
		t.Fatalf("one-day result = %#v, calls = %d, want insufficient_days after one read", got, client.calls)
	}
	meta := payload["_meta"].(map[string]any)
	if meta["source_tools"].([]any)[0] != "get_training_summary" || meta["insufficient_sample"] != true {
		t.Fatalf("one-day meta = %#v", meta)
	}
}

func TestComputeTrainingMonotonyEmptyMultiDayUsesNoDailyRows(t *testing.T) {
	tool := newComputeTrainingMonotonyTool(&monotonyTestClient{}, "test", false)
	result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-02"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := monotonyPayload(t, result)["result"].(map[string]any)
	if got["reason"] != "no_daily_rows" {
		t.Fatalf("empty multi-day result = %#v, want no_daily_rows", got)
	}
}

func TestInspectTrainingMonotonyRowsUsesOnlyRawRequiredFields(t *testing.T) {
	rows := monotonyRows(t,
		`{"date":"2026-01-01","training_load":0,"count":99,"byCategory":[{"training_load":999}]}`,
		`{"date":"2026-01-02","training_load":12.5,"sport":"Run"}`,
	)
	coverage, loads, _ := inspectTrainingMonotonyRows(rows, mustDate(t, "2026-01-01"), mustDate(t, "2026-01-02"))
	if coverage.ValidRows != 2 || len(loads) != 2 || loads[0] != 0 || loads[1] != 12.5 || len(coverage.MissingDates) != 0 {
		t.Fatalf("coverage = %#v, loads = %#v, want explicit zero/fraction raw loads only", coverage, loads)
	}
	invalid := monotonyRows(t,
		`{"date":"2026-01-01"}`,
		`{"date":"2026-01-02","training_load":"12"}`,
		`{"date":"2026-01-03","training_load":-1}`,
	)
	invalid = append(invalid, intervals.RawSummaryRow{RawJSON: json.RawMessage(`[]`), DecodeError: "not object"})
	badCoverage, _, _ := inspectTrainingMonotonyRows(invalid, mustDate(t, "2026-01-01"), mustDate(t, "2026-01-04"))
	if coverageDefectReason(badCoverage) != "non_object_daily_row" || len(badCoverage.MalformedLoadRows) != 2 || len(badCoverage.NegativeLoadRows) != 1 || len(badCoverage.NonObjectRows) != 1 {
		t.Fatalf("invalid reason = %q, coverage = %#v, want raw required-field defects and non-object", coverageDefectReason(badCoverage), badCoverage)
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}
	return date
}

func TestInspectTrainingMonotonyRowsCountsValidOutOfWindowDatesAndIgnoresDecodeMarkers(t *testing.T) {
	rows := monotonyRows(t,
		`{"date":"2025-12-31","training_load":5}`,
		`{"date":"2026-01-01","training_load":0}`,
	)
	rows[1].DecodeError = "optional field decode marker"
	coverage, loads, _ := inspectTrainingMonotonyRows(rows, mustDate(t, "2026-01-01"), mustDate(t, "2026-01-02"))
	if coverage.UniqueDates != 2 || coverage.ValidRows != 1 || len(loads) != 1 || len(coverage.DecodeErrorRows) != 0 || coverageDefectReason(coverage) != "out_of_window_daily_row" {
		t.Fatalf("coverage = %#v, loads = %#v, want out-of-window refusal and ignored decode marker", coverage, loads)
	}
}

func TestComputeTrainingMonotonyCoverageSerializesEveryKey(t *testing.T) {
	tool := newComputeTrainingMonotonyTool(&monotonyTestClient{}, "test", false)
	result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-02"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := monotonyPayload(t, result)["_meta"].(map[string]any)
	coverage := meta["assumptions"].(map[string]any)["coverage"].(map[string]any)
	for _, key := range []string{"source_tool", "source_field", "date_basis", "expected_days", "expected_dates", "received_rows", "unique_dates", "valid_rows", "valid_dates", "missing_dates", "duplicate_dates", "invalid_dates", "out_of_window_dates", "malformed_load_rows", "negative_load_rows", "non_object_rows", "decode_error_rows", "rejected_rows"} {
		if _, ok := coverage[key]; !ok {
			t.Fatalf("coverage missing key %q: %#v", key, coverage)
		}
	}
	if coverage["decode_error_rows"] == nil || len(coverage["decode_error_rows"].([]any)) != 0 {
		t.Fatalf("decode_error_rows = %#v, want empty array", coverage["decode_error_rows"])
	}
}

func TestComputeTrainingMonotonyDuplicateRowsRejectEveryOccurrence(t *testing.T) {
	client := &monotonyTestClient{rows: monotonyRows(t,
		`{"date":"2026-01-01","training_load":10}`,
		`{"date":"2026-01-01","training_load":10}`,
		`{"date":"2026-01-02","training_load":20}`)}
	tool := newComputeTrainingMonotonyTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-02"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := monotonyPayload(t, result)
	got := payload["result"].(map[string]any)
	if got["reason"] != "duplicate_daily_date" {
		t.Fatalf("duplicate result = %#v", got)
	}
	meta := payload["_meta"].(map[string]any)
	coverage := meta["assumptions"].(map[string]any)["coverage"].(map[string]any)
	if len(coverage["duplicate_dates"].([]any)) != 1 || len(coverage["rejected_rows"].([]any)) != 2 {
		t.Fatalf("duplicate coverage = %#v", coverage)
	}
}

func TestComputeTrainingMonotonyTerseAndFullEvidenceShapes(t *testing.T) {
	client := &monotonyTestClient{rows: monotonyRows(t,
		`{"date":"2026-01-01","training_load":0,"sport":"Ride"}`,
		`{"date":"2026-01-02","training_load":20,"sport":"Run"}`,
	)}
	tool := newComputeTrainingMonotonyTool(client, "test", false)
	terse, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-02"}`)})
	if err != nil {
		t.Fatalf("terse Handler() error = %v", err)
	}
	tersePayload := monotonyPayload(t, terse)
	if _, ok := tersePayload["series"]; ok {
		t.Fatalf("terse response includes daily series: %#v", tersePayload)
	}
	full, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-02","include_full":true}`)})
	if err != nil {
		t.Fatalf("full Handler() error = %v", err)
	}
	series := monotonyPayload(t, full)["series"].([]any)
	if len(series) != 2 || series[0].(map[string]any)["training_load"] != float64(0) || series[1].(map[string]any)["training_load"] != float64(20) {
		t.Fatalf("full series = %#v, want validated daily loads only", series)
	}
}

func TestComputeTrainingMonotonyRefusalAndZeroVarianceOutputsOmitStatistics(t *testing.T) {
	cases := []struct {
		name   string
		rows   []intervals.RawSummaryRow
		range_ string
		status string
		reason string
		n      float64
	}{
		{name: "missing", range_: `{"start_date":"2026-01-01","end_date":"2026-01-02"}`, status: "unavailable", reason: "no_daily_rows", n: 0},
		{name: "zero variance", rows: monotonyRows(t, `{"date":"2026-01-01","training_load":10}`, `{"date":"2026-01-02","training_load":10}`), range_: `{"start_date":"2026-01-01","end_date":"2026-01-02"}`, status: "undefined", reason: "zero_variance", n: 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tool := newComputeTrainingMonotonyTool(&monotonyTestClient{rows: tc.rows}, "test", false)
			result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(tc.range_)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := monotonyPayload(t, result)
			body := payload["result"].(map[string]any)
			if body["status"] != tc.status || body["reason"] != tc.reason {
				t.Fatalf("result = %#v, want %s/%s", body, tc.status, tc.reason)
			}
			for _, key := range []string{"mean_daily_load", "standard_deviation", "monotony"} {
				if _, ok := body[key]; ok {
					t.Fatalf("refusal result includes %s: %#v", key, body)
				}
			}
			meta := payload["_meta"].(map[string]any)
			wantMissing := float64(0)
			if tc.n == 0 {
				wantMissing = 2
			}
			if meta["n"] != tc.n || meta["missing_days"] != wantMissing || meta["insufficient_sample"] != (tc.n == 0) {
				t.Fatalf("meta = %#v", meta)
			}
		})
	}
}

func TestParseSummaryLoadPinsFiniteJSONNumberBoundary(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		state string
	}{
		{raw: "0.25", state: "valid"},
		{raw: "0", state: "valid"},
		{raw: "-0.5", state: "negative"},
		{raw: "1e-400", state: "malformed"},
		{raw: "1e400", state: "malformed"},
		{raw: "-1e400", state: "malformed"},
		{raw: "null", state: "malformed"},
		{raw: `"1"`, state: "malformed"},
	} {
		_, state := parseSummaryLoad(json.RawMessage(tc.raw))
		if state != tc.state {
			t.Fatalf("parseSummaryLoad(%s) state = %q, want %q", tc.raw, state, tc.state)
		}
	}
}

func TestRoundMonotonyUsesHalfAwayFromZeroAndProtectsLargeValues(t *testing.T) {
	if got := roundMonotony(1.23485); got != 1.2349 {
		t.Fatalf("roundMonotony(tie) = %v, want 1.2349", got)
	}
	nearMax := math.MaxFloat64 / 10000
	if got := roundMonotony(nearMax); got != math.Round(nearMax*10000)/10000 {
		t.Fatalf("roundMonotony(near max) = %v, want finite rounded value", got)
	}
	above := math.Nextafter(nearMax, math.Inf(1))
	if got := roundMonotony(above); got != above || math.IsInf(got, 0) {
		t.Fatalf("roundMonotony(above max precision) = %v, want unchanged finite value", got)
	}
}

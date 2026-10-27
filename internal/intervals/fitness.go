package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
)

// AthleteSummaryParams contains date filters for athlete summary rows.
type AthleteSummaryParams struct {
	Start string
	End   string
}

// SummaryWithCats contains athlete summary fields used by fitness and training-summary tools.
type SummaryWithCats struct {
	Raw map[string]any `json:"-"`

	AthleteID          string            `json:"athlete_id"`
	Date               string            `json:"date"`
	Count              int               `json:"count"`
	Time               int               `json:"time"`
	MovingTime         int               `json:"moving_time"`
	ElapsedTime        int               `json:"elapsed_time"`
	Calories           int               `json:"calories"`
	TotalElevationGain float64           `json:"total_elevation_gain"`
	TrainingLoad       int               `json:"training_load"`
	SRPE               int               `json:"srpe"`
	Distance           float64           `json:"distance"`
	Fitness            float64           `json:"fitness"`
	Fatigue            float64           `json:"fatigue"`
	Form               float64           `json:"form"`
	TimeInZones        []float64         `json:"timeInZones"`
	TimeInZonesTot     int               `json:"timeInZonesTot"`
	ByCategory         []CategorySummary `json:"byCategory"`
}

// UnmarshalJSON decodes SummaryWithCats while retaining the original object for full responses.
func (s *SummaryWithCats) UnmarshalJSON(data []byte) error {
	type summaryAlias SummaryWithCats
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded summaryAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*s = SummaryWithCats(decoded)
	s.Raw = raw
	return nil
}

// CategorySummary contains per-category athlete summary totals.
type CategorySummary struct {
	Raw map[string]any `json:"-"`

	Category           string  `json:"category"`
	Count              int     `json:"count"`
	Time               int     `json:"time"`
	MovingTime         int     `json:"moving_time"`
	ElapsedTime        int     `json:"elapsed_time"`
	Calories           int     `json:"calories"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	TrainingLoad       int     `json:"training_load"`
	SRPE               int     `json:"srpe"`
	Distance           float64 `json:"distance"`
}

// UnmarshalJSON decodes CategorySummary while retaining the original object for full responses.
func (s *CategorySummary) UnmarshalJSON(data []byte) error {
	type categoryAlias CategorySummary
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded categoryAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*s = CategorySummary(decoded)
	s.Raw = raw
	return nil
}

// ListAthleteSummary retrieves request-target rows whose upstream bucket anchors are inside the requested date window.
func (c *Client) ListAthleteSummary(ctx context.Context, params AthleteSummaryParams) ([]SummaryWithCats, error) {
	params, err := normalizeAthleteSummaryParams(params)
	if err != nil {
		return nil, fmt.Errorf("listing athlete summary: %w", err)
	}
	targetAthleteID, err := c.athleteSummaryTargetID(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing athlete summary: %w", err)
	}
	var elements []json.RawMessage
	if err := c.doJSONQuery(ctx, &elements, athleteSummaryQuery(params), "athlete", c.athleteID, "athlete-summary.json"); err != nil {
		return nil, fmt.Errorf("listing athlete summary: %w", err)
	}
	rows := make([]SummaryWithCats, 0, len(elements))
	for _, element := range elements {
		raw, rowAthleteID, owned, err := athleteSummaryElementOwnership(element, targetAthleteID)
		if err != nil {
			return nil, fmt.Errorf("listing athlete summary: %w", err)
		}
		if !owned {
			continue
		}
		withinWindow, err := athleteSummaryRowWithinWindow(raw, params)
		if err != nil {
			return nil, fmt.Errorf("listing athlete summary: %w", err)
		}
		if !withinWindow {
			continue
		}
		var row SummaryWithCats
		if err := json.Unmarshal(element, &row); err != nil {
			return nil, fmt.Errorf("listing athlete summary: decoding target row: %w", err)
		}
		row.Raw = raw
		row.AthleteID = rowAthleteID
		rows = append(rows, row)
	}
	return rows, nil
}

// RawSummaryRow preserves one athlete-summary array element for strict analyzer coverage checks.
type RawSummaryRow struct {
	Raw         map[string]any
	RawJSON     json.RawMessage
	DecodeError string
}

// ListAthleteSummaryRaw retrieves target-athlete summary elements without discarding date evidence required by strict analyzers.
func (c *Client) ListAthleteSummaryRaw(ctx context.Context, params AthleteSummaryParams) ([]RawSummaryRow, error) {
	params, err := normalizeAthleteSummaryParams(params)
	if err != nil {
		return nil, fmt.Errorf("listing raw athlete summary: %w", err)
	}
	targetAthleteID, err := c.athleteSummaryTargetID(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing raw athlete summary: %w", err)
	}
	var elements []json.RawMessage
	if err := c.doJSONQuery(ctx, &elements, athleteSummaryQuery(params), "athlete", c.athleteID, "athlete-summary.json"); err != nil {
		return nil, fmt.Errorf("listing raw athlete summary: %w", err)
	}
	rows := make([]RawSummaryRow, 0, len(elements))
	for _, element := range elements {
		raw, _, owned, err := athleteSummaryElementOwnership(element, targetAthleteID)
		if err != nil {
			return nil, fmt.Errorf("listing raw athlete summary: %w", err)
		}
		if !owned {
			continue
		}
		row := RawSummaryRow{Raw: raw, RawJSON: append(json.RawMessage(nil), element...)}
		var typed SummaryWithCats
		if err := json.Unmarshal(element, &typed); err != nil {
			row.DecodeError = err.Error()
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (c *Client) athleteSummaryTargetID(ctx context.Context) (string, error) {
	targetAthleteID := c.athleteID
	if requestTarget, ok := targetAthleteIDFromContext(ctx); ok {
		targetAthleteID = requestTarget
	}
	normalized, err := config.NormalizeAthleteID(targetAthleteID)
	if err != nil {
		return "", ErrTargetAthleteMismatch
	}
	return normalized, nil
}

func athleteSummaryRowOwnedBy(rowAthleteID string, targetAthleteID string) (bool, error) {
	normalized, err := config.NormalizeAthleteID(rowAthleteID)
	if err != nil {
		return false, ErrTargetAthleteMismatch
	}
	return normalized == targetAthleteID, nil
}

func athleteSummaryElementOwnership(element json.RawMessage, targetAthleteID string) (map[string]any, string, bool, error) {
	var raw map[string]any
	if err := json.Unmarshal(element, &raw); err != nil || raw == nil {
		return nil, "", false, ErrTargetAthleteMismatch
	}
	athleteID, ok := raw["athlete_id"].(string)
	if !ok {
		return nil, "", false, ErrTargetAthleteMismatch
	}
	owned, err := athleteSummaryRowOwnedBy(athleteID, targetAthleteID)
	if err != nil {
		return nil, "", false, err
	}
	return raw, athleteID, owned, nil
}

func normalizeAthleteSummaryParams(params AthleteSummaryParams) (AthleteSummaryParams, error) {
	params.Start = strings.TrimSpace(params.Start)
	params.End = strings.TrimSpace(params.End)
	for _, field := range []struct {
		label string
		value string
	}{{label: "start", value: params.Start}, {label: "end", value: params.End}} {
		if field.value == "" {
			continue
		}
		if _, err := time.Parse(time.DateOnly, field.value); err != nil {
			return AthleteSummaryParams{}, fmt.Errorf("%s must be YYYY-MM-DD", field.label)
		}
	}
	if params.Start != "" && params.End != "" && params.End < params.Start {
		return AthleteSummaryParams{}, fmt.Errorf("end must be on or after start")
	}
	return params, nil
}

func athleteSummaryRowWithinWindow(raw map[string]any, params AthleteSummaryParams) (bool, error) {
	date, ok := raw["date"].(string)
	if !ok {
		return false, fmt.Errorf("target row date must be YYYY-MM-DD")
	}
	date = strings.TrimSpace(date)
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return false, fmt.Errorf("target row date must be YYYY-MM-DD")
	}
	if params.Start != "" && date < params.Start {
		return false, nil
	}
	if params.End != "" && date > params.End {
		return false, nil
	}
	return true, nil
}

func athleteSummaryQuery(params AthleteSummaryParams) url.Values {
	query := url.Values{}
	if start := strings.TrimSpace(params.Start); start != "" {
		query.Set("start", start)
	}
	if end := strings.TrimSpace(params.End); end != "" {
		query.Set("end", end)
	}
	return query
}

// CurveParams contains query parameters for athlete curve endpoints.
type CurveParams struct {
	Sport           string
	Now             string
	CurveSpec       string
	DurationSeconds []int
	DistanceMeters  []int
}

// DataCurveSet contains intervals.icu athlete curve rows and related activity metadata.
type DataCurveSet struct {
	Raw map[string]any `json:"-"`

	List       []DataCurve    `json:"list"`
	Activities map[string]any `json:"activities"`
}

// UnmarshalJSON decodes DataCurveSet while retaining the original object for full responses.
func (s *DataCurveSet) UnmarshalJSON(data []byte) error {
	type setAlias DataCurveSet
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded setAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*s = DataCurveSet(decoded)
	s.Raw = raw
	return nil
}

// DataCurve contains one upstream athlete curve.
type DataCurve struct {
	Raw map[string]any `json:"-"`

	ID             string       `json:"id"`
	Label          string       `json:"label"`
	FilterLabel    string       `json:"filter_label"`
	StartDateLocal string       `json:"start_date_local"`
	EndDateLocal   string       `json:"end_date_local"`
	Days           int          `json:"days"`
	MovingTime     int          `json:"moving_time"`
	TrainingLoad   int          `json:"training_load"`
	Weight         float64      `json:"weight"`
	AfterKJ        *int         `json:"after_kj"`
	Secs           []float64    `json:"secs"`
	Distance       []float64    `json:"distance"`
	Values         []float64    `json:"values"`
	ActivityID     []string     `json:"activity_id"`
	Watts          []float64    `json:"watts"`
	WattsPerKG     []*float64   `json:"watts_per_kg"`
	WKGActivityID  []string     `json:"wkg_activity_id"`
	PowerModels    []PowerModel `json:"-"`
}

// PowerModel contains optional model values fitted by intervals.icu, never local estimates.
type PowerModel struct {
	Type          string   `json:"type"`
	CriticalPower *float64 `json:"criticalPower"`
	WPrime        *float64 `json:"wPrime"`
	PMax          *float64 `json:"pMax"`
	FTP           *float64 `json:"ftp"`
}

// UnmarshalJSON decodes DataCurve while retaining the original object for full responses.
func (c *DataCurve) UnmarshalJSON(data []byte) error {
	type curveAlias DataCurve
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded curveAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*c = DataCurve(decoded)
	// Model data is optional: malformed model values must not discard valid effort arrays.
	var modelPayload struct {
		Models []json.RawMessage `json:"powerModels"`
	}
	if json.Unmarshal(data, &modelPayload) == nil {
		for _, modelJSON := range modelPayload.Models {
			var model PowerModel
			if json.Unmarshal(modelJSON, &model) == nil {
				c.PowerModels = append(c.PowerModels, model)
			}
		}
	}
	c.Raw = raw
	return nil
}

// ListAthletePowerCurves retrieves upstream-computed athlete power curves.
func (c *Client) ListAthletePowerCurves(ctx context.Context, params CurveParams) (DataCurveSet, error) {
	set, err := c.listAthleteCurveSet(ctx, params, "power-curves.json", true)
	if err != nil {
		return DataCurveSet{}, fmt.Errorf("listing athlete power curves: %w", err)
	}
	return set, nil
}

// ListAthleteHRCurves retrieves upstream-computed athlete heart-rate curves.
func (c *Client) ListAthleteHRCurves(ctx context.Context, params CurveParams) (DataCurveSet, error) {
	set, err := c.listAthleteCurveSet(ctx, params, "hr-curves.json", false)
	if err != nil {
		return DataCurveSet{}, fmt.Errorf("listing athlete heart-rate curves: %w", err)
	}
	return set, nil
}

// ListAthletePaceCurves retrieves upstream-computed athlete pace curves.
func (c *Client) ListAthletePaceCurves(ctx context.Context, params CurveParams) (DataCurveSet, error) {
	set, err := c.listAthleteCurveSet(ctx, params, "pace-curves.json", false)
	if err != nil {
		return DataCurveSet{}, fmt.Errorf("listing athlete pace curves: %w", err)
	}
	return set, nil
}

func (c *Client) listAthleteCurveSet(ctx context.Context, params CurveParams, endpoint string, requireType bool) (DataCurveSet, error) {
	query := url.Values{}
	if sport := strings.TrimSpace(params.Sport); sport != "" {
		query.Set("type", sport)
	} else if requireType {
		return DataCurveSet{}, fmt.Errorf("sport type is required")
	}
	if now := strings.TrimSpace(params.Now); now != "" {
		if _, err := time.Parse(time.DateOnly, now); err != nil {
			return DataCurveSet{}, fmt.Errorf("now must be athlete-local YYYY-MM-DD")
		}
		query.Set("now", now)
	}
	if curve := strings.TrimSpace(params.CurveSpec); curve != "" {
		query.Set("curves", curve)
	}
	if len(params.DurationSeconds) > 0 {
		query.Set("secs", joinInts(params.DurationSeconds))
	}
	if len(params.DistanceMeters) > 0 {
		query.Set("distances", joinInts(params.DistanceMeters))
	}
	var set DataCurveSet
	if err := c.doJSONQuery(ctx, &set, query, "athlete", c.athleteID, endpoint); err != nil {
		return DataCurveSet{}, err
	}
	return set, nil
}

func joinInts(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if value > 0 {
			parts = append(parts, strconv.Itoa(value))
		}
	}
	return strings.Join(parts, ",")
}

package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// AthleteSummaryParams contains date filters for athlete summary rows.
type AthleteSummaryParams struct {
	Start string
	End   string
}

// SummaryWithCats contains athlete summary fields used by fitness and training-summary tools.
type SummaryWithCats struct {
	Raw map[string]any `json:"-"`

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

// ListAthleteSummary retrieves daily athlete summary rows for the configured athlete.
func (c *Client) ListAthleteSummary(ctx context.Context, params AthleteSummaryParams) ([]SummaryWithCats, error) {
	query := url.Values{}
	if start := strings.TrimSpace(params.Start); start != "" {
		query.Set("start", start)
	}
	if end := strings.TrimSpace(params.End); end != "" {
		query.Set("end", end)
	}
	var rows []SummaryWithCats
	if err := c.doJSONQuery(ctx, &rows, query, "athlete", c.athleteID, "athlete-summary.json"); err != nil {
		return nil, fmt.Errorf("listing athlete summary: %w", err)
	}
	return rows, nil
}

// CurveParams contains query parameters for athlete curve endpoints.
type CurveParams struct {
	Sport           string
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

	ID             string    `json:"id"`
	Label          string    `json:"label"`
	FilterLabel    string    `json:"filter_label"`
	StartDateLocal string    `json:"start_date_local"`
	EndDateLocal   string    `json:"end_date_local"`
	Days           int       `json:"days"`
	MovingTime     int       `json:"moving_time"`
	TrainingLoad   int       `json:"training_load"`
	Weight         float64   `json:"weight"`
	Secs           []float64 `json:"secs"`
	Distance       []float64 `json:"distance"`
	Values         []float64 `json:"values"`
	ActivityID     []string  `json:"activity_id"`
	Watts          []float64 `json:"watts"`
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

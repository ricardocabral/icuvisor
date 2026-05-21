package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// WellnessParams contains date filters for wellness rows.
type WellnessParams struct {
	Oldest string
	Newest string
	Fields []string
}

// WriteWellnessParams contains sparse writable wellness fields for one local date.
type WriteWellnessParams struct {
	Date           string
	Feel           *int
	Fatigue        *int
	Mood           *int
	SleepQuality   *int
	Motivation     *int
	Soreness       *int
	Stress         *int
	Weight         *float64
	BodyFat        *float64
	Systolic       *int
	Diastolic      *int
	BloodGlucose   *float64
	Lactate        *float64
	SpO2           *float64
	VO2Max         *float64
	Abdomen        *float64
	Respiration    *float64
	MenstrualPhase *string
	RestingHR      *int
	HRV            *float64
	Injury         *string
	Locked         *bool
}

// Wellness contains typed intervals.icu wellness fields while preserving raw upstream fields.
type Wellness struct {
	Raw               map[string]any            `json:"-"`
	Native            map[string]map[string]any `json:"-"`
	NativeClaimedKeys []string                  `json:"-"`

	ID                      *string  `json:"id"`
	CTL                     *float64 `json:"ctl"`
	ATL                     *float64 `json:"atl"`
	RampRate                *float64 `json:"rampRate"`
	CTLLoad                 *float64 `json:"ctlLoad"`
	ATLLoad                 *float64 `json:"atlLoad"`
	SportInfo               any      `json:"sportInfo"`
	Updated                 *string  `json:"updated"`
	Weight                  *float64 `json:"weight"`
	RestingHR               *int     `json:"restingHR"`
	HRV                     *float64 `json:"hrv"`
	HRVSDNN                 *float64 `json:"hrvSDNN"`
	MenstrualPhase          *string  `json:"menstrualPhase"`
	MenstrualPhasePredicted *string  `json:"menstrualPhasePredicted"`
	KcalConsumed            *int     `json:"kcalConsumed"`
	SleepSecs               *int     `json:"sleepSecs"`
	SleepScore              *float64 `json:"sleepScore"`
	SleepQuality            *int     `json:"sleepQuality"`
	AvgSleepingHR           *float64 `json:"avgSleepingHR"`
	Feel                    *int     `json:"feel"`
	Soreness                *int     `json:"soreness"`
	Fatigue                 *int     `json:"fatigue"`
	Stress                  *int     `json:"stress"`
	Mood                    *int     `json:"mood"`
	Motivation              *int     `json:"motivation"`
	Injury                  any      `json:"injury"`
	SpO2                    *float64 `json:"spO2"`
	Systolic                *int     `json:"systolic"`
	Diastolic               *int     `json:"diastolic"`
	Hydration               *int     `json:"hydration"`
	HydrationVolume         *float64 `json:"hydrationVolume"`
	Readiness               *float64 `json:"readiness"`
	BaevskySI               *float64 `json:"baevskySI"`
	BloodGlucose            *float64 `json:"bloodGlucose"`
	Lactate                 *float64 `json:"lactate"`
	BodyFat                 *float64 `json:"bodyFat"`
	Abdomen                 *float64 `json:"abdomen"`
	VO2Max                  *float64 `json:"vo2max"`
	Comments                *string  `json:"comments"`
	Steps                   *int     `json:"steps"`
	Respiration             *float64 `json:"respiration"`
	Carbohydrates           *float64 `json:"carbohydrates"`
	Protein                 *float64 `json:"protein"`
	FatTotal                *float64 `json:"fatTotal"`
	Locked                  *bool    `json:"locked"`
	TempWeight              *bool    `json:"tempWeight"`
	TempRestingHR           *bool    `json:"tempRestingHR"`
}

// UnmarshalJSON decodes Wellness while retaining raw and native provider fields.
func (w *Wellness) UnmarshalJSON(data []byte) error {
	type wellnessAlias Wellness
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded wellnessAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*w = Wellness(decoded)
	w.Raw = raw
	w.Native, w.NativeClaimedKeys = extractWellnessNative(raw)
	return nil
}

// ListWellness retrieves wellness rows in ascending local-date range for the configured athlete.
func (c *Client) ListWellness(ctx context.Context, params WellnessParams) ([]Wellness, error) {
	query := url.Values{}
	oldest := strings.TrimSpace(params.Oldest)
	if oldest == "" {
		return nil, fmt.Errorf("listing wellness: oldest is required")
	}
	query.Set("oldest", oldest)
	if newest := strings.TrimSpace(params.Newest); newest != "" {
		query.Set("newest", newest)
	}
	if len(params.Fields) > 0 {
		fields := compactStrings(params.Fields)
		if len(fields) > 0 {
			query.Set("fields", strings.Join(fields, ","))
		}
	}

	var rows []Wellness
	if err := c.doJSONQuery(ctx, &rows, query, "athlete", c.athleteID, "wellness.json"); err != nil {
		return nil, fmt.Errorf("listing wellness: %w", err)
	}
	return rows, nil
}

// UpdateWellness updates sparse manual wellness fields for one athlete-local date.
func (c *Client) UpdateWellness(ctx context.Context, params WriteWellnessParams) (Wellness, error) {
	date := strings.TrimSpace(params.Date)
	if date == "" {
		return Wellness{}, fmt.Errorf("updating wellness: date is required")
	}
	body, err := writeWellnessBody(params)
	if err != nil {
		return Wellness{}, err
	}
	var row Wellness
	if err := c.doJSONBody(ctx, http.MethodPut, body, &row, "athlete", c.athleteID, "wellness", date); err != nil {
		return Wellness{}, fmt.Errorf("updating wellness %s: %w", date, err)
	}
	return row, nil
}

func writeWellnessBody(params WriteWellnessParams) (map[string]any, error) {
	if params.Feel != nil {
		return nil, ErrUnsupportedWellnessFeel
	}
	body := map[string]any{}
	setSparse(body, "fatigue", params.Fatigue)
	setSparse(body, "mood", params.Mood)
	setSparse(body, "sleepQuality", params.SleepQuality)
	setSparse(body, "motivation", params.Motivation)
	setSparse(body, "soreness", params.Soreness)
	setSparse(body, "stress", params.Stress)
	setSparse(body, "weight", params.Weight)
	setSparse(body, "bodyFat", params.BodyFat)
	setSparse(body, "systolic", params.Systolic)
	setSparse(body, "diastolic", params.Diastolic)
	setSparse(body, "bloodGlucose", params.BloodGlucose)
	setSparse(body, "lactate", params.Lactate)
	setSparse(body, "spO2", params.SpO2)
	setSparse(body, "vo2max", params.VO2Max)
	setSparse(body, "abdomen", params.Abdomen)
	setSparse(body, "respiration", params.Respiration)
	setSparse(body, "menstrualPhase", params.MenstrualPhase)
	setSparse(body, "restingHR", params.RestingHR)
	setSparse(body, "hrv", params.HRV)
	setSparse(body, "injury", params.Injury)
	setSparse(body, "locked", params.Locked)
	if len(body) == 0 {
		return nil, fmt.Errorf("updating wellness: at least one field is required")
	}
	return body, nil
}

func setSparse[T any](body map[string]any, key string, value *T) {
	if value != nil {
		body[key] = *value
	}
}

func extractWellnessNative(raw map[string]any) (map[string]map[string]any, []string) {
	native := map[string]map[string]any{}
	claimed := []string{}
	claim := func(source, field, key string, value any) {
		if value == nil {
			return
		}
		if native[source] == nil {
			native[source] = map[string]any{}
		}
		native[source][field] = value
		if key != "" {
			claimed = append(claimed, key)
		}
	}
	claimNested := func(source string, aliases map[string]string) {
		provider, ok := raw[source].(map[string]any)
		if !ok {
			return
		}
		matched := false
		for key, field := range aliases {
			if value, ok := provider[key]; ok {
				claim(source, field, "", value)
				matched = true
			}
		}
		if matched {
			claimed = append(claimed, source)
		}
	}

	claimNested("polar", map[string]string{"ans_charge": "ans_charge", "nightly_recharge_status": "nightly_recharge_status", "sleep_score": "sleep_score"})
	claimNested("garmin", map[string]string{"body_battery_min": "body_battery_min", "body_battery_max": "body_battery_max", "bodyBatteryMin": "body_battery_min", "bodyBatteryMax": "body_battery_max", "sleep_score": "sleep_score", "sleepScore": "sleep_score"})
	claimNested("oura", map[string]string{"readiness_score": "readiness_score", "readinessScore": "readiness_score", "sleep_score": "sleep_score", "sleepScore": "sleep_score"})
	claimNested("whoop", map[string]string{"recovery_score": "recovery_score", "recoveryScore": "recovery_score", "sleep_performance_percentage": "sleep_performance_percentage", "sleepPerformancePercentage": "sleep_performance_percentage"})

	for key, spec := range map[string]struct{ source, field string }{
		"ans_charge":                               {"polar", "ans_charge"},
		"nightly_recharge_status":                  {"polar", "nightly_recharge_status"},
		"polar_ans_charge":                         {"polar", "ans_charge"},
		"polar_sleep_score":                        {"polar", "sleep_score"},
		"polar_nightly_recharge_status":            {"polar", "nightly_recharge_status"},
		"body_battery_min":                         {"garmin", "body_battery_min"},
		"body_battery_max":                         {"garmin", "body_battery_max"},
		"garmin_body_battery_min":                  {"garmin", "body_battery_min"},
		"garmin_body_battery_max":                  {"garmin", "body_battery_max"},
		"garmin_sleep_score":                       {"garmin", "sleep_score"},
		"oura_readiness_score":                     {"oura", "readiness_score"},
		"oura_sleep_score":                         {"oura", "sleep_score"},
		"recovery_score":                           {"whoop", "recovery_score"},
		"sleep_performance_percentage":             {"whoop", "sleep_performance_percentage"},
		"whoop_recovery_score":                     {"whoop", "recovery_score"},
		"whoop_sleep_performance_percentage":       {"whoop", "sleep_performance_percentage"},
		"whoop_sleep_performance_percent":          {"whoop", "sleep_performance_percentage"},
		"whoop_sleep_performance_percentage_score": {"whoop", "sleep_performance_percentage"},
	} {
		if value, ok := raw[key]; ok {
			claim(spec.source, spec.field, key, value)
		}
	}
	if value, ok := raw["sleep_score"]; ok {
		source := nativeSleepScoreSource(raw)
		if source == "polar" || source == "oura" {
			claim(source, "sleep_score", "sleep_score", value)
		}
	}
	if len(native) == 0 {
		return nil, nil
	}
	return native, dedupeStrings(claimed)
}

func nativeSleepScoreSource(raw map[string]any) string {
	for _, key := range []string{"source", "provider", "device", "wellnessSource", "wellness_source", "integration"} {
		value, ok := raw[key].(string)
		if !ok {
			continue
		}
		lower := strings.ToLower(value)
		if strings.Contains(lower, "garmin") {
			return "garmin"
		}
		if strings.Contains(lower, "oura") {
			return "oura"
		}
		if strings.Contains(lower, "polar") {
			return "polar"
		}
		if strings.Contains(lower, "whoop") {
			return "whoop"
		}
	}
	return ""
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

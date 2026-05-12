package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/units"
)

const (
	getExtendedMetricsName                 = "get_extended_metrics"
	getExtendedMetricsDescription          = "Get one activity's upstream-exposed extended metrics by activity_id. Terse mode drops unavailable fields and never computes or zero-fills missing metrics; include_full returns raw upstream payloads."
	invalidExtendedMetricsArgumentsMessage = "invalid get_extended_metrics arguments; provide activity_id and optional include_full"
	fetchExtendedMetricsMessage            = "could not fetch extended metrics; check activity_id and intervals.icu credentials"
)

var droppedExtendedMetricFields = []string{"ground_contact_time_ms", "vertical_oscillation_cm", "ground_contact_time_balance_percent", "core_temperature_c", "hr_drift_percent", "cadence_by_zone"}

// ExtendedMetricsClient retrieves activity sources used by get_extended_metrics.
type ExtendedMetricsClient interface {
	GetActivity(context.Context, string) (intervals.Activity, error)
	GetActivityIntervals(context.Context, string) (intervals.IntervalsDTO, error)
	GetActivityPowerVsHR(context.Context, string) (intervals.PowerVsHR, error)
}

type extendedMetricsRequest struct {
	ActivityID  string `json:"activity_id"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type extendedMetricsResponse struct {
	ActivityID     string                    `json:"activity_id"`
	Metrics        *extendedActivityMetrics  `json:"metrics,omitempty"`
	StravaImported bool                      `json:"strava_imported,omitempty"`
	Unavailable    *unavailableReason        `json:"unavailable,omitempty"`
	Intervals      []extendedIntervalMetrics `json:"intervals,omitempty"`
	Full           map[string]any            `json:"full,omitempty"`
	Meta           extendedMetricsMeta       `json:"_meta"`
}

type extendedActivityMetrics struct {
	StrideLengthM                *float64  `json:"stride_length_m,omitempty"`
	CardiacDecouplingPercent     *float64  `json:"cardiac_decoupling_percent,omitempty"`
	PWHR                         *float64  `json:"pw_hr,omitempty"`
	AerobicDecouplingPercent     *float64  `json:"aerobic_decoupling_percent,omitempty"`
	PowerZoneDistributionSeconds []float64 `json:"power_zone_distribution_seconds,omitempty"`
	PaceZoneTimeSeconds          []float64 `json:"pace_zone_time_seconds,omitempty"`
	JoulesAboveFTPKJ             *float64  `json:"joules_above_ftp_kj,omitempty"`
	IntensityFactor              *float64  `json:"intensity_factor,omitempty"`
	VariabilityIndex             *float64  `json:"variability_index,omitempty"`
	PolarizationIndex            *float64  `json:"polarization_index,omitempty"`
	TRIMP                        *float64  `json:"trimp,omitempty"`
	StrainScore                  *float64  `json:"strain_score,omitempty"`
	HRLoad                       *float64  `json:"hr_load,omitempty"`
	PaceLoad                     *float64  `json:"pace_load,omitempty"`
	PowerLoad                    *float64  `json:"power_load,omitempty"`
	TrainingLoad                 *float64  `json:"training_load,omitempty"`
	LeftRightBalancePercent      *float64  `json:"left_right_balance_percent,omitempty"`
	RPE                          *float64  `json:"rpe,omitempty"`
	Feel                         *float64  `json:"feel,omitempty"`
	SessionRPE                   *float64  `json:"session_rpe,omitempty"`
	CompliancePercent            *float64  `json:"compliance_percent,omitempty"`
	DeviceName                   string    `json:"device_name,omitempty"`
}

type extendedIntervalMetrics struct {
	IntervalID               string   `json:"interval_id,omitempty"`
	Label                    string   `json:"label,omitempty"`
	DFAAlpha1                *float64 `json:"dfa_alpha1,omitempty"`
	WPrimeBalanceStartKJ     *float64 `json:"w_prime_balance_start_kj,omitempty"`
	WPrimeBalanceEndKJ       *float64 `json:"w_prime_balance_end_kj,omitempty"`
	JoulesAboveFTPKJ         *float64 `json:"joules_above_ftp_kj,omitempty"`
	AerobicDecouplingPercent *float64 `json:"aerobic_decoupling_percent,omitempty"`
	LeftRightBalancePercent  *float64 `json:"left_right_balance_percent,omitempty"`
	StrideLengthM            *float64 `json:"stride_length_m,omitempty"`
	StrainScore              *float64 `json:"strain_score,omitempty"`
	TrainingLoad             *float64 `json:"training_load,omitempty"`
}

type extendedMetricsMeta struct {
	ServerVersion       string            `json:"server_version"`
	IncludeFull         bool              `json:"include_full"`
	ExtendedMetricUnits map[string]string `json:"extended_metric_units"`
	DroppedFields       []string          `json:"dropped_fields"`
	Partial             bool              `json:"partial,omitempty"`
	UnavailableSources  []string          `json:"unavailable_sources,omitempty"`
}

func newGetExtendedMetricsTool(client ExtendedMetricsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return Tool{Name: getExtendedMetricsName, Description: getExtendedMetricsDescription, InputSchema: extendedMetricsInputSchema(), OutputSchema: genericOutputSchema("Upstream-exposed extended metrics for one activity."), Handler: getExtendedMetricsHandler(client, profileClient, version, timezoneFallback, debugMetadata)}
}

func getExtendedMetricsHandler(client ExtendedMetricsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeExtendedMetricsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidExtendedMetricsArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchExtendedMetricsMessage, err)
		}
		activity, err := client.GetActivity(ctx, args.ActivityID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchExtendedMetricsMessage, err)
		}
		if isStravaBlocked(activity) {
			payload := stravaUnavailableExtendedMetricsResponse(args.ActivityID, activity, args.IncludeFull, version)
			return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, getExtendedMetricsName, unitSystem)
		}
		var unavailable []string
		intervalsDTO, intervalsOK, err := optionalIntervals(ctx, client, args.ActivityID)
		if err != nil {
			return Result{}, NewUserError(fetchExtendedMetricsMessage, err)
		}
		if !intervalsOK {
			unavailable = append(unavailable, "intervals")
		}
		powerVsHR, powerVsHROK, err := optionalPowerVsHR(ctx, client, args.ActivityID)
		if err != nil {
			return Result{}, NewUserError(fetchExtendedMetricsMessage, err)
		}
		if !powerVsHROK {
			unavailable = append(unavailable, "power_vs_hr")
		}
		payload := shapeExtendedMetrics(args.ActivityID, activity, intervalsDTO, intervalsOK, powerVsHR, powerVsHROK, args.IncludeFull, version, unavailable)
		return encodeShaped(payload, args.IncludeFull, []string{"intervals"}, version, debugMetadata, getExtendedMetricsName, unitSystem)
	}
}

func decodeExtendedMetricsRequest(raw json.RawMessage) (extendedMetricsRequest, error) {
	var args extendedMetricsRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.ActivityID = strings.TrimSpace(args.ActivityID)
	if args.ActivityID == "" {
		return args, errors.New("activity_id is required")
	}
	return args, nil
}

func optionalIntervals(ctx context.Context, client ExtendedMetricsClient, activityID string) (intervals.IntervalsDTO, bool, error) {
	dto, err := client.GetActivityIntervals(ctx, activityID)
	if err == nil {
		return dto, true, nil
	}
	if errors.Is(err, intervals.ErrNotFound) || errors.Is(err, intervals.ErrUnauthorized) {
		return intervals.IntervalsDTO{}, false, nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return intervals.IntervalsDTO{}, false, err
	}
	return intervals.IntervalsDTO{}, false, err
}

func optionalPowerVsHR(ctx context.Context, client ExtendedMetricsClient, activityID string) (intervals.PowerVsHR, bool, error) {
	payload, err := client.GetActivityPowerVsHR(ctx, activityID)
	if err == nil {
		return payload, true, nil
	}
	if errors.Is(err, intervals.ErrNotFound) || errors.Is(err, intervals.ErrUnauthorized) {
		return intervals.PowerVsHR{}, false, nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return intervals.PowerVsHR{}, false, err
	}
	return intervals.PowerVsHR{}, false, err
}

func stravaUnavailableExtendedMetricsResponse(activityID string, activity intervals.Activity, includeFull bool, version string) extendedMetricsResponse {
	out := extendedMetricsResponse{ActivityID: firstNonEmpty(activity.ID, activityID), StravaImported: true, Unavailable: &unavailableReason{Reason: "strava_tos", Workaround: stravaWorkaround}, Meta: extendedMetricsMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull, ExtendedMetricUnits: extendedMetricUnits(), DroppedFields: droppedExtendedMetricFields}}
	if includeFull {
		out.Full = map[string]any{"activity": activity.Raw}
	}
	return out
}

func shapeExtendedMetrics(activityID string, activity intervals.Activity, dto intervals.IntervalsDTO, intervalsOK bool, powerVsHR intervals.PowerVsHR, powerVsHROK bool, includeFull bool, version string, unavailable []string) extendedMetricsResponse {
	metrics := extendedMetricsFromActivity(activity.Raw, powerVsHR, powerVsHROK)
	out := extendedMetricsResponse{ActivityID: firstNonEmpty(activity.ID, activityID), Metrics: &metrics, Meta: extendedMetricsMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull, ExtendedMetricUnits: extendedMetricUnits(), DroppedFields: droppedExtendedMetricFields, Partial: len(unavailable) > 0, UnavailableSources: unavailable}}
	if intervalsOK {
		out.Intervals = extendedIntervals(dto.ICUIntervals)
	}
	if includeFull {
		out.Full = map[string]any{"activity": activity.Raw}
		if intervalsOK {
			out.Full["intervals"] = dto.Raw
		}
		if powerVsHROK {
			out.Full["power_vs_hr"] = powerVsHR.Raw
		}
	}
	return out
}

func extendedMetricsFromActivity(raw map[string]any, powerVsHR intervals.PowerVsHR, powerVsHROK bool) extendedActivityMetrics {
	var out extendedActivityMetrics
	out.StrideLengthM = rawNumberPtr(raw, "average_stride")
	out.PWHR = firstNumberPtr(rawNumberPtr(raw, "icu_power_hr"), powerVsHR.PowerHR)
	out.CardiacDecouplingPercent = firstNumberPtr(rawNumberPtr(raw, "decoupling"), powerVsHR.Decoupling)
	out.AerobicDecouplingPercent = firstNumberPtr(rawNumberPtr(raw, "decoupling"), powerVsHR.Decoupling)
	out.PowerZoneDistributionSeconds = rawNumberSlice(raw, "icu_zone_times")
	if useGap, ok := rawBool(raw, "use_gap_zone_times"); ok && useGap {
		out.PaceZoneTimeSeconds = rawNumberSlice(raw, "gap_zone_times")
	} else {
		out.PaceZoneTimeSeconds = rawNumberSlice(raw, "pace_zone_times")
	}
	out.JoulesAboveFTPKJ = rawJoulesKJPtr(raw, "icu_joules_above_ftp")
	out.IntensityFactor = rawNumberPtr(raw, "icu_intensity")
	out.VariabilityIndex = rawNumberPtr(raw, "icu_variability_index")
	out.PolarizationIndex = rawNumberPtr(raw, "polarization_index")
	out.TRIMP = rawNumberPtr(raw, "trimp")
	out.StrainScore = rawNumberPtr(raw, "strain_score")
	out.HRLoad = rawNumberPtr(raw, "hr_load")
	out.PaceLoad = rawNumberPtr(raw, "pace_load")
	out.PowerLoad = rawNumberPtr(raw, "power_load")
	out.TrainingLoad = rawNumberPtr(raw, "icu_training_load")
	out.LeftRightBalancePercent = rawNumberPtr(raw, "avg_lr_balance")
	out.RPE = rawNumberPtr(raw, "icu_rpe")
	out.Feel = rawNumberPtr(raw, "feel")
	out.SessionRPE = rawNumberPtr(raw, "session_rpe")
	out.CompliancePercent = rawNumberPtr(raw, "compliance")
	out.DeviceName = rawString(raw, "device_name")
	if powerVsHROK && out.PWHR == nil {
		out.PWHR = powerVsHR.PowerHRZ2
	}
	return out
}

func extendedIntervals(rows []intervals.ActivityInterval) []extendedIntervalMetrics {
	out := []extendedIntervalMetrics{}
	for _, row := range rows {
		metric := extendedIntervalMetrics{IntervalID: anyString(row.Raw["id"])}
		metric.Label = rawString(row.Raw, "label")
		if metric.Label == "" {
			metric.Label = rawString(row.Raw, "name")
		}
		metric.DFAAlpha1 = rawNumberPtr(row.Raw, "average_dfa_a1")
		metric.WPrimeBalanceStartKJ = rawJoulesKJPtr(row.Raw, "wbal_start")
		metric.WPrimeBalanceEndKJ = rawJoulesKJPtr(row.Raw, "wbal_end")
		metric.JoulesAboveFTPKJ = rawJoulesKJPtr(row.Raw, "joules_above_ftp")
		metric.AerobicDecouplingPercent = rawNumberPtr(row.Raw, "decoupling")
		metric.LeftRightBalancePercent = rawNumberPtr(row.Raw, "avg_lr_balance")
		metric.StrideLengthM = rawNumberPtr(row.Raw, "average_stride")
		metric.StrainScore = rawNumberPtr(row.Raw, "strain_score")
		metric.TrainingLoad = rawNumberPtr(row.Raw, "training_load")
		if hasExtendedIntervalMetric(metric) {
			out = append(out, metric)
		}
	}
	return out
}

func hasExtendedIntervalMetric(row extendedIntervalMetrics) bool {
	return row.DFAAlpha1 != nil || row.WPrimeBalanceStartKJ != nil || row.WPrimeBalanceEndKJ != nil || row.JoulesAboveFTPKJ != nil || row.AerobicDecouplingPercent != nil || row.LeftRightBalancePercent != nil || row.StrideLengthM != nil || row.StrainScore != nil || row.TrainingLoad != nil
}

func rawNumberPtr(raw map[string]any, key string) *float64 {
	value, ok := rawNumber(raw, key)
	if !ok {
		return nil
	}
	return roundPtr(value)
}

func rawJoulesKJPtr(raw map[string]any, key string) *float64 {
	value, ok := rawNumber(raw, key)
	if !ok {
		return nil
	}
	return roundPtr(value / 1000)
}

func rawNumber(raw map[string]any, key string) (float64, bool) {
	value, ok := raw[key]
	if !ok || value == nil {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func rawNumberSlice(raw map[string]any, key string) []float64 {
	value, ok := raw[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []float64:
		return typed
	case []any:
		out := make([]float64, 0, len(typed))
		for _, item := range typed {
			if number, ok := numberFromAny(item); ok {
				out = append(out, number)
			}
		}
		return out
	default:
		return nil
	}
}

func numberFromAny(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func rawBool(raw map[string]any, key string) (bool, bool) {
	value, ok := raw[key]
	if !ok || value == nil {
		return false, false
	}
	parsed, ok := value.(bool)
	return parsed, ok
}

func rawString(raw map[string]any, key string) string {
	value, ok := raw[key]
	if !ok || value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func firstNumberPtr(values ...*float64) *float64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func extendedMetricUnits() map[string]string {
	return map[string]string{
		"stride_length_m":                 string(units.UnitM),
		"cardiac_decoupling_percent":      string(units.UnitPercent),
		"aerobic_decoupling_percent":      string(units.UnitPercent),
		"power_zone_distribution_seconds": string(units.UnitSecs),
		"pace_zone_time_seconds":          string(units.UnitSecs),
		"joules_above_ftp_kj":             string(units.UnitKJ),
		"w_prime_balance_start_kj":        string(units.UnitKJ),
		"w_prime_balance_end_kj":          string(units.UnitKJ),
		"left_right_balance_percent":      string(units.UnitPercent),
		"rpe":                             string(units.UnitRPE),
	}
}

func extendedMetricsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Intervals.icu activity ID to fetch extended metrics for."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream activity, interval, and power-vs-HR payloads. Default terse mode only returns upstream-exposed metrics and drops unavailable fields."},
	}}
}

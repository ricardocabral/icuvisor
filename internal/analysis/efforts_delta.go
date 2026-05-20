package analysis

import "math"

const (
	EffortFamilyPower     = "power"
	EffortFamilyHeartRate = "heart_rate"
	EffortFamilyPace      = "pace"
)

// EffortBucketValue is one current or baseline curve bucket value.
type EffortBucketValue struct {
	Bucket     int
	Value      *float64
	ActivityID string
}

// EffortsDeltaInput contains comparable effort buckets.
type EffortsDeltaInput struct {
	Sport      string
	Family     string
	UnitSystem string
	Current    []EffortBucketValue
	Baseline   []EffortBucketValue
}

// EffortDeltaRow is one unit-explicit efforts delta row.
type EffortDeltaRow struct {
	DurationSeconds                 int      `json:"duration_seconds,omitempty"`
	DistanceMeters                  int      `json:"distance_meters,omitempty"`
	CurrentPowerWatts               *float64 `json:"current_power_watts,omitempty"`
	BaselinePowerWatts              *float64 `json:"baseline_power_watts,omitempty"`
	AbsoluteDeltaWatts              *float64 `json:"absolute_delta_watts,omitempty"`
	CurrentHeartRateBPM             *float64 `json:"current_heart_rate_bpm,omitempty"`
	BaselineHeartRateBPM            *float64 `json:"baseline_heart_rate_bpm,omitempty"`
	AbsoluteDeltaBPM                *float64 `json:"absolute_delta_bpm,omitempty"`
	CurrentElapsedSeconds           *float64 `json:"current_elapsed_seconds,omitempty"`
	BaselineElapsedSeconds          *float64 `json:"baseline_elapsed_seconds,omitempty"`
	AbsoluteDeltaSeconds            *float64 `json:"absolute_delta_seconds,omitempty"`
	CurrentPaceSecondsPerKM         *float64 `json:"current_pace_seconds_per_km,omitempty"`
	BaselinePaceSecondsPerKM        *float64 `json:"baseline_pace_seconds_per_km,omitempty"`
	AbsoluteDeltaPaceSecondsPerKM   *float64 `json:"absolute_delta_pace_seconds_per_km,omitempty"`
	CurrentPaceSecondsPerMile       *float64 `json:"current_pace_seconds_per_mile,omitempty"`
	BaselinePaceSecondsPerMile      *float64 `json:"baseline_pace_seconds_per_mile,omitempty"`
	AbsoluteDeltaPaceSecondsPerMile *float64 `json:"absolute_delta_pace_seconds_per_mile,omitempty"`
	PercentDelta                    *float64 `json:"percent_delta,omitempty"`
	CurrentActivityID               string   `json:"current_activity_id,omitempty"`
	BaselineActivityID              string   `json:"baseline_activity_id,omitempty"`
	CurrentMissing                  bool     `json:"current_missing,omitempty"`
	BaselineMissing                 bool     `json:"baseline_missing,omitempty"`
}

// EffortsDeltaResult is the terse efforts-delta result.
type EffortsDeltaResult struct {
	Sport           string           `json:"sport"`
	Family          string           `json:"effort_family"`
	UnitSystem      string           `json:"unit_system,omitempty"`
	BetterDirection string           `json:"better_direction"`
	N               int              `json:"n"`
	Buckets         []EffortDeltaRow `json:"buckets"`
}

// ComputeEffortsDelta compares current and baseline curve bucket values.
func ComputeEffortsDelta(input EffortsDeltaInput) EffortsDeltaResult {
	result := EffortsDeltaResult{Sport: input.Sport, Family: input.Family, UnitSystem: input.UnitSystem, BetterDirection: effortBetterDirection(input.Family)}
	current := effortBucketMap(input.Current)
	baseline := effortBucketMap(input.Baseline)
	keys := orderedEffortKeys(input.Current, input.Baseline)
	for _, key := range keys {
		cur, curOK := current[key]
		base, baseOK := baseline[key]
		row := EffortDeltaRow{CurrentMissing: !curOK || cur.Value == nil, BaselineMissing: !baseOK || base.Value == nil}
		if input.Family == EffortFamilyPace {
			row.DistanceMeters = key
		} else {
			row.DurationSeconds = key
		}
		if curOK {
			row.CurrentActivityID = cur.ActivityID
		}
		if baseOK {
			row.BaselineActivityID = base.ActivityID
		}
		if curOK && baseOK && cur.Value != nil && base.Value != nil {
			result.N++
			fillEffortDeltaValues(&row, input.Family, input.UnitSystem, key, *cur.Value, *base.Value)
		} else {
			fillSingleEffortValues(&row, input.Family, input.UnitSystem, key, cur.Value, base.Value)
		}
		result.Buckets = append(result.Buckets, row)
	}
	return result
}

func fillEffortDeltaValues(row *EffortDeltaRow, family string, unitSystem string, bucket int, current float64, baseline float64) {
	fillSingleEffortValues(row, family, unitSystem, bucket, &current, &baseline)
	delta := current - baseline
	if baseline != 0 {
		row.PercentDelta = roundPtr(delta / math.Abs(baseline) * 100)
	}
	switch family {
	case EffortFamilyPower:
		row.AbsoluteDeltaWatts = roundPtr(delta)
	case EffortFamilyHeartRate:
		row.AbsoluteDeltaBPM = roundPtr(delta)
	case EffortFamilyPace:
		row.AbsoluteDeltaSeconds = roundPtr(delta)
		currentPaceKM, baselinePaceKM := paceSecondsPer(current, bucket, 1000), paceSecondsPer(baseline, bucket, 1000)
		currentPaceMI, baselinePaceMI := paceSecondsPer(current, bucket, 1609.344), paceSecondsPer(baseline, bucket, 1609.344)
		if unitSystem == "imperial" {
			row.AbsoluteDeltaPaceSecondsPerMile = roundPtr(currentPaceMI - baselinePaceMI)
		} else {
			row.AbsoluteDeltaPaceSecondsPerKM = roundPtr(currentPaceKM - baselinePaceKM)
		}
	}
}

func fillSingleEffortValues(row *EffortDeltaRow, family string, unitSystem string, bucket int, current *float64, baseline *float64) {
	switch family {
	case EffortFamilyPower:
		row.CurrentPowerWatts = roundOptional(current)
		row.BaselinePowerWatts = roundOptional(baseline)
	case EffortFamilyHeartRate:
		row.CurrentHeartRateBPM = roundOptional(current)
		row.BaselineHeartRateBPM = roundOptional(baseline)
	case EffortFamilyPace:
		row.CurrentElapsedSeconds = roundOptional(current)
		row.BaselineElapsedSeconds = roundOptional(baseline)
		if current != nil {
			if unitSystem == "imperial" {
				row.CurrentPaceSecondsPerMile = roundPtr(paceSecondsPer(*current, bucket, 1609.344))
			} else {
				row.CurrentPaceSecondsPerKM = roundPtr(paceSecondsPer(*current, bucket, 1000))
			}
		}
		if baseline != nil {
			if unitSystem == "imperial" {
				row.BaselinePaceSecondsPerMile = roundPtr(paceSecondsPer(*baseline, bucket, 1609.344))
			} else {
				row.BaselinePaceSecondsPerKM = roundPtr(paceSecondsPer(*baseline, bucket, 1000))
			}
		}
	}
}

func roundOptional(value *float64) *float64 {
	if value == nil {
		return nil
	}
	return roundPtr(*value)
}

func paceSecondsPer(elapsedSeconds float64, distanceMeters int, unitMeters float64) float64 {
	if distanceMeters <= 0 {
		return math.NaN()
	}
	return elapsedSeconds / (float64(distanceMeters) / unitMeters)
}

func effortBetterDirection(family string) string {
	switch family {
	case EffortFamilyPace:
		return "lower"
	case EffortFamilyHeartRate:
		return "contextual"
	default:
		return "higher"
	}
}

func effortBucketMap(values []EffortBucketValue) map[int]EffortBucketValue {
	out := make(map[int]EffortBucketValue, len(values))
	for _, value := range values {
		out[value.Bucket] = value
	}
	return out
}

func orderedEffortKeys(left, right []EffortBucketValue) []int {
	seen := map[int]bool{}
	keys := []int{}
	for _, values := range [][]EffortBucketValue{left, right} {
		for _, value := range values {
			if !seen[value.Bucket] {
				seen[value.Bucket] = true
				keys = append(keys, value.Bucket)
			}
		}
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

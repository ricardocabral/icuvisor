package analysis

import "math"

const (
	PolarizationStateOK                    = "ok"
	PolarizationStateUnavailable           = "unavailable"
	PolarizationStateUndefinedModerateZero = "undefined_moderate_zero"
	PolarizationStateUndefinedHighZero     = "undefined_high_zero"
)

// ZoneBalance summarizes an upstream zone-time distribution into three intensity buckets.
type ZoneBalance struct {
	LowSeconds      float64
	ModerateSeconds float64
	HighSeconds     float64
	TotalSeconds    float64
	LowShare        float64
	ModerateShare   float64
	HighShare       float64
	Index           *float64
	State           string
	Classification  string
}

// ComputeZoneBalance calculates low/moderate/high shares and the polarization index.
func ComputeZoneBalance(zones []float64) ZoneBalance {
	var out ZoneBalance
	for idx, value := range zones {
		if value <= 0 {
			continue
		}
		out.TotalSeconds += value
		switch {
		case idx <= 1:
			out.LowSeconds += value
		case idx == 2:
			out.ModerateSeconds += value
		default:
			out.HighSeconds += value
		}
	}
	if out.TotalSeconds <= 0 {
		out.State = PolarizationStateUnavailable
		out.Classification = "unclassified"
		return out
	}
	out.LowShare = out.LowSeconds / out.TotalSeconds
	out.ModerateShare = out.ModerateSeconds / out.TotalSeconds
	out.HighShare = out.HighSeconds / out.TotalSeconds
	out.State = PolarizationStateOK
	if out.ModerateShare == 0 {
		out.State = PolarizationStateUndefinedModerateZero
	} else if out.HighShare == 0 {
		out.State = PolarizationStateUndefinedHighZero
	} else {
		index := math.Log10((out.LowShare / out.ModerateShare) * (out.HighShare / out.ModerateShare) * 100)
		out.Index = &index
	}
	out.Classification = classifyZoneBalance(out)
	return out
}

func classifyZoneBalance(balance ZoneBalance) string {
	switch {
	case balance.LowShare >= 0.70 && balance.HighShare >= balance.ModerateShare:
		return "polarized"
	case balance.LowShare > balance.ModerateShare && balance.ModerateShare > balance.HighShare:
		return "pyramidal"
	case (balance.ModerateShare >= balance.LowShare || balance.ModerateShare >= balance.HighShare) && balance.ModerateShare >= 0.30:
		return "threshold"
	default:
		return "unclassified"
	}
}

// BaselineStats contains deterministic baseline comparison statistics.
type BaselineStats struct {
	BaselineMean   *float64
	BaselineStdDev *float64
	CurrentValue   *float64
	ZScore         *float64
	Status         string
	Reason         string
}

// ComputeBaselineStats calculates mean, sample standard deviation, and z-score.
func ComputeBaselineStats(baseline []float64, current []float64, minSamples int, sumCurrent bool) BaselineStats {
	if minSamples <= 0 {
		minSamples = MinBaselineSamples
	}
	if len(baseline) < minSamples {
		return BaselineStats{Status: "insufficient_sample", Reason: "not_enough_baseline_samples"}
	}
	baselineMean := computeMean(baseline)
	std := sampleStdDev(baseline, baselineMean)
	if len(current) == 0 {
		return BaselineStats{BaselineMean: &baselineMean, BaselineStdDev: &std, Status: "insufficient_current_sample", Reason: "no_current_samples"}
	}
	currentValue := computeMean(current)
	if sumCurrent {
		currentValue = computeSum(current)
	}
	if std == 0 {
		return BaselineStats{BaselineMean: &baselineMean, BaselineStdDev: &std, CurrentValue: &currentValue, Status: "insufficient_variance", Reason: "zero_baseline_variance"}
	}
	z := (currentValue - baselineMean) / std
	return BaselineStats{BaselineMean: &baselineMean, BaselineStdDev: &std, CurrentValue: &currentValue, ZScore: &z, Status: "ok"}
}

// InterpretBaselineZScore maps wellness z-scores to deterministic states.
func InterpretBaselineZScore(metric Metric, z *float64) string {
	if z == nil {
		return "not_interpreted"
	}
	switch metric {
	case "hrv", "hrv_sdnn", "sleep_secs", "sleep_score", "sleep_quality", "readiness", "feel", "mood", "motivation", "steps":
		if *z <= -1 {
			return "suppressed"
		}
		if *z >= 1 {
			return "elevated_beneficial"
		}
		return "typical"
	case "rhr", "avg_sleeping_hr", "soreness", "fatigue", "stress", "baevsky_si", "blood_glucose", "lactate":
		if *z >= 1 {
			return "elevated"
		}
		if *z <= -1 {
			return "suppressed_beneficial"
		}
		return "typical"
	default:
		return "not_interpreted"
	}
}

func computeMean(values []float64) float64 { return computeSum(values) / float64(len(values)) }

func computeSum(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}
	return total
}

func sampleStdDev(values []float64, meanValue float64) float64 {
	if len(values) < 2 {
		return 0
	}
	var ss float64
	for _, value := range values {
		d := value - meanValue
		ss += d * d
	}
	return math.Sqrt(ss / float64(len(values)-1))
}

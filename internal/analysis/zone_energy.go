package analysis

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	// ZoneEnergyMethod identifies the analyzer's timestamp-weighted integration rule.
	ZoneEnergyMethod = "left_endpoint_power_timestamp_integration"
	// ZoneEnergyFormulaRef identifies the pinned formula resource entry.
	ZoneEnergyFormulaRef = "icuvisor://analysis-formulas#power_zone_mechanical_work"
	// ZoneEnergyMaxIntervalSeconds prevents silent power interpolation across large gaps.
	ZoneEnergyMaxIntervalSeconds = 60
	// ZoneEnergyInterpretation keeps mechanical work distinct from metabolic-energy claims.
	ZoneEnergyInterpretation = "Power-derived kJ is external mechanical work only; it is not metabolic energy, calorie expenditure, or food calories."
)

var (
	// ErrInvalidPowerZoneConfig indicates unusable configured power-zone boundaries.
	ErrInvalidPowerZoneConfig = errors.New("invalid power zone configuration")
)

// ZoneEnergyBoundaries are the ordered model-visible interpretation limits.
var ZoneEnergyBoundaries = []string{
	"Mechanical work from recorded power is not metabolic energy, calorie expenditure, or food calories.",
	"Left-endpoint integration; the final sample contributes no duration or work.",
	"Intervals longer than 60 seconds and invalid samples are skipped; missing power is not interpolated.",
	"Raw stream samples are never returned.",
}

// PowerZoneConfig is one athlete sport setting's ordered power-zone definition.
type PowerZoneConfig struct {
	Sport           string
	SportSettingID  int
	BoundariesWatts []float64
	Names           []string
}

// ZoneEnergyInput contains aligned elapsed-second and watt samples for one activity.
type ZoneEnergyInput struct {
	TimestampsSeconds []float64
	PowerWatts        []float64
	ZoneConfig        PowerZoneConfig
}

// ZoneEnergyDiagnostics reports sample coverage without imputing invalid data.
type ZoneEnergyDiagnostics struct {
	InputSamples              int `json:"input_samples"`
	AlignedSamples            int `json:"aligned_samples"`
	UsableIntervals           int `json:"usable_intervals"`
	SkippedIntervals          int `json:"skipped_intervals"`
	MisalignedSamples         int `json:"misaligned_samples"`
	SkippedNonFiniteTimestamp int `json:"skipped_non_finite_timestamp"`
	SkippedDuplicateTimestamp int `json:"skipped_duplicate_timestamp"`
	SkippedReversedTimestamp  int `json:"skipped_reversed_timestamp"`
	SkippedLargeGap           int `json:"skipped_large_gap"`
	SkippedNonFinitePower     int `json:"skipped_non_finite_power"`
	SkippedNegativePower      int `json:"skipped_negative_power"`
}

// ZoneEnergyZone reports time and mechanical work for one configured zone.
type ZoneEnergyZone struct {
	Zone        int      `json:"zone"`
	Name        string   `json:"name"`
	LowerWatts  float64  `json:"lower_watts"`
	UpperWatts  *float64 `json:"upper_watts,omitempty"`
	Seconds     float64  `json:"seconds"`
	KJ          float64  `json:"kj"`
	TimeShare   float64  `json:"time_share"`
	EnergyShare float64  `json:"energy_share"`
}

// ZoneEnergyResult reports timestamp-weighted mechanical work for one activity.
type ZoneEnergyResult struct {
	TotalSeconds float64               `json:"total_seconds"`
	TotalKJ      float64               `json:"total_kj"`
	Zones        []ZoneEnergyZone      `json:"zones,omitempty"`
	Diagnostics  ZoneEnergyDiagnostics `json:"diagnostics"`
}

// ValidatePowerZoneConfig rejects definition drift rather than sorting or repairing boundaries.
func ValidatePowerZoneConfig(config PowerZoneConfig) error {
	if len(config.BoundariesWatts) == 0 {
		return fmt.Errorf("%w: boundaries are required", ErrInvalidPowerZoneConfig)
	}
	for i, boundary := range config.BoundariesWatts {
		if math.IsNaN(boundary) || math.IsInf(boundary, 0) {
			return fmt.Errorf("%w: boundary %d must be finite", ErrInvalidPowerZoneConfig, i)
		}
		if boundary < 0 || (i > 0 && boundary <= 0) {
			return fmt.Errorf("%w: boundary %d must be positive except an initial zero", ErrInvalidPowerZoneConfig, i)
		}
		if i > 0 && boundary <= config.BoundariesWatts[i-1] {
			return fmt.Errorf("%w: boundaries must be strictly increasing", ErrInvalidPowerZoneConfig)
		}
	}
	return nil
}

// ZoneEnergyInputDiagnostics defines mismatch and short-input counter semantics before integration.
func ZoneEnergyInputDiagnostics(input ZoneEnergyInput) ZoneEnergyDiagnostics {
	powerN := len(input.PowerWatts)
	timeN := len(input.TimestampsSeconds)
	diagnostics := ZoneEnergyDiagnostics{
		InputSamples:   max(powerN, timeN),
		AlignedSamples: min(powerN, timeN),
	}
	if powerN != timeN {
		diagnostics.MisalignedSamples = absInt(powerN - timeN)
		diagnostics.SkippedIntervals = max(diagnostics.InputSamples-1, 0)
	}
	return diagnostics
}

func powerZoneName(config PowerZoneConfig, index int) string {
	if index >= 0 && index < len(config.Names) {
		if name := strings.TrimSpace(config.Names[index]); name != "" {
			return name
		}
	}
	return fmt.Sprintf("Zone %d", index+1)
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

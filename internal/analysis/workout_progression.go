package analysis

import (
	"encoding/json"
	"math"
	"sort"
	"strings"
)

const (
	// WorkoutProgressionMethod is the stable analyzer method identifier.
	WorkoutProgressionMethod = "ordered_workout_progression_evidence_v1"
	// WorkoutProgressionFormulaRef is the canonical formula resource reference.
	WorkoutProgressionFormulaRef = "icuvisor://analysis-formulas#workout_progression_evidence"
)

// WorkoutProgressionActivity is the normalized, read-only evidence for one requested activity.
type WorkoutProgressionActivity struct {
	ID                        string
	Label                     string
	Sport                     string
	Date                      string
	InitialReasons            []string
	Prescription              *WorkoutPrescription
	Completed                 *WorkoutCompleted
	DurationSeconds           *float64
	DurationSource            string
	RecoverySeconds           *float64
	Feel                      *float64
	RPE                       *float64
	UpstreamCompliancePercent *float64
	PowerHR                   *float64
	DecouplingPercent         *float64
	Readiness                 *ReadinessEvidence
}

// WorkoutPrescription is a validated planned interval structure.
type WorkoutPrescription struct {
	Source     string
	Intervals  []PrescriptionInterval
	Warnings   []PrescriptionWarning
	Saturated  bool
	TotalCount int
}

// PrescriptionWarning preserves a non-fatal workout document diagnostic.
type PrescriptionWarning struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	StepIndex *int   `json:"step_index,omitempty"`
}

// PrescriptionInterval is one normalized prescribed interval.
type PrescriptionInterval struct {
	DurationSeconds *float64
	DistanceMeters  *float64
	Recovery        bool
	Ramp            bool
	Freeride        bool
	Target          WorkoutTarget
}

// WorkoutTarget is a normalized target. Only Value and Min/Max are eligible for adherence.
type WorkoutTarget struct {
	Kind  string
	Unit  string
	Value *float64
	Min   *float64
	Max   *float64
	Start *float64
	End   *float64
	Text  string
}

// WorkoutCompleted is the normalized completed interval evidence.
type WorkoutCompleted struct {
	Intervals            []CompletedInterval
	IntervalSource       IntervalSource
	AutoLapSuspected     bool
	IntervalSourceCaveat string
}

// CompletedInterval is one normalized completed interval.
type CompletedInterval struct {
	DurationSeconds      *float64
	DistanceMeters       *float64
	Kind                 string
	ObservedPowerWatts   *float64
	ObservedHeartRateBPM *float64
	ObservedPace         *float64
	ObservedPaceUnit     string
}

// ReadinessEvidence contains source-labelled fields from one athlete-local wellness row.
type ReadinessEvidence struct {
	Date           string
	Fields         map[string]ReadinessField
	ExpectedFields []string
	Reasons        []string
}

// ReadinessField is one upstream wellness value and its provenance.
type ReadinessField struct {
	Value       any    `json:"value"`
	Source      string `json:"source"`
	NativeScale string `json:"native_scale"`
	FetchedAt   string `json:"fetched_at,omitempty"`
	Stale       bool   `json:"stale"`
}

// WorkoutProgressionResult is the deterministic progression evidence result.
type WorkoutProgressionResult struct {
	Status     string                    `json:"status"`
	Rows       []WorkoutProgressionRow   `json:"rows"`
	Deltas     []WorkoutProgressionDelta `json:"deltas"`
	Comparison WorkoutComparison         `json:"comparison"`
	Tolerances WorkoutTolerances         `json:"tolerances"`
}

// WorkoutProgressionRow is one ordered activity evidence row.
type WorkoutProgressionRow struct {
	ActivityID   string                `json:"activity_id"`
	Label        string                `json:"label,omitempty"`
	Sport        string                `json:"sport"`
	Date         string                `json:"date,omitempty"`
	Status       string                `json:"status"`
	Reasons      []string              `json:"reasons"`
	Prescription *PrescriptionEvidence `json:"prescription,omitempty"`
	Completed    *CompletedEvidence    `json:"completed,omitempty"`
	Extended     *ExtendedEvidence     `json:"extended,omitempty"`
	Duration     *DurationEvidence     `json:"duration,omitempty"`
	Recovery     *RecoveryEvidence     `json:"recovery,omitempty"`
	Adherence    *AdherenceEvidence    `json:"adherence,omitempty"`
	Stability    *StabilityEvidence    `json:"stability,omitempty"`
	Subjective   *SubjectiveEvidence   `json:"subjective,omitempty"`
	Readiness    *ReadinessOutput      `json:"readiness,omitempty"`
	Audit        *AuditEvidence        `json:"audit,omitempty"`
}

// PrescriptionEvidence is the serialized prescription summary.
type PrescriptionEvidence struct {
	Status               string                `json:"status"`
	SampleCount          int                   `json:"sample_count"`
	Reasons              []string              `json:"reasons"`
	Source               string                `json:"source,omitempty"`
	IntervalCount        int                   `json:"interval_count,omitempty"`
	TotalIntervalCount   int                   `json:"total_interval_count,omitempty"`
	PrescriptionWarnings []PrescriptionWarning `json:"prescription_warnings,omitempty"`
}

// CompletedEvidence is the serialized completed structure summary.
type CompletedEvidence struct {
	Status                  string   `json:"status"`
	SampleCount             int      `json:"sample_count"`
	Reasons                 []string `json:"reasons"`
	IntervalCount           int      `json:"interval_count,omitempty"`
	RecoveryCount           int      `json:"recovery_count,omitempty"`
	RecoveryDurationSeconds *float64 `json:"recovery_duration_seconds,omitempty"`
	IntervalSource          string   `json:"interval_source,omitempty"`
	AutoLapSuspected        bool     `json:"auto_lap_suspected"`
	IntervalSourceCaveat    string   `json:"interval_source_caveat,omitempty"`
}

// ExtendedEvidence is the bounded power-vs-heart-rate evidence summary.
type ExtendedEvidence struct {
	Status            string          `json:"status"`
	SampleCount       int             `json:"sample_count"`
	Reasons           []string        `json:"reasons"`
	PowerHR           *ScalarEvidence `json:"power_hr,omitempty"`
	DecouplingPercent *ScalarEvidence `json:"decoupling_percent,omitempty"`
}

// ScalarEvidence is one source-labelled scalar value.
type ScalarEvidence struct {
	Value  float64 `json:"value"`
	Unit   string  `json:"unit"`
	Source string  `json:"source"`
}

// DurationEvidence is a duration metric in seconds.
type DurationEvidence struct {
	Status       string   `json:"status"`
	SampleCount  int      `json:"sample_count"`
	Reasons      []string `json:"reasons"`
	ValueSeconds *float64 `json:"value_seconds,omitempty"`
	Source       string   `json:"source,omitempty"`
}

// RecoveryEvidence is recovery interval count and duration evidence.
type RecoveryEvidence struct {
	Status       string   `json:"status"`
	SampleCount  int      `json:"sample_count"`
	Reasons      []string `json:"reasons"`
	Count        int      `json:"count,omitempty"`
	ValueSeconds *float64 `json:"value_seconds,omitempty"`
}

// StabilityEvidence contains factual per-family means and half-to-half drift.
type StabilityEvidence struct {
	Status      string                     `json:"status"`
	SampleCount int                        `json:"sample_count"`
	Reasons     []string                   `json:"reasons"`
	Metrics     map[string]StabilityMetric `json:"metrics,omitempty"`
}

// StabilityMetric is one completed-observation family.
type StabilityMetric struct {
	Status       string   `json:"status"`
	SampleCount  int      `json:"sample_count"`
	Reasons      []string `json:"reasons"`
	Unit         string   `json:"unit"`
	Mean         float64  `json:"mean"`
	Min          float64  `json:"min"`
	Max          float64  `json:"max"`
	DriftPercent *float64 `json:"drift_percent,omitempty"`
}

// AdherenceEvidence is activity-level upstream compliance plus per-target-family evidence.
type AdherenceEvidence struct {
	Status                    string                     `json:"status"`
	SampleCount               int                        `json:"sample_count"`
	Reasons                   []string                   `json:"reasons"`
	UpstreamCompliancePercent *float64                   `json:"upstream_compliance_percent,omitempty"`
	Families                  map[string]AdherenceFamily `json:"families,omitempty"`
}

// AdherenceFamily is a target-family aggregate with independent absolute/relative availability.
type AdherenceFamily struct {
	Status                   string   `json:"status"`
	SampleCount              int      `json:"sample_count"`
	Reasons                  []string `json:"reasons"`
	EligibleCount            int      `json:"eligible_count"`
	WithinToleranceCount     int      `json:"within_tolerance_count"`
	MeanAbsoluteError        float64  `json:"mean_absolute_error"`
	MeanRelativeErrorPercent *float64 `json:"mean_relative_error_percent,omitempty"`
	Unit                     string   `json:"unit"`
}

// SubjectiveEvidence preserves source scales without deriving a score.
type SubjectiveEvidence struct {
	Status      string           `json:"status"`
	SampleCount int              `json:"sample_count"`
	Reasons     []string         `json:"reasons"`
	Feel        *SubjectiveValue `json:"feel,omitempty"`
	RPE         *SubjectiveValue `json:"rpe,omitempty"`
}

// SubjectiveValue is a source-labelled activity scalar.
type SubjectiveValue struct {
	Value  float64 `json:"value"`
	Scale  string  `json:"scale"`
	Source string  `json:"source"`
}

// ReadinessOutput is the serialized optional wellness context.
type ReadinessOutput struct {
	Status      string                    `json:"status"`
	SampleCount int                       `json:"sample_count"`
	Reasons     []string                  `json:"reasons"`
	Date        string                    `json:"date,omitempty"`
	Fields      map[string]ReadinessField `json:"fields,omitempty"`
}

// AuditEvidence contains only bounded normalized interval rows.
type AuditEvidence struct {
	Prescription           []NormalizedPrescriptionInterval `json:"prescription,omitempty"`
	Completed              []NormalizedCompletedInterval    `json:"completed,omitempty"`
	TotalPrescriptionCount int                              `json:"total_prescription_count"`
	TotalCompletedCount    int                              `json:"total_completed_count"`
	AuditTruncated         bool                             `json:"audit_truncated"`
	AuditTruncatedReason   string                           `json:"audit_truncated_reason,omitempty"`
}

// NormalizedPrescriptionInterval is the safe full-mode prescription audit row.
type NormalizedPrescriptionInterval struct {
	Index           int      `json:"index"`
	DurationSeconds *float64 `json:"duration_seconds,omitempty"`
	DistanceMeters  *float64 `json:"distance_meters,omitempty"`
	Kind            string   `json:"kind"`
	Recovery        bool     `json:"recovery"`
	TargetKind      string   `json:"target_kind,omitempty"`
	TargetUnit      string   `json:"target_unit,omitempty"`
}

// NormalizedCompletedInterval is the safe full-mode completed audit row.
type NormalizedCompletedInterval struct {
	Index                int      `json:"index"`
	DurationSeconds      *float64 `json:"duration_seconds,omitempty"`
	DistanceMeters       *float64 `json:"distance_meters,omitempty"`
	Kind                 string   `json:"kind"`
	Recovery             bool     `json:"recovery"`
	TargetKind           string   `json:"target_kind,omitempty"`
	TargetUnit           string   `json:"target_unit,omitempty"`
	ObservedPowerWatts   *float64 `json:"observed_power_watts,omitempty"`
	ObservedHeartRateBPM *float64 `json:"observed_heart_rate_bpm,omitempty"`
	ObservedPace         *float64 `json:"observed_pace,omitempty"`
	ObservedPaceUnit     string   `json:"observed_pace_unit,omitempty"`
}

// WorkoutProgressionDelta is an adjacent factual delta row.
type WorkoutProgressionDelta struct {
	FromActivityID       string            `json:"from_activity_id"`
	ToActivityID         string            `json:"to_activity_id"`
	Status               string            `json:"status"`
	Reasons              []string          `json:"reasons"`
	FieldStatus          map[string]string `json:"field_status"`
	DurationSecondsDelta *float64          `json:"duration_seconds_delta,omitempty"`
	DurationPercentDelta *float64          `json:"duration_percent_delta,omitempty"`
	RecoverySecondsDelta *float64          `json:"recovery_seconds_delta,omitempty"`
	RecoveryPercentDelta *float64          `json:"recovery_percent_delta,omitempty"`
	FeelDelta            *float64          `json:"feel_delta,omitempty"`
	RPEDelta             *float64          `json:"rpe_delta,omitempty"`
	SampleCount          int               `json:"sample_count"`
}

// WorkoutComparison contains explicit pair counts.
type WorkoutComparison struct {
	RequestedActivityCount int `json:"requested_activity_count"`
	ComparablePairCount    int `json:"comparable_pair_count"`
}

// WorkoutTolerances documents factual comparison units and target tolerance.
type WorkoutTolerances struct {
	TargetRelativePercent float64 `json:"target_relative_percent"`
	TargetAbsoluteFloor   float64 `json:"target_absolute_floor"`
	DurationDeltaUnit     string  `json:"duration_delta_unit"`
	RecoveryDeltaUnit     string  `json:"recovery_delta_unit"`
}

// AnalyzeWorkoutProgression computes the deterministic report from normalized evidence.
func AnalyzeWorkoutProgression(input []WorkoutProgressionActivity, includeFull bool) WorkoutProgressionResult {
	rows := make([]WorkoutProgressionRow, len(input))
	for i := range input {
		rows[i] = analyzeWorkoutRow(input[i], includeFull)
	}
	result := WorkoutProgressionResult{
		Rows:       rows,
		Comparison: WorkoutComparison{RequestedActivityCount: len(input)},
		Tolerances: WorkoutTolerances{TargetRelativePercent: 10, TargetAbsoluteFloor: 1, DurationDeltaUnit: "seconds", RecoveryDeltaUnit: "seconds"},
	}
	for i := 1; i < len(input); i++ {
		delta, comparable := analyzeWorkoutDelta(input[i-1], input[i], rows[i-1], rows[i])
		if comparable {
			result.Comparison.ComparablePairCount++
		}
		result.Deltas = append(result.Deltas, delta)
	}
	result.Status = "ok"
	for _, row := range result.Rows {
		if row.Status != "ok" {
			result.Status = "insufficient_evidence"
			break
		}
	}
	for _, delta := range result.Deltas {
		if delta.Status == "not_comparable" {
			result.Status = "not_comparable"
			break
		}
		if delta.Status != "ok" {
			result.Status = "insufficient_evidence"
		}
	}
	return result
}

func analyzeWorkoutRow(in WorkoutProgressionActivity, includeFull bool) WorkoutProgressionRow {
	row := WorkoutProgressionRow{ActivityID: strings.TrimSpace(in.ID), Label: strings.TrimSpace(in.Label), Sport: normalizeWorkoutSport(in.Sport), Date: strings.TrimSpace(in.Date), Status: "ok", Reasons: append([]string{}, in.InitialReasons...)}
	if row.Sport == "unknown" {
		addReason(&row.Reasons, "unknown_sport")
	}
	row.Prescription = prescriptionEvidence(in.Prescription)
	row.Completed = completedEvidence(in.Completed)
	row.Duration = durationEvidence(in.DurationSeconds, in.DurationSource)
	recoverySeconds := in.RecoverySeconds
	if recoverySeconds == nil && in.Completed != nil {
		recoverySeconds = completedRecoverySeconds(in.Completed)
	}
	row.Recovery = recoveryEvidence(in.Completed, recoverySeconds)
	row.Extended = extendedEvidence(in.PowerHR, in.DecouplingPercent)
	row.Subjective = subjectiveEvidence(in.Feel, in.RPE)
	row.Readiness = readinessOutput(in.Readiness)
	row.Stability = stabilityEvidence(in.Completed)
	row.Adherence = adherenceEvidence(in.Prescription, in.Completed, in.UpstreamCompliancePercent)
	if includeFull {
		row.Audit = auditEvidence(in.Prescription, in.Completed)
	}
	if in.Prescription == nil || in.Prescription.TotalCount == 0 && len(in.Prescription.Intervals) == 0 {
		addReason(&row.Reasons, "missing_prescription")
	}
	if in.Prescription != nil && in.Prescription.Saturated {
		addReason(&row.Reasons, "repeat_expansion_bounded")
	}
	if in.Completed == nil || len(in.Completed.Intervals) == 0 {
		addReason(&row.Reasons, "missing_intervals")
	} else if in.Completed.IntervalSource != IntervalSourceStructuredWorkout || in.Completed.AutoLapSuspected {
		addReason(&row.Reasons, "execution_source_unverified")
	}
	if row.Duration.Status != "ok" {
		addReason(&row.Reasons, "missing_metric")
	}
	if row.Recovery == nil || row.Recovery.Status != "ok" {
		addReason(&row.Reasons, "missing_metric")
	}
	if row.Extended.Status != "ok" {
		addReason(&row.Reasons, "missing_metric")
	}
	if row.Stability != nil && row.Stability.Status != "ok" {
		for _, reason := range row.Stability.Reasons {
			addReason(&row.Reasons, reason)
		}
	}
	if row.Subjective.Status != "ok" {
		addReason(&row.Reasons, "missing_subjective")
	}
	if in.Readiness != nil && row.Readiness.Status != "ok" {
		addReason(&row.Reasons, "missing_readiness")
	}
	if row.Adherence.Status != "ok" {
		for _, reason := range row.Adherence.Reasons {
			if reason == "missing_metric" || reason == "target_unsupported" || reason == "unit_incompatible" || reason == "structure_mismatch" || reason == "execution_source_unverified" || reason == "zero_denominator" {
				addReason(&row.Reasons, reason)
			}
		}
	}
	row.Reasons = sortedReasons(row.Reasons)
	if len(row.Reasons) > 0 {
		row.Status = "insufficient_evidence"
	}
	return row
}

func prescriptionEvidence(p *WorkoutPrescription) *PrescriptionEvidence {
	if p == nil {
		return nil
	}
	reasons := []string{}
	status := "ok"
	if p.Saturated {
		status = "insufficient_evidence"
		addReason(&reasons, "repeat_expansion_bounded")
	}
	if len(p.Intervals) == 0 && !p.Saturated {
		status = "insufficient_evidence"
		addReason(&reasons, "missing_prescription")
	}
	return &PrescriptionEvidence{Status: status, SampleCount: len(p.Intervals), Reasons: sortedReasons(reasons), Source: p.Source, IntervalCount: len(p.Intervals), TotalIntervalCount: totalIntervalCount(p), PrescriptionWarnings: p.Warnings}
}

func completedEvidence(c *WorkoutCompleted) *CompletedEvidence {
	if c == nil {
		return nil
	}
	reasons := []string{}
	status := "ok"
	if len(c.Intervals) == 0 {
		status = "insufficient_evidence"
		addReason(&reasons, "missing_intervals")
	}
	for _, interval := range c.Intervals {
		if strings.TrimSpace(interval.Kind) == "" {
			status = "insufficient_evidence"
			addReason(&reasons, "completed_kind_unknown")
		}
	}
	if c.IntervalSource != IntervalSourceStructuredWorkout || c.AutoLapSuspected {
		status = "insufficient_evidence"
		addReason(&reasons, "execution_source_unverified")
	}
	recoveryCount := 0
	var recoverySeconds float64
	recoveryFinite := true
	for _, interval := range c.Intervals {
		if isRecoveryKind(interval.Kind) {
			recoveryCount++
			if interval.DurationSeconds == nil || !wpFinite(*interval.DurationSeconds) || *interval.DurationSeconds < 0 {
				recoveryFinite = false
			} else {
				recoverySeconds += *interval.DurationSeconds
			}
		}
	}
	var recovery *float64
	if recoveryCount > 0 && recoveryFinite {
		recovery = roundedPtr(recoverySeconds)
	}
	if recoveryCount == 0 {
		addReason(&reasons, "missing_metric")
	} else if recovery == nil {
		addReason(&reasons, "missing_metric")
	}
	return &CompletedEvidence{Status: status, SampleCount: len(c.Intervals), Reasons: sortedReasons(reasons), IntervalCount: len(c.Intervals), RecoveryCount: recoveryCount, RecoveryDurationSeconds: recovery, IntervalSource: string(c.IntervalSource), AutoLapSuspected: c.AutoLapSuspected, IntervalSourceCaveat: c.IntervalSourceCaveat}
}

func durationEvidence(value *float64, source string) *DurationEvidence {
	if value == nil || !wpFinite(*value) || *value < 0 {
		return &DurationEvidence{Status: "insufficient_evidence", SampleCount: 0, Reasons: []string{"missing_metric"}}
	}
	return &DurationEvidence{Status: "ok", SampleCount: 1, ValueSeconds: roundedPtr(*value), Source: source}
}

func completedRecoverySeconds(c *WorkoutCompleted) *float64 {
	if c == nil {
		return nil
	}
	count := 0
	total := 0.0
	for _, interval := range c.Intervals {
		if !isRecoveryKind(interval.Kind) {
			continue
		}
		count++
		if interval.DurationSeconds == nil || !wpFinite(*interval.DurationSeconds) || *interval.DurationSeconds < 0 {
			return nil
		}
		total += *interval.DurationSeconds
	}
	if count == 0 {
		return nil
	}
	return roundedPtr(total)
}

func recoveryEvidence(c *WorkoutCompleted, value *float64) *RecoveryEvidence {
	if c == nil {
		return nil
	}
	count := 0
	for _, interval := range c.Intervals {
		if isRecoveryKind(interval.Kind) {
			count++
		}
	}
	if count == 0 || value == nil || !wpFinite(*value) || *value < 0 {
		return &RecoveryEvidence{Status: "insufficient_evidence", SampleCount: 0, Reasons: []string{"missing_metric"}, Count: count}
	}
	return &RecoveryEvidence{Status: "ok", SampleCount: count, Reasons: []string{}, Count: count, ValueSeconds: roundedPtr(*value)}
}

func extendedEvidence(powerHR, decoupling *float64) *ExtendedEvidence {
	out := &ExtendedEvidence{Status: "insufficient_evidence", SampleCount: 0, Reasons: []string{"missing_metric"}}
	if powerHR != nil && wpFinite(*powerHR) {
		out.PowerHR = &ScalarEvidence{Value: round(*powerHR), Unit: "power_per_heart_rate", Source: "power-vs-hr.json"}
	}
	if decoupling != nil && wpFinite(*decoupling) {
		out.DecouplingPercent = &ScalarEvidence{Value: round(*decoupling), Unit: "percent", Source: "power-vs-hr.json"}
	}
	if out.PowerHR != nil || out.DecouplingPercent != nil {
		out.Status = "ok"
		out.SampleCount = 1
		out.Reasons = nil
	}
	return out
}

func subjectiveEvidence(feel, rpe *float64) *SubjectiveEvidence {
	out := &SubjectiveEvidence{Status: "insufficient_evidence", Reasons: []string{"missing_subjective"}}
	if feel != nil && wpFinite(*feel) {
		out.Feel = &SubjectiveValue{Value: round(*feel), Scale: "1-5 athlete-reported feel", Source: "activity"}
	}
	if rpe != nil && wpFinite(*rpe) {
		out.RPE = &SubjectiveValue{Value: round(*rpe), Scale: "1-10 native icu_rpe", Source: "activity"}
	}
	if out.Feel != nil || out.RPE != nil {
		out.Status = "ok"
		out.SampleCount = 1
		out.Reasons = nil
	}
	return out
}

func readinessOutput(in *ReadinessEvidence) *ReadinessOutput {
	if in == nil {
		return nil
	}
	fields := map[string]ReadinessField{}
	for key, field := range in.Fields {
		fields[key] = field
	}
	status := "insufficient_evidence"
	reasons := sortedReasons(in.Reasons)
	if len(reasons) == 0 {
		reasons = []string{"missing_readiness"}
	}
	if len(fields) > 0 {
		if len(in.ExpectedFields) == 0 || len(fields) == len(in.ExpectedFields) {
			status = "ok"
			reasons = nil
		} else {
			status = "insufficient_evidence"
			addReason(&reasons, "missing_readiness")
		}
	}
	return &ReadinessOutput{Status: status, SampleCount: len(fields), Reasons: sortedReasons(reasons), Date: in.Date, Fields: fields}
}

func stabilityEvidence(c *WorkoutCompleted) *StabilityEvidence {
	out := &StabilityEvidence{Status: "insufficient_evidence", Reasons: []string{"missing_metric"}, Metrics: map[string]StabilityMetric{}}
	if c == nil {
		return out
	}
	families := map[string][]float64{}
	units := map[string]string{}
	for _, interval := range c.Intervals {
		if interval.ObservedPowerWatts != nil && wpFinite(*interval.ObservedPowerWatts) {
			families["power_watts"] = append(families["power_watts"], *interval.ObservedPowerWatts)
			units["power_watts"] = "W"
		}
		if interval.ObservedHeartRateBPM != nil && wpFinite(*interval.ObservedHeartRateBPM) {
			families["heart_rate_bpm"] = append(families["heart_rate_bpm"], *interval.ObservedHeartRateBPM)
			units["heart_rate_bpm"] = "BPM"
		}
		if interval.ObservedPace != nil && wpFinite(*interval.ObservedPace) {
			key := paceFamily(interval.ObservedPaceUnit)
			families[key] = append(families[key], *interval.ObservedPace)
			units[key] = strings.TrimSpace(interval.ObservedPaceUnit)
			if units[key] == "" {
				units[key] = "upstream_unspecified"
			}
		}
	}
	if len(families) == 0 {
		return out
	}
	out.Status = "ok"
	out.Reasons = nil
	for family, values := range families {
		metric := StabilityMetric{Status: "ok", SampleCount: len(values), Unit: units[family], Mean: round(wpMean(values)), Min: round(wpMinFloat(values)), Max: round(wpMaxFloat(values))}
		if len(values) < 4 {
			metric.Status = "insufficient_evidence"
			metric.Reasons = []string{"insufficient_evidence"}
			out.Status = "insufficient_evidence"
			addReason(&out.Reasons, "insufficient_evidence")
		} else {
			first := wpMean(values[:len(values)/2])
			second := wpMean(values[len(values)/2:])
			if first == 0 {
				metric.Reasons = []string{"zero_denominator"}
			} else {
				metric.DriftPercent = roundedPtr(100 * (second - first) / math.Abs(first))
			}
		}
		metric.Reasons = sortedReasons(metric.Reasons)
		out.Metrics[family] = metric
		out.SampleCount += len(values)
	}
	return out
}

func adherenceEvidence(p *WorkoutPrescription, c *WorkoutCompleted, upstream *float64) *AdherenceEvidence {
	out := &AdherenceEvidence{Status: "insufficient_evidence"}
	if upstream != nil && wpFinite(*upstream) {
		value := round(*upstream)
		out.UpstreamCompliancePercent = &value
	}
	if p == nil || c == nil {
		return out
	}
	if p.Saturated {
		out.Reasons = []string{"repeat_expansion_bounded"}
		return out
	}
	if len(p.Intervals) != len(c.Intervals) {
		out.Reasons = []string{"structure_mismatch"}
		return out
	}
	if c.IntervalSource != IntervalSourceStructuredWorkout || c.AutoLapSuspected {
		out.Reasons = []string{"execution_source_unverified"}
		return out
	}
	families := map[string]*adherenceAccumulator{}
	for index, prescribed := range p.Intervals {
		completed := c.Intervals[index]
		if strings.TrimSpace(completed.Kind) == "" {
			addReason(&out.Reasons, "completed_kind_unknown")
			continue
		}
		if isRecoveryKind(completed.Kind) != prescribed.Recovery {
			addReason(&out.Reasons, "structure_mismatch")
			continue
		}
		family, target, ok, reason := eligibleTarget(prescribed.Target, completed)
		if !ok {
			if reason != "" {
				addReason(&out.Reasons, reason)
			}
			continue
		}
		acc := families[family]
		if acc == nil {
			acc = &adherenceAccumulator{unit: target.unit}
			families[family] = acc
		}
		acc.eligible++
		errorValue, relative, relativeOK, within := targetError(target, completed)
		acc.absolute += errorValue
		if relativeOK {
			acc.relative = append(acc.relative, relative)
		}
		if targetHasZeroDenominator(target, completed) {
			acc.zeroDenominator = true
		}
		if within {
			acc.within++
		}
	}
	if containsReason(out.Reasons, "structure_mismatch") || containsReason(out.Reasons, "completed_kind_unknown") {
		out.Reasons = sortedReasons(out.Reasons)
		return out
	}
	if len(families) == 0 {
		if len(out.Reasons) == 0 {
			out.Reasons = []string{"target_unsupported"}
		}
		out.Reasons = sortedReasons(out.Reasons)
		return out
	}
	partialTarget := false
	for _, reason := range out.Reasons {
		if reason == "target_unsupported" || reason == "unit_incompatible" {
			partialTarget = true
		}
	}
	if partialTarget {
		out.Status = "insufficient_evidence"
	} else {
		out.Status = "ok"
	}
	out.Families = map[string]AdherenceFamily{}
	for family, acc := range families {
		out.SampleCount += acc.eligible
		familyOut := AdherenceFamily{Status: "ok", SampleCount: acc.eligible, Reasons: []string{}, EligibleCount: acc.eligible, WithinToleranceCount: acc.within, MeanAbsoluteError: round(acc.absolute / float64(acc.eligible)), Unit: acc.unit}
		if len(acc.relative) > 0 {
			value := round(wpMean(acc.relative))
			familyOut.MeanRelativeErrorPercent = &value
		}
		if acc.zeroDenominator {
			familyOut.Reasons = append(familyOut.Reasons, "zero_denominator")
			out.Status = "insufficient_evidence"
			addReason(&out.Reasons, "zero_denominator")
		}
		familyOut.Reasons = sortedReasons(familyOut.Reasons)
		out.Families[family] = familyOut
	}
	out.Reasons = sortedReasons(out.Reasons)
	return out
}

type adherenceAccumulator struct {
	unit            string
	eligible        int
	within          int
	absolute        float64
	relative        []float64
	zeroDenominator bool
}

type eligibleTargetValue struct {
	unit  string
	value *float64
	min   *float64
	max   *float64
}

func eligibleTarget(target WorkoutTarget, completed CompletedInterval) (string, eligibleTargetValue, bool, string) {
	kind := strings.ToLower(strings.TrimSpace(target.Kind))
	unit := strings.TrimSpace(target.Unit)
	if kind == "" {
		return "", eligibleTargetValue{}, false, ""
	}
	if kind != "power" && kind != "heart_rate" && kind != "pace" {
		return "", eligibleTargetValue{}, false, "target_unsupported"
	}
	if target.Value == nil && (target.Min == nil || target.Max == nil) {
		return "", eligibleTargetValue{}, false, "target_unsupported"
	}
	if target.Value != nil && !wpFinite(*target.Value) {
		return "", eligibleTargetValue{}, false, "target_unsupported"
	}
	if target.Min != nil && target.Max != nil && (!wpFinite(*target.Min) || !wpFinite(*target.Max) || *target.Min > *target.Max) {
		return "", eligibleTargetValue{}, false, "target_unsupported"
	}
	if kind == "power" {
		if !isPowerUnit(unit) || completed.ObservedPowerWatts == nil || !wpFinite(*completed.ObservedPowerWatts) {
			return "", eligibleTargetValue{}, false, "unit_incompatible"
		}
		return "power_watts", eligibleTargetValue{unit: "W", value: target.Value, min: target.Min, max: target.Max}, true, ""
	}
	if kind == "heart_rate" {
		if !isHRUnit(unit) || completed.ObservedHeartRateBPM == nil || !wpFinite(*completed.ObservedHeartRateBPM) {
			return "", eligibleTargetValue{}, false, "unit_incompatible"
		}
		return "heart_rate_bpm", eligibleTargetValue{unit: "BPM", value: target.Value, min: target.Min, max: target.Max}, true, ""
	}
	paceUnit := canonicalPaceUnit(unit)
	if paceUnit == "" || completed.ObservedPace == nil || !wpFinite(*completed.ObservedPace) || canonicalPaceUnit(completed.ObservedPaceUnit) != paceUnit {
		return "", eligibleTargetValue{}, false, "unit_incompatible"
	}
	return "pace_" + paceUnit, eligibleTargetValue{unit: paceUnit, value: target.Value, min: target.Min, max: target.Max}, true, ""
}

func targetHasZeroDenominator(target eligibleTargetValue, completed CompletedInterval) bool {
	if target.value != nil {
		return *target.value == 0
	}
	if target.min == nil || target.max == nil {
		return false
	}
	var observed float64
	switch target.unit {
	case "W":
		observed = *completed.ObservedPowerWatts
	case "BPM":
		observed = *completed.ObservedHeartRateBPM
	default:
		observed = *completed.ObservedPace
	}
	if observed >= *target.min && observed <= *target.max {
		return false
	}
	if observed < *target.min {
		return *target.min == 0
	}
	return *target.max == 0
}

func targetError(target eligibleTargetValue, completed CompletedInterval) (float64, float64, bool, bool) {
	var observed float64
	switch target.unit {
	case "W":
		observed = *completed.ObservedPowerWatts
	case "BPM":
		observed = *completed.ObservedHeartRateBPM
	default:
		observed = *completed.ObservedPace
	}
	var errorValue, denominator float64
	if target.value != nil {
		errorValue = math.Abs(observed - *target.value)
		denominator = math.Abs(*target.value)
	} else {
		if observed >= *target.min && observed <= *target.max {
			return 0, 0, false, true
		}
		if observed < *target.min {
			errorValue = *target.min - observed
			denominator = math.Abs(*target.min)
		} else {
			errorValue = observed - *target.max
			denominator = math.Abs(*target.max)
		}
	}
	if denominator == 0 {
		return errorValue, 0, false, errorValue <= 1
	}
	return errorValue, 100 * errorValue / denominator, true, errorValue <= math.Max(1, 0.1*denominator)
}

func analyzeWorkoutDelta(previous, current WorkoutProgressionActivity, previousRow, currentRow WorkoutProgressionRow) (WorkoutProgressionDelta, bool) {
	delta := WorkoutProgressionDelta{FromActivityID: previous.ID, ToActivityID: current.ID, Status: "ok", Reasons: []string{}, FieldStatus: map[string]string{}}
	for _, key := range []string{"duration_seconds", "duration_percent", "recovery_seconds", "recovery_percent", "feel", "rpe"} {
		delta.FieldStatus[key] = "missing"
	}
	prevSport := normalizeWorkoutSport(previous.Sport)
	currentSport := normalizeWorkoutSport(current.Sport)
	if prevSport == "unknown" || currentSport == "unknown" || prevSport != currentSport {
		delta.Status = "not_comparable"
		addReason(&delta.Reasons, "unknown_sport")
		for key := range delta.FieldStatus {
			delta.FieldStatus[key] = "not_comparable"
		}
		return delta, false
	}
	prevSignature, prevOK := prescriptionSignature(previous.Prescription)
	currentSignature, currentOK := prescriptionSignature(current.Prescription)
	if !prevOK || !currentOK {
		delta.Status = "insufficient_evidence"
		addReason(&delta.Reasons, "missing_prescription")
	} else if prevSignature != currentSignature {
		delta.Status = "not_comparable"
		addReason(&delta.Reasons, "structure_mismatch")
		for key := range delta.FieldStatus {
			delta.FieldStatus[key] = "not_comparable"
		}
		return delta, false
	}
	if previousRow.Status != "ok" || currentRow.Status != "ok" {
		delta.Status = "insufficient_evidence"
		for _, reason := range append(append([]string{}, previousRow.Reasons...), currentRow.Reasons...) {
			if reason != "unknown_sport" {
				addReason(&delta.Reasons, reason)
			}
		}
	}
	if value, ok := valueDelta(previous.DurationSeconds, current.DurationSeconds); ok {
		delta.DurationSecondsDelta = roundedPtr(value)
		delta.FieldStatus["duration_seconds"] = "ok"
		delta.SampleCount = 1
	} else {
		delta.FieldStatus["duration_seconds"] = "missing"
	}
	if previous.DurationSeconds != nil && current.DurationSeconds != nil && wpFinite(*previous.DurationSeconds) && wpFinite(*current.DurationSeconds) && *previous.DurationSeconds == 0 {
		delta.FieldStatus["duration_percent"] = "zero_denominator"
		addReason(&delta.Reasons, "zero_denominator")
	} else if percent, ok := percentDelta(previous.DurationSeconds, current.DurationSeconds); ok {
		delta.DurationPercentDelta = roundedPtr(percent)
		delta.FieldStatus["duration_percent"] = "ok"
		delta.SampleCount = 1
	}
	if value, ok := valueDelta(previous.RecoverySeconds, current.RecoverySeconds); ok {
		delta.RecoverySecondsDelta = roundedPtr(value)
		delta.FieldStatus["recovery_seconds"] = "ok"
		delta.SampleCount = 1
	}
	if previous.RecoverySeconds != nil && current.RecoverySeconds != nil && wpFinite(*previous.RecoverySeconds) && wpFinite(*current.RecoverySeconds) && *previous.RecoverySeconds == 0 {
		delta.FieldStatus["recovery_percent"] = "zero_denominator"
		addReason(&delta.Reasons, "zero_denominator")
	} else if percent, ok := percentDelta(previous.RecoverySeconds, current.RecoverySeconds); ok {
		delta.RecoveryPercentDelta = roundedPtr(percent)
		delta.FieldStatus["recovery_percent"] = "ok"
		delta.SampleCount = 1
	}
	if value, ok := valueDelta(previous.Feel, current.Feel); ok {
		delta.FeelDelta = roundedPtr(value)
		delta.FieldStatus["feel"] = "ok"
		delta.SampleCount = 1
	}
	if value, ok := valueDelta(previous.RPE, current.RPE); ok {
		delta.RPEDelta = roundedPtr(value)
		delta.FieldStatus["rpe"] = "ok"
		delta.SampleCount = 1
	}
	delta.Reasons = sortedReasons(delta.Reasons)
	return delta, prevOK && currentOK
}

func auditEvidence(p *WorkoutPrescription, c *WorkoutCompleted) *AuditEvidence {
	out := &AuditEvidence{TotalPrescriptionCount: totalIntervalCount(p), TotalCompletedCount: lenCompleted(c)}
	if p != nil {
		for i, interval := range p.Intervals {
			if i >= 200 {
				out.AuditTruncated = true
				out.AuditTruncatedReason = "audit_truncated"
				break
			}
			out.Prescription = append(out.Prescription, NormalizedPrescriptionInterval{Index: i, DurationSeconds: roundedOptional(interval.DurationSeconds), DistanceMeters: roundedOptional(interval.DistanceMeters), Kind: prescriptionKind(interval), Recovery: interval.Recovery, TargetKind: interval.Target.Kind, TargetUnit: interval.Target.Unit})
		}
		if p.Saturated {
			out.AuditTruncated = true
			out.AuditTruncatedReason = "repeat_expansion_bounded"
		}
	}
	if c != nil {
		for i, interval := range c.Intervals {
			if i >= 200 {
				out.AuditTruncated = true
				if out.AuditTruncatedReason == "" {
					out.AuditTruncatedReason = "audit_truncated"
				}
				break
			}
			out.Completed = append(out.Completed, NormalizedCompletedInterval{Index: i, DurationSeconds: roundedOptional(interval.DurationSeconds), DistanceMeters: roundedOptional(interval.DistanceMeters), Kind: interval.Kind, Recovery: isRecoveryKind(interval.Kind), ObservedPowerWatts: roundedOptional(interval.ObservedPowerWatts), ObservedHeartRateBPM: roundedOptional(interval.ObservedHeartRateBPM), ObservedPace: roundedOptional(interval.ObservedPace), ObservedPaceUnit: interval.ObservedPaceUnit})
		}
	}
	return out
}

func prescriptionSignature(p *WorkoutPrescription) (string, bool) {
	if p == nil || p.Saturated || len(p.Intervals) == 0 {
		return "", false
	}
	var parts []string
	for _, interval := range p.Intervals {
		parts = append(parts, canonicalIntervalSignature(interval))
	}
	return strings.Join(parts, "|"), true
}

func canonicalIntervalSignature(interval PrescriptionInterval) string {
	return jsonString([]any{wpCanonicalOptional(interval.DurationSeconds), wpCanonicalOptional(interval.DistanceMeters), interval.Recovery, interval.Ramp, interval.Freeride, interval.Target.Kind, interval.Target.Unit, wpCanonicalOptional(interval.Target.Value), wpCanonicalOptional(interval.Target.Min), wpCanonicalOptional(interval.Target.Max), wpCanonicalOptional(interval.Target.Start), wpCanonicalOptional(interval.Target.End), strings.TrimSpace(interval.Target.Text)})
}

func jsonString(value any) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func totalIntervalCount(p *WorkoutPrescription) int {
	if p == nil {
		return 0
	}
	if p.Saturated {
		return 1001
	}
	if p.TotalCount > 0 {
		return p.TotalCount
	}
	return len(p.Intervals)
}

func lenCompleted(c *WorkoutCompleted) int {
	if c == nil {
		return 0
	}
	return len(c.Intervals)
}

func prescriptionKind(interval PrescriptionInterval) string {
	if interval.Recovery {
		return "recovery"
	}
	return "work"
}

func isRecoveryKind(kind string) bool {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "recovery", "rest", "cooldown", "cool_down":
		return true
	default:
		return false
	}
}

func normalizeWorkoutSport(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "virtualride" {
		return "ride"
	}
	if value == "" {
		return "unknown"
	}
	return value
}

func paceFamily(unit string) string {
	canonical := canonicalPaceUnit(unit)
	if canonical == "" {
		return "pace_upstream_unspecified"
	}
	return "pace_" + canonical
}

func canonicalPaceUnit(unit string) string {
	normalized := strings.ToUpper(strings.TrimSpace(unit))
	switch normalized {
	case "MINS_KM", "MIN/KM", "MINS/KM", "SEC/KM", "SECS_KM", "SECONDS_PER_KM":
		return "seconds_per_km"
	case "MINS_MILE", "MIN/MI", "MINS/MI", "SEC/MI", "SECS_MILE", "SECONDS_PER_MILE":
		return "seconds_per_mile"
	case "SECS_100M", "SECONDS_PER_100M":
		return "seconds_per_100m"
	case "SECS_100Y", "SECONDS_PER_100Y":
		return "seconds_per_100y"
	case "SECS_250M", "SECONDS_PER_250M":
		return "seconds_per_250m"
	case "SECS_400M", "SECONDS_PER_400M":
		return "seconds_per_400m"
	case "SECS_500M", "SECONDS_PER_500M":
		return "seconds_per_500m"
	default:
		return ""
	}
}

func isPowerUnit(unit string) bool {
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "W", "WATT", "WATTS":
		return true
	default:
		return false
	}
}

func isHRUnit(unit string) bool {
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "BPM", "BEATS_PER_MINUTE", "HEART_RATE_BPM":
		return true
	default:
		return false
	}
}

func valueDelta(previous, current *float64) (float64, bool) {
	if previous == nil || current == nil || !wpFinite(*previous) || !wpFinite(*current) {
		return 0, false
	}
	return *current - *previous, true
}

func percentDelta(previous, current *float64) (float64, bool) {
	if previous == nil || current == nil || !wpFinite(*previous) || !wpFinite(*current) || *previous <= 0 {
		return 0, false
	}
	return 100 * (*current - *previous) / *previous, true
}

func round(value float64) float64 {
	return math.Round(value*1e6) / 1e6
}

func roundedPtr(value float64) *float64 {
	value = round(value)
	return &value
}

func roundedOptional(value *float64) *float64 {
	if value == nil || !wpFinite(*value) {
		return nil
	}
	return roundedPtr(*value)
}

func wpCanonicalOptional(value *float64) any {
	if value == nil || !wpFinite(*value) {
		return nil
	}
	return *value
}

func wpFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func wpMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func wpMinFloat(values []float64) float64 {
	value := values[0]
	for _, candidate := range values[1:] {
		if candidate < value {
			value = candidate
		}
	}
	return value
}

func wpMaxFloat(values []float64) float64 {
	value := values[0]
	for _, candidate := range values[1:] {
		if candidate > value {
			value = candidate
		}
	}
	return value
}

func containsReason(reasons []string, target string) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}

func addReason(reasons *[]string, reason string) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return
	}
	for _, existing := range *reasons {
		if existing == reason {
			return
		}
	}
	*reasons = append(*reasons, reason)
}

func sortedReasons(reasons []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		if strings.TrimSpace(reason) == "" || seen[reason] {
			continue
		}
		seen[reason] = true
		out = append(out, reason)
	}
	sort.Strings(out)
	return out
}

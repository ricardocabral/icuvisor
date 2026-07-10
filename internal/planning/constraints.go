// Package planning provides a pure, deterministic constraint model and validator
// for the plan-filler scheduling domain. It validates candidate sessions against
// structured week, day, and slot constraints and computes reconciliation totals.
//
// This package contains no intervals.icu client calls, no calendar writes, no model
// inference, and no physiology classification. All inputs are caller-supplied structs
// with numeric fields; free-text instructions are never treated as hard constraints.
package planning

import "slices"

// SlotConstraint defines limits for one available training window within a day.
// Two slots are independent — a session must fit within exactly one slot and
// cannot span multiple slots.
type SlotConstraint struct {
	// MaxDurationMinutes caps session duration for this slot. Zero means uncapped.
	MaxDurationMinutes float64
	// MaxIndoorMinutes caps duration specifically for indoor sessions.
	// Zero means no indoor cap. Outdoor sessions are not affected by this field.
	MaxIndoorMinutes float64
	// AllowedSports restricts which sports may fill this slot.
	// Empty means any sport is allowed.
	AllowedSports []string
	// AllowedModes restricts which training modes may fill this slot.
	// Empty means any mode is allowed.
	AllowedModes []string
}

// DayConstraints defines training availability for one calendar day.
// A day that is absent from WeekConstraints.AvailableDays is considered unavailable.
type DayConstraints struct {
	// Date is the athlete-local calendar date in YYYY-MM-DD format.
	Date string
	// MaxSessionsPerDay is the upper bound on sessions placed on this day.
	// Zero means the day is effectively unavailable regardless of slot count.
	MaxSessionsPerDay int
	// MaxTotalDailyMinutes caps the combined duration of all sessions on this day.
	// Zero means uncapped. Two sessions of 45 min each produce a combined total of 90 min.
	MaxTotalDailyMinutes float64
	// Slots lists the independent training windows available on this day.
	// A candidate session must fit within exactly one slot; slots do not combine.
	Slots []SlotConstraint
}

// WeekConstraints encodes the planning parameters for one calendar week.
//
// Availability (AvailableDays) captures where sessions may be placed.
// RequestedSessionCount captures how many sessions the caller wants placed.
// These are separate concepts: having 5 available days does not imply 5 sessions
// are requested, and requesting 3 sessions does not create availability on days
// that are absent from AvailableDays.
type WeekConstraints struct {
	// WeekStartDate is the athlete-local Monday in YYYY-MM-DD format.
	WeekStartDate string
	// WeeklyTargetMinutes is the full-week training-time target (e.g. for a complete week).
	// For an in-progress week, use the original full target here and report
	// actual completed time in CompletedMinutes; the validator derives RemainingMinutes
	// from WeeklyTargetMinutes - CompletedMinutes - FixedMinutes.
	WeeklyTargetMinutes float64
	// WeeklyTargetLoad is the full-week training-load target (e.g. TSS, ATL points).
	WeeklyTargetLoad float64
	// CompletedMinutes is already-logged training time this week (read-only past data).
	// Callers must not redistribute or zero this to create headroom.
	CompletedMinutes float64
	// CompletedLoad is already-logged training load this week (read-only past data).
	CompletedLoad float64
	// FixedMinutes is committed future training time from locked events
	// (e.g. races, A-priority events, unavailable blocks). These reduce the
	// remaining scheduling budget without being candidates themselves.
	FixedMinutes float64
	// FixedLoad is committed future training load from locked events.
	FixedLoad float64
	// RequestedSessionCount is the number of sessions the caller wants placed
	// into available slots. It is a scheduling intent, not an availability cap —
	// it may legally exceed the number of available slots, which produces a warning.
	RequestedSessionCount int
	// AvailableDays lists the days within this week where sessions may be placed.
	// Days absent from this list are unavailable for scheduling.
	AvailableDays []DayConstraints
}

// CandidateSession describes a proposed training session to be validated.
type CandidateSession struct {
	// Date is the proposed athlete-local date in YYYY-MM-DD format.
	Date string
	// Sport identifies the training discipline (e.g. "Ride", "Run", "Swim").
	Sport string
	// Mode identifies the training mode (e.g. "EnduranceRide", "Intervals").
	Mode string
	// Indoor indicates an indoor trainer, treadmill, pool, or similar facility.
	Indoor bool
	// DurationMinutes is the proposed session length.
	DurationMinutes float64
	// Load is the proposed training load contribution (e.g. TSS, ATL points).
	Load float64
}

// ViolationCode identifies a hard constraint breach that blocks session placement.
type ViolationCode string

const (
	// ViolationDayUnavailable fires when the candidate date has no DayConstraints
	// in WeekConstraints.AvailableDays, or when MaxSessionsPerDay is zero.
	ViolationDayUnavailable ViolationCode = "day_unavailable"

	// ViolationDailySessionCount fires when adding the candidate would exceed MaxSessionsPerDay.
	ViolationDailySessionCount ViolationCode = "daily_session_count_exceeded"

	// ViolationDailyTimeExceeded fires when the candidate would push the combined daily
	// duration over DayConstraints.MaxTotalDailyMinutes.
	ViolationDailyTimeExceeded ViolationCode = "daily_time_exceeded"

	// ViolationSlotDuration fires when the candidate duration exceeds every available
	// slot's MaxDurationMinutes. Two 45-minute slots cannot accommodate a 95-minute session.
	ViolationSlotDuration ViolationCode = "slot_duration_exceeded"

	// ViolationIndoorDuration fires when an indoor candidate duration exceeds
	// a slot's MaxIndoorMinutes. Outdoor sessions with the same duration are not affected.
	ViolationIndoorDuration ViolationCode = "indoor_duration_exceeded"

	// ViolationSportNotAllowed fires when the candidate sport is excluded by
	// every available slot's AllowedSports list.
	ViolationSportNotAllowed ViolationCode = "sport_not_allowed"

	// ViolationModeNotAllowed fires when the candidate mode is excluded by
	// every available slot's AllowedModes list.
	ViolationModeNotAllowed ViolationCode = "mode_not_allowed"

	// ViolationWeeklyLoadOvershoot fires when the candidate load would push the
	// projected weekly load over the remaining load budget
	// (WeeklyTargetLoad - CompletedLoad - FixedLoad - prior candidate load).
	ViolationWeeklyLoadOvershoot ViolationCode = "weekly_load_overshoot"

	// ViolationWeeklyTimeOvershoot fires when the candidate duration would push the
	// projected weekly time over the remaining time budget.
	ViolationWeeklyTimeOvershoot ViolationCode = "weekly_time_overshoot"
)

// WarningCode identifies a soft constraint concern that does not block placement
// but deserves caller attention.
type WarningCode string

const (
	// WarnInfeasibleSessionCount fires when RequestedSessionCount exceeds the
	// total number of available slots across all available days. Sessions beyond
	// the slot count cannot be placed.
	WarnInfeasibleSessionCount WarningCode = "infeasible_session_count"

	// WarnInfeasibleLoad fires when the total candidate load is less than the
	// remaining weekly load target, meaning the target cannot be met with the
	// provided candidates.
	WarnInfeasibleLoad WarningCode = "infeasible_load"

	// WarnZeroRemainingLoad fires when the remaining load budget is zero or
	// negative, meaning completed and fixed events already meet or exceed the
	// weekly target. No additional load is needed.
	WarnZeroRemainingLoad WarningCode = "zero_remaining_load"
)

// Violation reports a hard constraint breach.
type Violation struct {
	Code    ViolationCode `json:"code"`
	Message string        `json:"message"`
	Field   string        `json:"field,omitempty"`
	Value   any           `json:"value,omitempty"`
}

// Warning reports a soft constraint concern.
type Warning struct {
	Code    WarningCode `json:"code"`
	Message string      `json:"message"`
	Field   string      `json:"field,omitempty"`
	Value   any         `json:"value,omitempty"`
}

// Reconciliation holds computed weekly time and load totals for a set of candidates.
// All fields are derived from WeekConstraints and caller-supplied candidates;
// no values are redistributed, inferred, or smoothed.
type Reconciliation struct {
	WeeklyTargetMinutes float64 `json:"weekly_target_minutes"`
	WeeklyTargetLoad    float64 `json:"weekly_target_load"`
	CompletedMinutes    float64 `json:"completed_minutes"`
	CompletedLoad       float64 `json:"completed_load"`
	FixedMinutes        float64 `json:"fixed_minutes"`
	FixedLoad           float64 `json:"fixed_load"`
	CandidateMinutes    float64 `json:"candidate_minutes"`
	CandidateLoad       float64 `json:"candidate_load"`
	// RemainingMinutes is WeeklyTargetMinutes - CompletedMinutes - FixedMinutes.
	// This is the scheduling budget for new sessions; negative when already over target.
	RemainingMinutes float64 `json:"remaining_minutes"`
	// RemainingLoad is WeeklyTargetLoad - CompletedLoad - FixedLoad.
	RemainingLoad float64 `json:"remaining_load"`
	// ProjectedMinutes is CompletedMinutes + FixedMinutes + CandidateMinutes.
	ProjectedMinutes float64 `json:"projected_minutes"`
	// ProjectedLoad is CompletedLoad + FixedLoad + CandidateLoad.
	ProjectedLoad float64 `json:"projected_load"`
}

// CandidateResult is the validation outcome for a single CandidateSession.
type CandidateResult struct {
	Candidate  CandidateSession `json:"candidate"`
	Valid      bool             `json:"valid"`
	Violations []Violation      `json:"violations"`
	Warnings   []Warning        `json:"warnings,omitempty"`
}

// BatchResult is the validation outcome for all candidate sessions in a week.
type BatchResult struct {
	// Results contains one CandidateResult per input candidate, in order.
	Results []CandidateResult `json:"results"`
	// Warnings contains week-level warnings that apply to the batch as a whole
	// rather than to any individual candidate.
	Warnings []Warning `json:"warnings,omitempty"`
	// Reconciliation holds the computed weekly totals for all candidates combined.
	Reconciliation Reconciliation `json:"reconciliation"`
}

// Reconcile computes weekly time and load totals from WeekConstraints and a set
// of candidates. It does not validate any constraints; call ValidateCandidates for
// full constraint checking.
func Reconcile(wc WeekConstraints, candidates []CandidateSession) Reconciliation {
	var candMin, candLoad float64
	for _, c := range candidates {
		candMin += c.DurationMinutes
		candLoad += c.Load
	}
	return buildReconciliation(wc, candMin, candLoad)
}

// ValidateCandidate validates a single candidate session against the week constraints,
// assuming it is the first (and only) session being considered for its date and no
// prior candidates have been processed in the current batch.
//
// For batch validation with per-day state tracking, use ValidateCandidates.
func ValidateCandidate(wc WeekConstraints, candidate CandidateSession) CandidateResult {
	return validateWithState(wc, 0, 0, 0, 0, candidate)
}

// ValidateCandidates validates all candidates in order, tracking per-day session counts,
// combined daily duration, and accumulated weekly load from prior candidates.
// Candidates are processed in the order given; position within the slice determines
// which session claims a slot when the daily cap would be exceeded.
func ValidateCandidates(wc WeekConstraints, candidates []CandidateSession) BatchResult {
	type dayState struct {
		sessions int
		minutes  float64
	}
	dayStates := map[string]*dayState{}

	var priorLoad, priorMinutes float64
	results := make([]CandidateResult, len(candidates))

	for i, candidate := range candidates {
		ds := dayStates[candidate.Date]
		if ds == nil {
			ds = &dayState{}
			dayStates[candidate.Date] = ds
		}

		result := validateWithState(wc, ds.sessions, ds.minutes, priorLoad, priorMinutes, candidate)
		results[i] = result

		// All proposed sessions occupy space regardless of validity,
		// so that position-based daily-count violations are deterministic.
		ds.sessions++
		ds.minutes += candidate.DurationMinutes
		priorLoad += candidate.Load
		priorMinutes += candidate.DurationMinutes
	}

	var weekWarnings []Warning

	// Warn when requested sessions cannot fit in available slots.
	totalSlots := availableSlotCount(wc)
	if wc.RequestedSessionCount > totalSlots {
		weekWarnings = append(weekWarnings, Warning{
			Code:    WarnInfeasibleSessionCount,
			Message: "requested session count exceeds total available slots across available days",
			Field:   "requested_session_count",
			Value:   wc.RequestedSessionCount,
		})
	}

	var candMin, candLoad float64
	for _, c := range candidates {
		candMin += c.DurationMinutes
		candLoad += c.Load
	}
	recon := buildReconciliation(wc, candMin, candLoad)

	// Warn when candidates cannot satisfy the remaining load target.
	if recon.RemainingLoad > 0 && candLoad < recon.RemainingLoad {
		weekWarnings = append(weekWarnings, Warning{
			Code:    WarnInfeasibleLoad,
			Message: "candidate load total is less than remaining weekly load target",
			Field:   "remaining_load",
			Value:   recon.RemainingLoad,
		})
	}

	return BatchResult{
		Results:        results,
		Warnings:       weekWarnings,
		Reconciliation: recon,
	}
}

// validateWithState is the internal implementation of single-candidate validation.
// sessionsAlreadyOnDay counts sessions already proposed for candidate.Date.
// dailyMinutesAlready is the total minutes already proposed for candidate.Date.
// priorLoad and priorMinutes are accumulated totals from candidates already processed.
func validateWithState(wc WeekConstraints, sessionsAlreadyOnDay int, dailyMinutesAlready float64, priorLoad float64, priorMinutes float64, candidate CandidateSession) CandidateResult {
	var violations []Violation
	var warnings []Warning

	// 1. Day availability.
	day, ok := findDay(wc.AvailableDays, candidate.Date)
	if !ok || day.MaxSessionsPerDay == 0 {
		violations = append(violations, Violation{
			Code:    ViolationDayUnavailable,
			Message: "session date is not available for scheduling",
			Field:   "date",
			Value:   candidate.Date,
		})
		return CandidateResult{
			Candidate:  candidate,
			Valid:      false,
			Violations: violations,
		}
	}

	// 2. Daily session count.
	if sessionsAlreadyOnDay >= day.MaxSessionsPerDay {
		violations = append(violations, Violation{
			Code:    ViolationDailySessionCount,
			Message: "maximum sessions per day already reached",
			Field:   "max_sessions_per_day",
			Value:   day.MaxSessionsPerDay,
		})
	}

	// 3. Combined daily duration.
	if day.MaxTotalDailyMinutes > 0 {
		newDailyTotal := dailyMinutesAlready + candidate.DurationMinutes
		if newDailyTotal > day.MaxTotalDailyMinutes {
			violations = append(violations, Violation{
				Code:    ViolationDailyTimeExceeded,
				Message: "combined daily training duration would exceed the daily cap",
				Field:   "max_total_daily_minutes",
				Value:   day.MaxTotalDailyMinutes,
			})
		}
	}

	// 4. Slot constraints (duration, indoor cap, sport, mode).
	if len(day.Slots) > 0 {
		slotViolations := checkSlotConstraints(day.Slots, candidate)
		violations = append(violations, slotViolations...)
	}

	// 5. Weekly remaining load/time.
	// Remaining is target minus already-committed (completed + fixed) and prior candidates.
	remainingLoad := wc.WeeklyTargetLoad - wc.CompletedLoad - wc.FixedLoad - priorLoad
	remainingMin := wc.WeeklyTargetMinutes - wc.CompletedMinutes - wc.FixedMinutes - priorMinutes

	if wc.WeeklyTargetLoad > 0 {
		if remainingLoad <= 0 {
			warnings = append(warnings, Warning{
				Code:    WarnZeroRemainingLoad,
				Message: "remaining weekly load budget is zero or negative; no additional load is needed",
				Field:   "remaining_load",
				Value:   remainingLoad,
			})
		} else if candidate.Load > remainingLoad {
			violations = append(violations, Violation{
				Code:    ViolationWeeklyLoadOvershoot,
				Message: "candidate load exceeds remaining weekly load budget",
				Field:   "weekly_target_load",
				Value:   remainingLoad,
			})
		}
	}

	if wc.WeeklyTargetMinutes > 0 && remainingMin > 0 && candidate.DurationMinutes > remainingMin {
		violations = append(violations, Violation{
			Code:    ViolationWeeklyTimeOvershoot,
			Message: "candidate duration exceeds remaining weekly time budget",
			Field:   "weekly_target_minutes",
			Value:   remainingMin,
		})
	}

	return CandidateResult{
		Candidate:  candidate,
		Valid:      len(violations) == 0,
		Violations: violations,
		Warnings:   warnings,
	}
}

// checkSlotConstraints checks whether the candidate fits in at least one slot.
// Returns nil if any slot can accommodate the candidate.
// Returns violations describing why no slot could accommodate the candidate.
func checkSlotConstraints(slots []SlotConstraint, candidate CandidateSession) []Violation {
	type slotReject struct {
		duration bool
		indoor   bool
		sport    bool
		mode     bool
	}

	anyFit := false
	var rejects []slotReject

	for _, slot := range slots {
		r := slotReject{}

		if slot.MaxDurationMinutes > 0 && candidate.DurationMinutes > slot.MaxDurationMinutes {
			r.duration = true
		}
		if candidate.Indoor && slot.MaxIndoorMinutes > 0 && candidate.DurationMinutes > slot.MaxIndoorMinutes {
			r.indoor = true
		}
		if len(slot.AllowedSports) > 0 && !slices.Contains(slot.AllowedSports, candidate.Sport) {
			r.sport = true
		}
		if len(slot.AllowedModes) > 0 && !slices.Contains(slot.AllowedModes, candidate.Mode) {
			r.mode = true
		}

		if !r.duration && !r.indoor && !r.sport && !r.mode {
			anyFit = true
			break
		}
		rejects = append(rejects, r)
	}

	if anyFit {
		return nil
	}

	// Collect unique violation codes from rejection reasons.
	// A code is included if any slot rejected the candidate for that reason —
	// together the violations explain why no slot could fit the candidate.
	seen := map[ViolationCode]struct{}{}
	for _, r := range rejects {
		if r.duration {
			seen[ViolationSlotDuration] = struct{}{}
		}
		if r.indoor {
			seen[ViolationIndoorDuration] = struct{}{}
		}
		if r.sport {
			seen[ViolationSportNotAllowed] = struct{}{}
		}
		if r.mode {
			seen[ViolationModeNotAllowed] = struct{}{}
		}
	}

	// Emit violations in a deterministic order.
	var violations []Violation
	if _, ok := seen[ViolationSlotDuration]; ok {
		violations = append(violations, Violation{
			Code:    ViolationSlotDuration,
			Message: "session duration exceeds the available slot duration cap",
			Field:   "duration_minutes",
			Value:   candidate.DurationMinutes,
		})
	}
	if _, ok := seen[ViolationIndoorDuration]; ok {
		violations = append(violations, Violation{
			Code:    ViolationIndoorDuration,
			Message: "indoor session duration exceeds the slot indoor cap",
			Field:   "duration_minutes",
			Value:   candidate.DurationMinutes,
		})
	}
	if _, ok := seen[ViolationSportNotAllowed]; ok {
		violations = append(violations, Violation{
			Code:    ViolationSportNotAllowed,
			Message: "session sport is not in the allowed sports list for this slot",
			Field:   "sport",
			Value:   candidate.Sport,
		})
	}
	if _, ok := seen[ViolationModeNotAllowed]; ok {
		violations = append(violations, Violation{
			Code:    ViolationModeNotAllowed,
			Message: "session mode is not in the allowed modes list for this slot",
			Field:   "mode",
			Value:   candidate.Mode,
		})
	}
	return violations
}

// buildReconciliation computes a Reconciliation from WeekConstraints and candidate totals.
func buildReconciliation(wc WeekConstraints, candMin, candLoad float64) Reconciliation {
	remainingMin := wc.WeeklyTargetMinutes - wc.CompletedMinutes - wc.FixedMinutes
	remainingLoad := wc.WeeklyTargetLoad - wc.CompletedLoad - wc.FixedLoad
	return Reconciliation{
		WeeklyTargetMinutes: wc.WeeklyTargetMinutes,
		WeeklyTargetLoad:    wc.WeeklyTargetLoad,
		CompletedMinutes:    wc.CompletedMinutes,
		CompletedLoad:       wc.CompletedLoad,
		FixedMinutes:        wc.FixedMinutes,
		FixedLoad:           wc.FixedLoad,
		CandidateMinutes:    candMin,
		CandidateLoad:       candLoad,
		RemainingMinutes:    remainingMin,
		RemainingLoad:       remainingLoad,
		ProjectedMinutes:    wc.CompletedMinutes + wc.FixedMinutes + candMin,
		ProjectedLoad:       wc.CompletedLoad + wc.FixedLoad + candLoad,
	}
}

// availableSlotCount returns the total number of schedulable slots across all available days.
func availableSlotCount(wc WeekConstraints) int {
	total := 0
	for _, day := range wc.AvailableDays {
		if day.MaxSessionsPerDay <= 0 {
			continue
		}
		slotCap := day.MaxSessionsPerDay
		if len(day.Slots) > 0 && len(day.Slots) < slotCap {
			slotCap = len(day.Slots)
		}
		total += slotCap
	}
	return total
}

// findDay returns the DayConstraints for the given date, if present.
func findDay(days []DayConstraints, date string) (DayConstraints, bool) {
	for _, d := range days {
		if d.Date == date {
			return d, true
		}
	}
	return DayConstraints{}, false
}

// Package planning provides a pure, deterministic constraint model and validator
// for the plan-filler scheduling domain. It validates candidate sessions against
// structured week, day, and slot constraints and computes reconciliation totals.
//
// This package contains no intervals.icu client calls, no calendar writes, no model
// inference, and no physiology classification. All inputs are caller-supplied structs
// with numeric fields; free-text instructions are never treated as hard constraints.
package planning

import (
	"fmt"
	"math"
	"slices"
	"time"
)

// SlotConstraint defines limits for one available training window within a day.
// Two slots are independent — a session must fit within exactly one slot and
// cannot span multiple slots. In ValidateCandidates, slot assignment uses maximum
// bipartite matching so feasible schedules are accepted regardless of candidate order.
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
// A day absent from WeekConstraints.AvailableDays is considered unavailable.
type DayConstraints struct {
	// Date is the athlete-local calendar date in YYYY-MM-DD format.
	Date string
	// MaxSessionsPerDay is the upper bound on sessions placed on this day.
	// Zero means the day is effectively unavailable regardless of slot count.
	MaxSessionsPerDay int
	// MaxTotalDailyMinutes caps the combined duration of all sessions on this day.
	// Zero means uncapped.
	MaxTotalDailyMinutes float64
	// Slots lists the independent training windows available on this day.
	// Each slot holds at most one session; assignment uses bipartite matching.
	Slots []SlotConstraint
}

// WeekConstraints encodes the planning parameters for one calendar week.
//
// Availability (AvailableDays) captures where sessions may be placed.
// RequestedSessionCount captures how many sessions the caller wants placed.
// These are separate concepts: having 5 available days does not imply 5 sessions
// are requested, and requesting 3 sessions does not create availability on days
// absent from AvailableDays.
//
// Nil pointer fields represent "not set / unlimited":
//   - WeeklyTargetMinutes == nil: no weekly time budget is tracked.
//   - WeeklyTargetLoad == nil: no weekly load budget is tracked.
//   - RequestedSessionCount == nil: no session-count cap is enforced.
//
// A pointer with value 0 represents an explicit hard zero (zero budget, zero sessions).
// Duplicate Date values in AvailableDays are invalid; use ValidateWeekConstraints.
type WeekConstraints struct {
	// WeekStartDate is the athlete-local Monday in YYYY-MM-DD format.
	WeekStartDate string
	// WeeklyTargetMinutes is the full-week training-time target.
	// Nil means no time budget tracking. Pointer to 0 means zero time budget.
	WeeklyTargetMinutes *float64
	// WeeklyTargetLoad is the full-week training-load target (e.g. TSS, ATL points).
	// Nil means no load budget tracking. Pointer to 0 means zero load budget.
	WeeklyTargetLoad *float64
	// CompletedMinutes is already-logged training time this week (read-only past data).
	CompletedMinutes float64
	// CompletedLoad is already-logged training load this week (read-only past data).
	CompletedLoad float64
	// FixedMinutes is committed future training time from locked events (races, etc.).
	FixedMinutes float64
	// FixedLoad is committed future training load from locked events.
	FixedLoad float64
	// RequestedSessionCount is the number of sessions the caller wants placed.
	// Nil means no session-count cap. Pointer to 0 means zero sessions are wanted
	// (all candidates are excess). Pointer to N enforces a cap of N valid sessions.
	RequestedSessionCount *int
	// AvailableDays lists the days within this week where sessions may be placed.
	// Duplicate Date values are invalid; call ValidateWeekConstraints before use.
	AvailableDays []DayConstraints
}

// CandidateSession describes a proposed training session to be validated.
// DurationMinutes and Load must be finite and non-negative; otherwise
// ViolationInvalidInput is returned without checking other constraints.
type CandidateSession struct {
	// Date is the proposed athlete-local date in YYYY-MM-DD format.
	Date string
	// Sport identifies the training discipline (e.g. "Ride", "Run", "Swim").
	Sport string
	// Mode identifies the training mode (e.g. "EnduranceRide", "Intervals").
	Mode string
	// Indoor indicates an indoor trainer, treadmill, pool, or similar facility.
	Indoor bool
	// DurationMinutes is the proposed session length. Must be finite and >= 0.
	DurationMinutes float64
	// Load is the proposed training load contribution (e.g. TSS, ATL points).
	// Must be finite and >= 0.
	Load float64
}

// ViolationCode identifies a hard constraint breach that blocks session placement.
type ViolationCode string

const (
	// ViolationInvalidInput fires when a candidate has non-finite or negative
	// DurationMinutes or Load. Checked before all other constraints.
	// The invalid field is replaced by 0 in the embedded Candidate for JSON-safe output.
	ViolationInvalidInput ViolationCode = "invalid_input"

	// ViolationDayUnavailable fires when the candidate date has no DayConstraints
	// in WeekConstraints.AvailableDays, or when MaxSessionsPerDay is zero.
	ViolationDayUnavailable ViolationCode = "day_unavailable"

	// ViolationDailySessionCount fires when adding the candidate would exceed MaxSessionsPerDay.
	ViolationDailySessionCount ViolationCode = "daily_session_count_exceeded"

	// ViolationDailyTimeExceeded fires when the candidate would push the combined daily
	// duration over DayConstraints.MaxTotalDailyMinutes.
	ViolationDailyTimeExceeded ViolationCode = "daily_time_exceeded"

	// ViolationSlotDuration fires when the candidate duration exceeds every available slot's
	// MaxDurationMinutes. Only emitted when ALL slots reject for this reason.
	ViolationSlotDuration ViolationCode = "slot_duration_exceeded"

	// ViolationIndoorDuration fires when an indoor candidate duration exceeds every slot's
	// MaxIndoorMinutes. Only emitted when ALL slots reject for this reason.
	ViolationIndoorDuration ViolationCode = "indoor_duration_exceeded"

	// ViolationSportNotAllowed fires when the candidate sport is excluded by every slot's
	// AllowedSports. Only emitted when ALL slots reject for this reason.
	ViolationSportNotAllowed ViolationCode = "sport_not_allowed"

	// ViolationModeNotAllowed fires when the candidate mode is excluded by every slot's
	// AllowedModes. Only emitted when ALL slots reject for this reason.
	ViolationModeNotAllowed ViolationCode = "mode_not_allowed"

	// ViolationNoCompatibleSlot fires when no slot can accommodate the candidate
	// and no single constraint reason is universal across all slots (mixed-reason
	// rejection, e.g. slot A rejects for duration, slot B rejects for sport).
	ViolationNoCompatibleSlot ViolationCode = "no_compatible_slot"

	// ViolationNoAvailableSlot fires when the candidate fits at least one slot's
	// constraints but all compatible slots are claimed by other candidates in the
	// bipartite matching (contention, not constraint failure).
	ViolationNoAvailableSlot ViolationCode = "no_available_slot"

	// ViolationWeeklyLoadOvershoot fires when the candidate would push projected load
	// over the remaining budget, including when the budget is already exhausted.
	// Fires for any candidate with Load > 0 when remainingLoad <= 0.
	ViolationWeeklyLoadOvershoot ViolationCode = "weekly_load_overshoot"

	// ViolationWeeklyTimeOvershoot fires when the candidate would push projected time
	// over the remaining budget, including when the budget is already exhausted.
	// Fires for any candidate with DurationMinutes > 0 when remainingMin <= 0.
	ViolationWeeklyTimeOvershoot ViolationCode = "weekly_time_overshoot"

	// ViolationRequestedSessionCountExceeded fires when the candidate would be the
	// (N+1)th valid session and RequestedSessionCount is pointer-to-N.
	// Position within the candidate batch determines which sessions are accepted.
	ViolationRequestedSessionCountExceeded ViolationCode = "requested_session_count_exceeded"
)

// WarningCode identifies a soft constraint concern.
type WarningCode string

const (
	// WarnInfeasibleSessionCount fires when *RequestedSessionCount exceeds the total
	// structural slot capacity across all available days. Structural capacity is
	// min(MaxSessionsPerDay, len(Slots)) per day; sport/mode filtering is not applied.
	WarnInfeasibleSessionCount WarningCode = "infeasible_session_count"

	// WarnRequestedSessionCountUnmet fires when the batch contains fewer valid sessions
	// than *RequestedSessionCount. The warning includes both the requested count and the
	// accepted count, so callers can detect underfilled schedules deterministically.
	WarnRequestedSessionCountUnmet WarningCode = "requested_session_count_unmet"

	// WarnInfeasibleLoad fires when the total load of valid-input candidates is less than
	// the remaining weekly load target. Invalid-input candidates (NaN/negative) are
	// excluded from this total.
	WarnInfeasibleLoad WarningCode = "infeasible_load"

	// WarnZeroRemainingLoad fires when the remaining load budget is zero or negative.
	// Fires unconditionally when remaining <= 0. When Load > 0, also accompanied by
	// ViolationWeeklyLoadOvershoot (hard block).
	WarnZeroRemainingLoad WarningCode = "zero_remaining_load"

	// WarnZeroRemainingTime fires when the remaining time budget is zero or negative.
	// Parallel to WarnZeroRemainingLoad. When DurationMinutes > 0, also accompanied
	// by ViolationWeeklyTimeOvershoot.
	WarnZeroRemainingTime WarningCode = "zero_remaining_time"
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

// Reconciliation holds computed weekly time and load totals.
// CandidateMinutes and CandidateLoad include only valid-input candidates;
// invalid-input candidates (NaN/negative) are excluded to ensure JSON-safe output.
//
// WeeklyTargetMinutes, WeeklyTargetLoad, RemainingMinutes, and RemainingLoad are
// pointer fields that mirror the nullable WeekConstraints targets:
//   - nil indicates the dimension was not tracked (WeeklyTarget* was nil in WeekConstraints).
//   - Non-nil with value 0 indicates an explicit zero budget.
//
// JSON output uses omitempty for these fields; absent keys indicate an untracked dimension.
type Reconciliation struct {
	// WeeklyTargetMinutes mirrors WeekConstraints.WeeklyTargetMinutes.
	// Nil when no time target was set.
	WeeklyTargetMinutes *float64 `json:"weekly_target_minutes,omitempty"`
	// WeeklyTargetLoad mirrors WeekConstraints.WeeklyTargetLoad.
	// Nil when no load target was set.
	WeeklyTargetLoad *float64 `json:"weekly_target_load,omitempty"`
	CompletedMinutes float64  `json:"completed_minutes"`
	CompletedLoad    float64  `json:"completed_load"`
	FixedMinutes     float64  `json:"fixed_minutes"`
	FixedLoad        float64  `json:"fixed_load"`
	// CandidateMinutes is the sum of DurationMinutes for valid-input candidates only.
	CandidateMinutes float64 `json:"candidate_minutes"`
	// CandidateLoad is the sum of Load for valid-input candidates only.
	CandidateLoad float64 `json:"candidate_load"`
	// RemainingMinutes is WeeklyTargetMinutes - CompletedMinutes - FixedMinutes.
	// Nil when WeeklyTargetMinutes is nil (dimension untracked).
	RemainingMinutes *float64 `json:"remaining_minutes,omitempty"`
	// RemainingLoad is WeeklyTargetLoad - CompletedLoad - FixedLoad.
	// Nil when WeeklyTargetLoad is nil (dimension untracked).
	RemainingLoad    *float64 `json:"remaining_load,omitempty"`
	ProjectedMinutes float64  `json:"projected_minutes"`
	ProjectedLoad    float64  `json:"projected_load"`
}

// CandidateResult is the validation outcome for a single CandidateSession.
type CandidateResult struct {
	// Candidate echoes the input. For invalid-input results, non-finite or negative
	// DurationMinutes/Load are replaced by 0 to ensure JSON-safe output.
	Candidate  CandidateSession `json:"candidate"`
	Valid      bool             `json:"valid"`
	Violations []Violation      `json:"violations"`
	Warnings   []Warning        `json:"warnings,omitempty"`
}

// BatchResult is the validation outcome for all candidate sessions in a week.
type BatchResult struct {
	Results        []CandidateResult `json:"results"`
	Warnings       []Warning         `json:"warnings,omitempty"`
	Reconciliation Reconciliation    `json:"reconciliation"`
}

// ValidateWeekConstraints checks a WeekConstraints struct for structural and numeric validity.
// Returns a non-nil error if any field is outside its valid domain. Errors are reported in
// a fixed deterministic order (date fields first, then numeric, then day entries).
func ValidateWeekConstraints(wc WeekConstraints) error {
	if wc.WeekStartDate == "" {
		return fmt.Errorf("week_start_date is required")
	}
	weekStart, err := time.Parse(time.DateOnly, wc.WeekStartDate)
	if err != nil {
		return fmt.Errorf("week_start_date must be YYYY-MM-DD, got %q", wc.WeekStartDate)
	}
	if weekStart.Weekday() != time.Monday {
		return fmt.Errorf("week_start_date must be a Monday, got %s (%s)", wc.WeekStartDate, weekStart.Weekday())
	}
	weekEnd := weekStart.AddDate(0, 0, 6)

	// Top-level numeric fields in fixed order for deterministic error reporting.
	type fieldCheck struct {
		name string
		val  float64
		set  bool
	}
	for _, fc := range []fieldCheck{
		{"weekly_target_minutes", derefFloat(wc.WeeklyTargetMinutes), wc.WeeklyTargetMinutes != nil},
		{"weekly_target_load", derefFloat(wc.WeeklyTargetLoad), wc.WeeklyTargetLoad != nil},
		{"completed_minutes", wc.CompletedMinutes, true},
		{"completed_load", wc.CompletedLoad, true},
		{"fixed_minutes", wc.FixedMinutes, true},
		{"fixed_load", wc.FixedLoad, true},
	} {
		if !fc.set {
			continue
		}
		if err := requireFiniteNonNegative(fc.name, fc.val); err != nil {
			return err
		}
	}
	if wc.RequestedSessionCount != nil && *wc.RequestedSessionCount < 0 {
		return fmt.Errorf("requested_session_count must be non-negative when set, got %d", *wc.RequestedSessionCount)
	}

	seen := map[string]struct{}{}
	for i, day := range wc.AvailableDays {
		dayDate, dayErr := time.Parse(time.DateOnly, day.Date)
		if dayErr != nil {
			return fmt.Errorf("available_days[%d].date must be YYYY-MM-DD, got %q", i, day.Date)
		}
		if dayDate.Before(weekStart) || dayDate.After(weekEnd) {
			return fmt.Errorf("available_days[%d].date %q is outside the week %s to %s",
				i, day.Date, wc.WeekStartDate, weekEnd.Format(time.DateOnly))
		}
		if day.MaxSessionsPerDay < 0 {
			return fmt.Errorf("available_days[%d] (%s): max_sessions_per_day must be non-negative, got %d", i, day.Date, day.MaxSessionsPerDay)
		}
		if err := requireFiniteNonNegative(fmt.Sprintf("available_days[%d].max_total_daily_minutes", i), day.MaxTotalDailyMinutes); err != nil {
			return err
		}
		if _, dup := seen[day.Date]; dup {
			return fmt.Errorf("available_days contains duplicate date %q", day.Date)
		}
		seen[day.Date] = struct{}{}
		for j, slot := range day.Slots {
			prefix := fmt.Sprintf("available_days[%d].slots[%d]", i, j)
			if err := requireFiniteNonNegative(prefix+".max_duration_minutes", slot.MaxDurationMinutes); err != nil {
				return err
			}
			if err := requireFiniteNonNegative(prefix+".max_indoor_minutes", slot.MaxIndoorMinutes); err != nil {
				return err
			}
		}
	}
	return nil
}

// Reconcile computes weekly time and load totals from WeekConstraints and candidates.
// Invalid-input candidates (NaN/negative) are excluded from CandidateMinutes and
// CandidateLoad to prevent non-finite values from propagating into the result.
func Reconcile(wc WeekConstraints, candidates []CandidateSession) Reconciliation {
	var candMin, candLoad float64
	for _, c := range candidates {
		if invalidCandidateInputViolation(c) != nil {
			continue
		}
		candMin += c.DurationMinutes
		candLoad += c.Load
	}
	return buildReconciliation(wc, candMin, candLoad)
}

// ValidateCandidate validates a single candidate session in isolation. All slots in the
// day are treated as available (no slot consumption or matching with other candidates).
// For batch validation with bipartite slot matching, use ValidateCandidates.
func ValidateCandidate(wc WeekConstraints, candidate CandidateSession) CandidateResult {
	if v := invalidCandidateInputViolation(candidate); v != nil {
		return CandidateResult{Candidate: sanitizeCandidateForResult(candidate), Valid: false, Violations: []Violation{*v}}
	}
	// A pointer-to-zero RequestedSessionCount means zero sessions are wanted;
	// this candidate is excess regardless of availability.
	if wc.RequestedSessionCount != nil && *wc.RequestedSessionCount == 0 {
		return CandidateResult{
			Candidate: candidate,
			Valid:     false,
			Violations: []Violation{{
				Code:    ViolationRequestedSessionCountExceeded,
				Message: "requested session count already reached; this candidate is excess",
				Field:   "requested_session_count",
				Value:   *wc.RequestedSessionCount,
			}},
		}
	}
	day, ok := findDay(wc.AvailableDays, candidate.Date)
	if !ok || day.MaxSessionsPerDay == 0 {
		return CandidateResult{
			Candidate: candidate,
			Valid:     false,
			Violations: []Violation{{
				Code:    ViolationDayUnavailable,
				Message: "session date is not available for scheduling",
				Field:   "date",
				Value:   candidate.Date,
			}},
		}
	}
	return validateCore(wc, day, day.Slots, false, 0, 0, 0, 0, candidate)
}

// ValidateCandidates validates all candidates in order with bipartite slot matching,
// per-day session-count and cumulative-minutes tracking, and RequestedSessionCount cap.
//
// Slot assignment: a maximum bipartite matching is pre-computed for each day's
// valid-input candidates before emitting per-candidate results. This ensures feasible
// schedules are accepted regardless of candidate input order.
//
// Invalid-input candidates (NaN/negative DurationMinutes or Load) increment the day
// session counter for positional tracking but are excluded from all numeric accumulations.
//
// Duplicate dates in AvailableDays: first entry wins (same as ValidateCandidate/findDay).
func ValidateCandidates(wc WeekConstraints, candidates []CandidateSession) BatchResult {
	type dayState struct {
		day      DayConstraints
		sessions int
		minutes  float64
	}
	dayStates := map[string]*dayState{}
	for _, day := range wc.AvailableDays {
		if _, exists := dayStates[day.Date]; exists {
			continue
		}
		d := day
		dayStates[day.Date] = &dayState{day: d}
	}

	// Pre-pass: group valid-input candidates by day and compute bipartite matching.
	type dayBatch struct {
		day      DayConstraints
		candIdxs []int // global indices for valid-input candidates
		cands    []CandidateSession
		// matchSlot[slotIdx] = local candidate index matched to that slot, or -1.
		matchSlot []int
	}
	dayLocalIdx := make([]int, len(candidates))
	for i := range dayLocalIdx {
		dayLocalIdx[i] = -1
	}
	dayBatches := map[string]*dayBatch{}
	for i, c := range candidates {
		if invalidCandidateInputViolation(c) != nil {
			continue
		}
		ds := dayStates[c.Date]
		if ds == nil || ds.day.MaxSessionsPerDay == 0 {
			continue
		}
		db := dayBatches[c.Date]
		if db == nil {
			dayBatches[c.Date] = &dayBatch{day: ds.day}
			db = dayBatches[c.Date]
		}
		dayLocalIdx[i] = len(db.candIdxs)
		db.candIdxs = append(db.candIdxs, i)
		db.cands = append(db.cands, c)
	}
	for _, db := range dayBatches {
		db.matchSlot = matchSlotsToOwner(db.day.Slots, db.cands)
	}

	// Build global candidate → matched slot index map.
	candidateMatchedSlot := make([]int, len(candidates))
	for i := range candidateMatchedSlot {
		candidateMatchedSlot[i] = -1
	}
	for _, db := range dayBatches {
		for slotIdx, localCandIdx := range db.matchSlot {
			if localCandIdx >= 0 {
				globalIdx := db.candIdxs[localCandIdx]
				candidateMatchedSlot[globalIdx] = slotIdx
			}
		}
	}

	var priorLoad, priorMinutes float64
	var validCount int
	results := make([]CandidateResult, len(candidates))

	for i, candidate := range candidates {
		var result CandidateResult

		if v := invalidCandidateInputViolation(candidate); v != nil {
			result = CandidateResult{
				Candidate:  sanitizeCandidateForResult(candidate),
				Valid:      false,
				Violations: []Violation{*v},
			}
			if ds := dayStates[candidate.Date]; ds != nil {
				ds.sessions++
			}
		} else {
			ds := dayStates[candidate.Date]

			if ds == nil || ds.day.MaxSessionsPerDay == 0 {
				result = CandidateResult{
					Candidate: candidate,
					Valid:     false,
					Violations: []Violation{{
						Code:    ViolationDayUnavailable,
						Message: "session date is not available for scheduling",
						Field:   "date",
						Value:   candidate.Date,
					}},
				}
				if ds != nil {
					ds.sessions++
					ds.minutes += candidate.DurationMinutes
				}
			} else {
				hasSlots := len(ds.day.Slots) > 0
				slotViolations := computeSlotViolations(ds.day.Slots, hasSlots, candidateMatchedSlot[i], candidate)

				result = validateCore(wc, ds.day, nil, true, ds.sessions, ds.minutes, priorLoad, priorMinutes, candidate)
				result.Violations = replaceSlotViolations(result.Violations, slotViolations)
				result.Valid = len(result.Violations) == 0

				ds.sessions++
				ds.minutes += candidate.DurationMinutes
			}

			priorLoad += candidate.Load
			priorMinutes += candidate.DurationMinutes
		}

		// Enforce RequestedSessionCount cap.
		if result.Valid && wc.RequestedSessionCount != nil && validCount >= *wc.RequestedSessionCount {
			result.Valid = false
			result.Violations = append(result.Violations, Violation{
				Code:    ViolationRequestedSessionCountExceeded,
				Message: "requested session count already reached; this candidate is excess",
				Field:   "requested_session_count",
				Value:   *wc.RequestedSessionCount,
			})
		}

		if result.Valid {
			validCount++
		}
		results[i] = result
	}

	var weekWarnings []Warning
	if wc.RequestedSessionCount != nil {
		totalSlots := availableSlotCount(wc)
		if *wc.RequestedSessionCount > totalSlots {
			weekWarnings = append(weekWarnings, Warning{
				Code:    WarnInfeasibleSessionCount,
				Message: "requested session count exceeds total structural slot capacity across available days",
				Field:   "requested_session_count",
				Value:   *wc.RequestedSessionCount,
			})
		}
		// Warn when the batch is underfilled: fewer valid sessions than requested.
		if validCount < *wc.RequestedSessionCount {
			weekWarnings = append(weekWarnings, Warning{
				Code:    WarnRequestedSessionCountUnmet,
				Message: "batch contains fewer valid sessions than requested; schedule is underfilled",
				Field:   "requested_session_count",
				Value:   map[string]int{"requested": *wc.RequestedSessionCount, "accepted": validCount},
			})
		}
	}

	recon := Reconcile(wc, candidates)

	if recon.RemainingLoad != nil && *recon.RemainingLoad > 0 && recon.CandidateLoad < *recon.RemainingLoad {
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

// matchSlotsToOwner runs maximum bipartite matching between slots and candidates using
// augmenting-path DFS in candidate input order for determinism.
// Returns matchSlot where matchSlot[slotIdx] = local candidate index, or -1.
func matchSlotsToOwner(slots []SlotConstraint, cands []CandidateSession) []int {
	n := len(slots)
	m := len(cands)
	matchSlot := make([]int, n)
	for i := range matchSlot {
		matchSlot[i] = -1
	}
	if n == 0 || m == 0 {
		return matchSlot
	}

	var augment func(candIdx int, visited []bool) bool
	augment = func(candIdx int, visited []bool) bool {
		for slotIdx, slot := range slots {
			if visited[slotIdx] {
				continue
			}
			if !slotFits(slot, cands[candIdx]) {
				continue
			}
			visited[slotIdx] = true
			if matchSlot[slotIdx] == -1 || augment(matchSlot[slotIdx], visited) {
				matchSlot[slotIdx] = candIdx
				return true
			}
		}
		return false
	}

	for i := range cands {
		visited := make([]bool, n)
		augment(i, visited)
	}
	return matchSlot
}

// computeSlotViolations returns slot violations for a batch candidate based on matching.
func computeSlotViolations(slots []SlotConstraint, hasSlots bool, matchedSlotIdx int, candidate CandidateSession) []Violation {
	if !hasSlots {
		return nil
	}
	if matchedSlotIdx >= 0 {
		return nil
	}
	if findCompatibleSlotIndex(slots, candidate) >= 0 {
		return []Violation{{
			Code:    ViolationNoAvailableSlot,
			Message: "candidate fits slot constraints but all compatible slots are claimed by other candidates",
			Field:   "date",
			Value:   candidate.Date,
		}}
	}
	return universalSlotViolations(slots, candidate)
}

var slotViolationCodes = map[ViolationCode]bool{
	ViolationSlotDuration:     true,
	ViolationIndoorDuration:   true,
	ViolationSportNotAllowed:  true,
	ViolationModeNotAllowed:   true,
	ViolationNoCompatibleSlot: true,
	ViolationNoAvailableSlot:  true,
}

func replaceSlotViolations(base []Violation, newSlot []Violation) []Violation {
	filtered := base[:0]
	for _, v := range base {
		if !slotViolationCodes[v.Code] {
			filtered = append(filtered, v)
		}
	}
	return append(filtered, newSlot...)
}

// validateCore validates non-slot constraints for a candidate.
// When batchSlotMode is true, slot constraints are skipped (handled by bipartite matching).
func validateCore(wc WeekConstraints, day DayConstraints, availableSlots []SlotConstraint, batchSlotMode bool, sessionsAlreadyOnDay int, dailyMinutesAlready float64, priorLoad float64, priorMinutes float64, candidate CandidateSession) CandidateResult {
	var violations []Violation
	var warnings []Warning

	if sessionsAlreadyOnDay >= day.MaxSessionsPerDay {
		violations = append(violations, Violation{
			Code:    ViolationDailySessionCount,
			Message: "maximum sessions per day already reached",
			Field:   "max_sessions_per_day",
			Value:   day.MaxSessionsPerDay,
		})
	}

	if day.MaxTotalDailyMinutes > 0 && dailyMinutesAlready+candidate.DurationMinutes > day.MaxTotalDailyMinutes {
		violations = append(violations, Violation{
			Code:    ViolationDailyTimeExceeded,
			Message: "combined daily training duration would exceed the daily cap",
			Field:   "max_total_daily_minutes",
			Value:   day.MaxTotalDailyMinutes,
		})
	}

	if !batchSlotMode && len(day.Slots) > 0 {
		if len(availableSlots) == 0 {
			violations = append(violations, Violation{
				Code:    ViolationNoAvailableSlot,
				Message: "all training slots for this day are already claimed",
				Field:   "date",
				Value:   candidate.Date,
			})
		} else {
			violations = append(violations, universalSlotViolations(availableSlots, candidate)...)
		}
	}

	if wc.WeeklyTargetLoad != nil {
		targetLoad := *wc.WeeklyTargetLoad
		remainingLoad := targetLoad - wc.CompletedLoad - wc.FixedLoad - priorLoad
		if remainingLoad <= 0 {
			warnings = append(warnings, Warning{
				Code:    WarnZeroRemainingLoad,
				Message: "remaining weekly load budget is zero or negative",
				Field:   "remaining_load",
				Value:   remainingLoad,
			})
			if candidate.Load > 0 {
				violations = append(violations, Violation{
					Code:    ViolationWeeklyLoadOvershoot,
					Message: "remaining weekly load budget is exhausted; candidate load cannot be placed",
					Field:   "weekly_target_load",
					Value:   remainingLoad,
				})
			}
		} else if candidate.Load > remainingLoad {
			violations = append(violations, Violation{
				Code:    ViolationWeeklyLoadOvershoot,
				Message: "candidate load exceeds remaining weekly load budget",
				Field:   "weekly_target_load",
				Value:   remainingLoad,
			})
		}
	}

	if wc.WeeklyTargetMinutes != nil {
		targetMin := *wc.WeeklyTargetMinutes
		remainingMin := targetMin - wc.CompletedMinutes - wc.FixedMinutes - priorMinutes
		if remainingMin <= 0 {
			warnings = append(warnings, Warning{
				Code:    WarnZeroRemainingTime,
				Message: "remaining weekly time budget is zero or negative",
				Field:   "remaining_minutes",
				Value:   remainingMin,
			})
			if candidate.DurationMinutes > 0 {
				violations = append(violations, Violation{
					Code:    ViolationWeeklyTimeOvershoot,
					Message: "remaining weekly time budget is exhausted; candidate duration cannot be placed",
					Field:   "weekly_target_minutes",
					Value:   remainingMin,
				})
			}
		} else if candidate.DurationMinutes > remainingMin {
			violations = append(violations, Violation{
				Code:    ViolationWeeklyTimeOvershoot,
				Message: "candidate duration exceeds remaining weekly time budget",
				Field:   "weekly_target_minutes",
				Value:   remainingMin,
			})
		}
	}

	return CandidateResult{
		Candidate:  candidate,
		Valid:      len(violations) == 0,
		Violations: violations,
		Warnings:   warnings,
	}
}

func sanitizeCandidateForResult(c CandidateSession) CandidateSession {
	if math.IsNaN(c.DurationMinutes) || math.IsInf(c.DurationMinutes, 0) || c.DurationMinutes < 0 {
		c.DurationMinutes = 0
	}
	if math.IsNaN(c.Load) || math.IsInf(c.Load, 0) || c.Load < 0 {
		c.Load = 0
	}
	return c
}

func invalidCandidateInputViolation(candidate CandidateSession) *Violation {
	if math.IsNaN(candidate.DurationMinutes) || math.IsInf(candidate.DurationMinutes, 0) || candidate.DurationMinutes < 0 {
		return &Violation{
			Code:    ViolationInvalidInput,
			Message: "duration_minutes must be a finite non-negative number",
			Field:   "duration_minutes",
			Value:   fmt.Sprintf("%v", candidate.DurationMinutes),
		}
	}
	if math.IsNaN(candidate.Load) || math.IsInf(candidate.Load, 0) || candidate.Load < 0 {
		return &Violation{
			Code:    ViolationInvalidInput,
			Message: "load must be a finite non-negative number",
			Field:   "load",
			Value:   fmt.Sprintf("%v", candidate.Load),
		}
	}
	return nil
}

func slotFits(slot SlotConstraint, candidate CandidateSession) bool {
	if slot.MaxDurationMinutes > 0 && candidate.DurationMinutes > slot.MaxDurationMinutes {
		return false
	}
	if candidate.Indoor && slot.MaxIndoorMinutes > 0 && candidate.DurationMinutes > slot.MaxIndoorMinutes {
		return false
	}
	if len(slot.AllowedSports) > 0 && !slices.Contains(slot.AllowedSports, candidate.Sport) {
		return false
	}
	if len(slot.AllowedModes) > 0 && !slices.Contains(slot.AllowedModes, candidate.Mode) {
		return false
	}
	return true
}

func findCompatibleSlotIndex(slots []SlotConstraint, candidate CandidateSession) int {
	for i, slot := range slots {
		if slotFits(slot, candidate) {
			return i
		}
	}
	return -1
}

// universalSlotViolations returns violations when no slot fits the candidate.
// A ViolationCode is emitted only when it applies to every slot (universal rejection).
// ViolationNoCompatibleSlot is the fallback for mixed-reason rejections.
func universalSlotViolations(slots []SlotConstraint, candidate CandidateSession) []Violation {
	if len(slots) == 0 {
		return nil
	}
	if findCompatibleSlotIndex(slots, candidate) >= 0 {
		return nil
	}

	allDuration, allIndoor, allSport, allMode := true, true, true, true
	for _, slot := range slots {
		if !(slot.MaxDurationMinutes > 0 && candidate.DurationMinutes > slot.MaxDurationMinutes) {
			allDuration = false
		}
		if !candidate.Indoor || slot.MaxIndoorMinutes == 0 || candidate.DurationMinutes <= slot.MaxIndoorMinutes {
			allIndoor = false
		}
		if len(slot.AllowedSports) == 0 || slices.Contains(slot.AllowedSports, candidate.Sport) {
			allSport = false
		}
		if len(slot.AllowedModes) == 0 || slices.Contains(slot.AllowedModes, candidate.Mode) {
			allMode = false
		}
	}

	var violations []Violation
	if allDuration {
		violations = append(violations, Violation{Code: ViolationSlotDuration, Message: "session duration exceeds every available slot duration cap", Field: "duration_minutes", Value: candidate.DurationMinutes})
	}
	if allIndoor {
		violations = append(violations, Violation{Code: ViolationIndoorDuration, Message: "indoor session duration exceeds every available slot indoor cap", Field: "duration_minutes", Value: candidate.DurationMinutes})
	}
	if allSport {
		violations = append(violations, Violation{Code: ViolationSportNotAllowed, Message: "session sport is excluded by every available slot", Field: "sport", Value: candidate.Sport})
	}
	if allMode {
		violations = append(violations, Violation{Code: ViolationModeNotAllowed, Message: "session mode is excluded by every available slot", Field: "mode", Value: candidate.Mode})
	}
	if len(violations) == 0 {
		violations = append(violations, Violation{Code: ViolationNoCompatibleSlot, Message: "no available slot can accommodate this candidate; constraints differ across slots", Field: "date", Value: candidate.Date})
	}
	return violations
}

func buildReconciliation(wc WeekConstraints, candMin, candLoad float64) Reconciliation {
	r := Reconciliation{
		CompletedMinutes: wc.CompletedMinutes,
		CompletedLoad:    wc.CompletedLoad,
		FixedMinutes:     wc.FixedMinutes,
		FixedLoad:        wc.FixedLoad,
		CandidateMinutes: candMin,
		CandidateLoad:    candLoad,
		ProjectedMinutes: wc.CompletedMinutes + wc.FixedMinutes + candMin,
		ProjectedLoad:    wc.CompletedLoad + wc.FixedLoad + candLoad,
	}
	if wc.WeeklyTargetMinutes != nil {
		target := *wc.WeeklyTargetMinutes
		r.WeeklyTargetMinutes = &target
		rem := target - wc.CompletedMinutes - wc.FixedMinutes
		r.RemainingMinutes = &rem
	}
	if wc.WeeklyTargetLoad != nil {
		target := *wc.WeeklyTargetLoad
		r.WeeklyTargetLoad = &target
		rem := target - wc.CompletedLoad - wc.FixedLoad
		r.RemainingLoad = &rem
	}
	return r
}

func availableSlotCount(wc WeekConstraints) int {
	total := 0
	for _, day := range wc.AvailableDays {
		if day.MaxSessionsPerDay <= 0 {
			continue
		}
		cap := day.MaxSessionsPerDay
		if len(day.Slots) > 0 && len(day.Slots) < cap {
			cap = len(day.Slots)
		}
		total += cap
	}
	return total
}

func findDay(days []DayConstraints, date string) (DayConstraints, bool) {
	for _, d := range days {
		if d.Date == date {
			return d, true
		}
	}
	return DayConstraints{}, false
}

func requireFiniteNonNegative(field string, v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%s must be a finite number, got %v", field, v)
	}
	if v < 0 {
		return fmt.Errorf("%s must be non-negative, got %v", field, v)
	}
	return nil
}

func derefFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

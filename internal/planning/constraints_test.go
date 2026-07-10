package planning_test

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/planning"
)

// ptrF returns a pointer to the given float64.
func ptrF(v float64) *float64 { return &v }

// ptrI returns a pointer to the given int.
func ptrI(v int) *int { return &v }

func hasViolation(r planning.CandidateResult, code planning.ViolationCode) bool {
	for _, v := range r.Violations {
		if v.Code == code {
			return true
		}
	}
	return false
}

func hasBatchWarning(b planning.BatchResult, code planning.WarningCode) bool {
	for _, w := range b.Warnings {
		if w.Code == code {
			return true
		}
	}
	return false
}

func hasCandidateWarning(r planning.CandidateResult, code planning.WarningCode) bool {
	for _, w := range r.Warnings {
		if w.Code == code {
			return true
		}
	}
	return false
}

// weekWithDay returns a minimal WeekConstraints with one available day.
func weekWithDay(date string, slots ...planning.SlotConstraint) planning.WeekConstraints {
	return planning.WeekConstraints{
		WeekStartDate: "2026-07-06", // Monday
		AvailableDays: []planning.DayConstraints{
			{Date: date, MaxSessionsPerDay: 2, Slots: slots},
		},
	}
}

// ─── ValidateCandidate boundary tests ───────────────────────────────────────

func TestValidateCandidate_InProgressWeekOvershoot(t *testing.T) {
	// Full-week target 300 load; 120 already completed → remaining 180.
	// A 200-load candidate should overshoot; a 180-load candidate should fit.
	wc := planning.WeekConstraints{
		WeekStartDate:    "2026-07-06",
		WeeklyTargetLoad: ptrF(300),
		CompletedLoad:    120,
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
		},
	}

	overshoot := planning.CandidateSession{Date: "2026-07-06", Load: 200}
	r := planning.ValidateCandidate(wc, overshoot)
	if r.Valid {
		t.Error("load 200 > remaining 180: expected invalid")
	}
	if !hasViolation(r, planning.ViolationWeeklyLoadOvershoot) {
		t.Errorf("expected weekly_load_overshoot, got violations: %v", r.Violations)
	}

	justFit := planning.CandidateSession{Date: "2026-07-06", Load: 180}
	r2 := planning.ValidateCandidate(wc, justFit)
	if !r2.Valid {
		t.Errorf("load 180 = remaining 180: expected valid, got violations: %v", r2.Violations)
	}
}

func TestValidateCandidate_TwoSlotsCannotCombine(t *testing.T) {
	// Two 45-minute slots; a 95-minute session must not fit.
	slot45 := planning.SlotConstraint{MaxDurationMinutes: 45}
	wc := weekWithDay("2026-07-06", slot45, slot45)

	long := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 95}
	r := planning.ValidateCandidate(wc, long)
	if r.Valid {
		t.Error("95-minute session in two 45-minute slots: expected invalid")
	}
	if !hasViolation(r, planning.ViolationSlotDuration) {
		t.Errorf("expected slot_duration_exceeded, got violations: %v", r.Violations)
	}

	fits := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 45}
	r2 := planning.ValidateCandidate(wc, fits)
	if !r2.Valid {
		t.Errorf("45-minute session in 45-minute slot: expected valid, got violations: %v", r2.Violations)
	}
}

func TestValidateCandidate_IndoorCapDoesNotConstrainOutdoor(t *testing.T) {
	// One slot: MaxDuration=120, MaxIndoor=60.
	slot := planning.SlotConstraint{MaxDurationMinutes: 120, MaxIndoorMinutes: 60}
	wc := weekWithDay("2026-07-06", slot)

	outdoor := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 90, Indoor: false}
	r := planning.ValidateCandidate(wc, outdoor)
	if !r.Valid {
		t.Errorf("90-min outdoor session in 120-min slot: expected valid, got violations: %v", r.Violations)
	}

	indoor := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 90, Indoor: true}
	r2 := planning.ValidateCandidate(wc, indoor)
	if r2.Valid {
		t.Error("90-min indoor session with 60-min indoor cap: expected invalid")
	}
	if !hasViolation(r2, planning.ViolationIndoorDuration) {
		t.Errorf("expected indoor_duration_exceeded, got violations: %v", r2.Violations)
	}
}

func TestValidateCandidate_FixedEventsReduceBudget(t *testing.T) {
	// Target 400 load; fixed events commit 150. Remaining = 250.
	wc := planning.WeekConstraints{
		WeekStartDate:    "2026-07-06",
		WeeklyTargetLoad: ptrF(400),
		FixedLoad:        150,
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
		},
	}

	over := planning.CandidateSession{Date: "2026-07-06", Load: 300}
	if r := planning.ValidateCandidate(wc, over); r.Valid {
		t.Error("300 load > remaining 250: expected invalid")
	}

	exact := planning.CandidateSession{Date: "2026-07-06", Load: 250}
	if r := planning.ValidateCandidate(wc, exact); !r.Valid {
		t.Errorf("250 load = remaining 250: expected valid, got violations: %v", r.Violations)
	}
}

func TestValidateCandidate_ZeroRemainingLoad(t *testing.T) {
	// Completed + fixed already exceed target → remaining is negative.
	wc := planning.WeekConstraints{
		WeekStartDate:    "2026-07-06",
		WeeklyTargetLoad: ptrF(300),
		CompletedLoad:    200,
		FixedLoad:        150,
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
		},
	}

	// Positive load: warn AND violate.
	pos := planning.CandidateSession{Date: "2026-07-06", Load: 50}
	r := planning.ValidateCandidate(wc, pos)
	if r.Valid {
		t.Error("positive load on exhausted budget: expected invalid")
	}
	if !hasViolation(r, planning.ViolationWeeklyLoadOvershoot) {
		t.Errorf("expected weekly_load_overshoot, got violations: %v", r.Violations)
	}
	if !hasCandidateWarning(r, planning.WarnZeroRemainingLoad) {
		t.Errorf("expected zero_remaining_load warning, got warnings: %v", r.Warnings)
	}

	// Zero load: warn only, no violation.
	zero := planning.CandidateSession{Date: "2026-07-06", Load: 0}
	r2 := planning.ValidateCandidate(wc, zero)
	if !r2.Valid {
		t.Errorf("zero-load session on exhausted budget: expected valid, got violations: %v", r2.Violations)
	}
	if !hasCandidateWarning(r2, planning.WarnZeroRemainingLoad) {
		t.Errorf("expected zero_remaining_load warning, got warnings: %v", r2.Warnings)
	}
}

func TestValidateCandidate_UnavailableDay(t *testing.T) {
	wc := planning.WeekConstraints{
		WeekStartDate: "2026-07-06",
		AvailableDays: []planning.DayConstraints{},
	}
	c := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 60}
	r := planning.ValidateCandidate(wc, c)
	if r.Valid {
		t.Error("day not in AvailableDays: expected invalid")
	}
	if !hasViolation(r, planning.ViolationDayUnavailable) {
		t.Errorf("expected day_unavailable, got violations: %v", r.Violations)
	}
}

func TestValidateCandidate_ZeroCapacityDay(t *testing.T) {
	// MaxSessionsPerDay=0 means unavailable.
	wc := planning.WeekConstraints{
		WeekStartDate: "2026-07-06",
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 0},
		},
	}
	c := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 30}
	r := planning.ValidateCandidate(wc, c)
	if r.Valid {
		t.Error("MaxSessionsPerDay=0: expected invalid")
	}
	if !hasViolation(r, planning.ViolationDayUnavailable) {
		t.Errorf("expected day_unavailable, got violations: %v", r.Violations)
	}
}

func TestValidateCandidate_InvalidInput_NaN(t *testing.T) {
	wc := planning.WeekConstraints{
		WeekStartDate: "2026-07-06",
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
		},
	}

	c := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: math.NaN()}
	r := planning.ValidateCandidate(wc, c)
	if r.Valid {
		t.Error("NaN duration: expected invalid")
	}
	if !hasViolation(r, planning.ViolationInvalidInput) {
		t.Errorf("expected invalid_input, got violations: %v", r.Violations)
	}

	// Result must be JSON-marshalable.
	if _, err := json.Marshal(r); err != nil {
		t.Errorf("JSON marshal of invalid-input result failed: %v", err)
	}
}

func TestValidateCandidate_NilTargets_NoChecks(t *testing.T) {
	// Nil WeeklyTargetLoad and WeeklyTargetMinutes → no budget checks → valid.
	wc := planning.WeekConstraints{
		WeekStartDate: "2026-07-06",
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
		},
	}
	c := planning.CandidateSession{Date: "2026-07-06", Load: 9999, DurationMinutes: 9999}
	r := planning.ValidateCandidate(wc, c)
	if !r.Valid {
		t.Errorf("nil targets: expected valid regardless of load/duration, got violations: %v", r.Violations)
	}
}

func TestValidateCandidate_ZeroRequestedSessions_BlocksCandidate(t *testing.T) {
	// ptrI(0) = zero sessions wanted; single-candidate API should also enforce this.
	wc := planning.WeekConstraints{
		WeekStartDate:         "2026-07-06",
		RequestedSessionCount: ptrI(0),
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
		},
	}
	c := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 30}
	r := planning.ValidateCandidate(wc, c)
	if r.Valid {
		t.Error("ValidateCandidate with RequestedSessionCount=0: expected invalid")
	}
	if !hasViolation(r, planning.ViolationRequestedSessionCountExceeded) {
		t.Errorf("expected requested_session_count_exceeded, got violations: %v", r.Violations)
	}
}

func TestValidateCandidate_ZeroTargets_BlockPositiveWork(t *testing.T) {
	// ptrF(0) means explicit zero budget → positive load/duration are blocked.
	wc := planning.WeekConstraints{
		WeekStartDate:       "2026-07-06",
		WeeklyTargetLoad:    ptrF(0),
		WeeklyTargetMinutes: ptrF(0),
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
		},
	}
	c := planning.CandidateSession{Date: "2026-07-06", Load: 1, DurationMinutes: 1}
	r := planning.ValidateCandidate(wc, c)
	if r.Valid {
		t.Error("explicit zero targets: positive candidate expected invalid")
	}
	if !hasViolation(r, planning.ViolationWeeklyLoadOvershoot) {
		t.Errorf("expected weekly_load_overshoot, got violations: %v", r.Violations)
	}
	if !hasViolation(r, planning.ViolationWeeklyTimeOvershoot) {
		t.Errorf("expected weekly_time_overshoot, got violations: %v", r.Violations)
	}
}

// ─── ValidateCandidates boundary tests ──────────────────────────────────────

func TestValidateCandidates_BipartiteMatchingAvoidsGreedyRejection(t *testing.T) {
	// Regression for R008: greedy first-fit causes Ride to claim any-sport slot,
	// leaving Run with only the Ride-only slot (which it can't use).
	// Bipartite matching should find: Ride → Ride-only, Run → any-sport.
	wc := planning.WeekConstraints{
		WeekStartDate: "2026-07-06",
		AvailableDays: []planning.DayConstraints{{
			Date:              "2026-07-06",
			MaxSessionsPerDay: 2,
			Slots: []planning.SlotConstraint{
				{MaxDurationMinutes: 60},                                  // any-sport
				{MaxDurationMinutes: 60, AllowedSports: []string{"Ride"}}, // Ride-only
			},
		}},
	}
	candidates := []planning.CandidateSession{
		{Date: "2026-07-06", Sport: "Ride", DurationMinutes: 60},
		{Date: "2026-07-06", Sport: "Run", DurationMinutes: 60},
	}

	batch := planning.ValidateCandidates(wc, candidates)
	if !batch.Results[0].Valid {
		t.Errorf("Ride should be valid, got violations: %v", batch.Results[0].Violations)
	}
	if !batch.Results[1].Valid {
		t.Errorf("Run should be valid after bipartite matching, got violations: %v", batch.Results[1].Violations)
	}
}

func TestValidateCandidates_RequestedSessionCount_Nil_NoCap(t *testing.T) {
	// nil RequestedSessionCount → no cap → all valid candidates accepted.
	wc := planning.WeekConstraints{
		WeekStartDate:         "2026-07-06",
		RequestedSessionCount: nil,
		AvailableDays: []planning.DayConstraints{{
			Date:              "2026-07-06",
			MaxSessionsPerDay: 3,
		}},
	}
	candidates := []planning.CandidateSession{
		{Date: "2026-07-06", DurationMinutes: 30},
		{Date: "2026-07-06", DurationMinutes: 30},
		{Date: "2026-07-06", DurationMinutes: 30},
	}
	batch := planning.ValidateCandidates(wc, candidates)
	for i, r := range batch.Results {
		if !r.Valid {
			t.Errorf("candidate %d should be valid (nil session cap), got violations: %v", i, r.Violations)
		}
	}
}

func TestValidateCandidates_RequestedSessionCount_Zero_BlocksAll(t *testing.T) {
	// ptrI(0) → zero sessions requested → all candidates are excess.
	wc := planning.WeekConstraints{
		WeekStartDate:         "2026-07-06",
		RequestedSessionCount: ptrI(0),
		AvailableDays: []planning.DayConstraints{{
			Date:              "2026-07-06",
			MaxSessionsPerDay: 1,
		}},
	}
	c := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 30}
	batch := planning.ValidateCandidates(wc, []planning.CandidateSession{c})
	if batch.Results[0].Valid {
		t.Error("RequestedSessionCount=0: candidate expected to be excess")
	}
	if !hasViolation(batch.Results[0], planning.ViolationRequestedSessionCountExceeded) {
		t.Errorf("expected requested_session_count_exceeded, got violations: %v", batch.Results[0].Violations)
	}
}

func TestValidateCandidates_UnderfillWarns(t *testing.T) {
	// RequestedSessionCount=3 but only 1 valid candidate → underfill warning.
	wc := planning.WeekConstraints{
		WeekStartDate:         "2026-07-06",
		RequestedSessionCount: ptrI(3),
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 3},
		},
	}
	candidates := []planning.CandidateSession{
		{Date: "2026-07-06", DurationMinutes: 30}, // valid
	}
	batch := planning.ValidateCandidates(wc, candidates)
	if !batch.Results[0].Valid {
		t.Errorf("single valid candidate should be valid, got violations: %v", batch.Results[0].Violations)
	}
	if !hasBatchWarning(batch, planning.WarnRequestedSessionCountUnmet) {
		t.Errorf("expected requested_session_count_unmet warning, got warnings: %v", batch.Warnings)
	}
}

func TestValidateCandidates_InfeasibleSessionCount_Warns(t *testing.T) {
	// RequestedSessionCount=5 but only 2 structural slots → warn.
	wc := planning.WeekConstraints{
		WeekStartDate:         "2026-07-06",
		RequestedSessionCount: ptrI(5),
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1, Slots: []planning.SlotConstraint{{MaxDurationMinutes: 60}}},
			{Date: "2026-07-07", MaxSessionsPerDay: 1, Slots: []planning.SlotConstraint{{MaxDurationMinutes: 60}}},
		},
	}
	batch := planning.ValidateCandidates(wc, nil)
	if !hasBatchWarning(batch, planning.WarnInfeasibleSessionCount) {
		t.Errorf("expected infeasible_session_count warning, got warnings: %v", batch.Warnings)
	}
}

func TestValidateCandidates_ZeroCapacityDay_MatchesSingleCandidate(t *testing.T) {
	// Both single and batch APIs should agree on day_unavailable for MaxSessionsPerDay=0.
	wc := planning.WeekConstraints{
		WeekStartDate: "2026-07-06",
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 0},
		},
	}
	c := planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 30}

	single := planning.ValidateCandidate(wc, c)
	if !hasViolation(single, planning.ViolationDayUnavailable) {
		t.Errorf("ValidateCandidate: expected day_unavailable, got violations: %v", single.Violations)
	}

	batch := planning.ValidateCandidates(wc, []planning.CandidateSession{c})
	if !hasViolation(batch.Results[0], planning.ViolationDayUnavailable) {
		t.Errorf("ValidateCandidates: expected day_unavailable, got violations: %v", batch.Results[0].Violations)
	}
}

func TestValidateCandidates_InvalidInputDoesNotPoisonSubsequent(t *testing.T) {
	// An invalid first candidate (NaN load) must not make subsequent budget checks fail.
	// Without NaN isolation, priorLoad becomes NaN and all comparisons fail.
	wc := planning.WeekConstraints{
		WeekStartDate:    "2026-07-06",
		WeeklyTargetLoad: ptrF(100),
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
			{Date: "2026-07-07", MaxSessionsPerDay: 1},
		},
	}
	candidates := []planning.CandidateSession{
		{Date: "2026-07-06", Load: math.NaN(), DurationMinutes: 60}, // invalid
		{Date: "2026-07-07", Load: 80, DurationMinutes: 45},         // should be valid (80 <= 100)
	}

	batch := planning.ValidateCandidates(wc, candidates)
	if batch.Results[0].Valid {
		t.Error("NaN candidate should be invalid")
	}
	if !hasViolation(batch.Results[0], planning.ViolationInvalidInput) {
		t.Errorf("expected invalid_input for NaN candidate, got violations: %v", batch.Results[0].Violations)
	}
	if !batch.Results[1].Valid {
		t.Errorf("valid-input candidate after NaN should be valid, got violations: %v", batch.Results[1].Violations)
	}

	// JSON marshal must succeed.
	if _, err := json.Marshal(batch); err != nil {
		t.Errorf("JSON marshal of batch with invalid input failed: %v", err)
	}
}

// ─── Reconcile tests ─────────────────────────────────────────────────────────

func TestReconcile_ExcludesInvalidInputCandidates(t *testing.T) {
	wc := planning.WeekConstraints{
		WeekStartDate:    "2026-07-06",
		WeeklyTargetLoad: ptrF(200),
	}
	candidates := []planning.CandidateSession{
		{Load: math.NaN(), DurationMinutes: 60},  // invalid, excluded
		{Load: math.Inf(1), DurationMinutes: 30}, // invalid, excluded
		{Load: -10, DurationMinutes: 20},         // invalid (negative), excluded
		{Load: 50, DurationMinutes: 45},          // valid
	}

	recon, err := planning.Reconcile(wc, candidates)
	if err != nil {
		t.Fatalf("Reconcile unexpectedly failed: %v", err)
	}
	if math.IsNaN(recon.CandidateLoad) || math.IsInf(recon.CandidateLoad, 0) {
		t.Errorf("CandidateLoad must be finite, got %v", recon.CandidateLoad)
	}
	if recon.CandidateLoad != 50 {
		t.Errorf("expected CandidateLoad=50, got %v", recon.CandidateLoad)
	}
	if recon.CandidateMinutes != 45 {
		t.Errorf("expected CandidateMinutes=45, got %v", recon.CandidateMinutes)
	}

	// Must be JSON-marshalable.
	if _, err := json.Marshal(recon); err != nil {
		t.Errorf("JSON marshal of reconciliation failed: %v", err)
	}
}

func TestReconcile_NilTargets(t *testing.T) {
	// Nil targets → WeeklyTargetMinutes and WeeklyTargetLoad are nil in output;
	// RemainingMinutes and RemainingLoad are nil; no negative remaining from completed/fixed.
	wc := planning.WeekConstraints{
		WeekStartDate:    "2026-07-06",
		CompletedLoad:    50, // untracked; must not produce negative RemainingLoad
		CompletedMinutes: 30,
	}
	recon, err := planning.Reconcile(wc, nil)
	if err != nil {
		t.Fatalf("Reconcile unexpectedly failed: %v", err)
	}
	if recon.WeeklyTargetMinutes != nil {
		t.Errorf("nil WeeklyTargetMinutes should produce nil in reconciliation, got %v", recon.WeeklyTargetMinutes)
	}
	if recon.WeeklyTargetLoad != nil {
		t.Errorf("nil WeeklyTargetLoad should produce nil in reconciliation, got %v", recon.WeeklyTargetLoad)
	}
	if recon.RemainingMinutes != nil {
		t.Errorf("nil target: RemainingMinutes should be nil, got %v", recon.RemainingMinutes)
	}
	if recon.RemainingLoad != nil {
		t.Errorf("nil target: RemainingLoad should be nil, got %v", recon.RemainingLoad)
	}

	// JSON must not contain the nil fields.
	b, err := json.Marshal(recon)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	s := string(b)
	if strings.Contains(s, "weekly_target") || strings.Contains(s, "remaining") {
		t.Errorf("nil target fields should be absent from JSON, got: %s", s)
	}
}

func TestReconcile_ExplicitZeroTargetPreserved(t *testing.T) {
	// pointer-to-0 is an explicit zero budget and must appear in reconciliation output.
	wc := planning.WeekConstraints{
		WeekStartDate:       "2026-07-06",
		WeeklyTargetMinutes: ptrF(0),
		WeeklyTargetLoad:    ptrF(0),
		CompletedLoad:       10,
		CompletedMinutes:    5,
	}
	recon, err := planning.Reconcile(wc, nil)
	if err != nil {
		t.Fatalf("Reconcile unexpectedly failed: %v", err)
	}
	if recon.WeeklyTargetLoad == nil || *recon.WeeklyTargetLoad != 0 {
		t.Errorf("explicit zero target should appear as 0 in reconciliation")
	}
	if recon.RemainingLoad == nil || *recon.RemainingLoad != -10 {
		t.Errorf("explicit zero target with completed=10 should have RemainingLoad=-10, got %v", recon.RemainingLoad)
	}
}

func TestReconcile_OverflowReturnsError(t *testing.T) {
	// Two candidates each with MaxFloat64 load: sum overflows to +Inf.
	wc := planning.WeekConstraints{WeekStartDate: "2026-07-06"}
	candidates := []planning.CandidateSession{
		{Load: math.MaxFloat64, DurationMinutes: 60},
		{Load: math.MaxFloat64, DurationMinutes: 60},
	}
	_, err := planning.Reconcile(wc, candidates)
	if err == nil {
		t.Error("expected overflow error from Reconcile")
	}
}

func TestValidateCandidates_OverflowWarnAndJSONSafe(t *testing.T) {
	wc := planning.WeekConstraints{
		WeekStartDate: "2026-07-06",
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 2},
		},
	}
	candidates := []planning.CandidateSession{
		{Date: "2026-07-06", Load: math.MaxFloat64, DurationMinutes: 60},
		{Date: "2026-07-06", Load: math.MaxFloat64, DurationMinutes: 60},
	}
	batch := planning.ValidateCandidates(wc, candidates)
	if !hasBatchWarning(batch, planning.WarnArithmeticOverflow) {
		t.Errorf("expected arithmetic_overflow warning, got warnings: %v", batch.Warnings)
	}
	// JSON must not fail.
	if _, jsonErr := json.Marshal(batch); jsonErr != nil {
		t.Errorf("JSON marshal of overflow batch failed: %v", jsonErr)
	}
}

// ─── ValidateWeekConstraints tests ───────────────────────────────────────────

func TestValidateWeekConstraints_Valid(t *testing.T) {
	wc := planning.WeekConstraints{
		WeekStartDate:    "2026-07-06",
		WeeklyTargetLoad: ptrF(300),
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
			{Date: "2026-07-07", MaxSessionsPerDay: 2},
		},
	}
	if err := planning.ValidateWeekConstraints(wc); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateWeekConstraints_NonMondayStartDate(t *testing.T) {
	wc := planning.WeekConstraints{WeekStartDate: "2026-07-07"} // Tuesday
	if err := planning.ValidateWeekConstraints(wc); err == nil {
		t.Error("expected error for non-Monday start date")
	}
}

func TestValidateWeekConstraints_OutOfWeekDay(t *testing.T) {
	wc := planning.WeekConstraints{
		WeekStartDate: "2026-07-06",
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-20", MaxSessionsPerDay: 1}, // outside July 6-12
		},
	}
	if err := planning.ValidateWeekConstraints(wc); err == nil {
		t.Error("expected error for out-of-week day")
	}
}

func TestValidateWeekConstraints_DuplicateDayDate(t *testing.T) {
	wc := planning.WeekConstraints{
		WeekStartDate: "2026-07-06",
		AvailableDays: []planning.DayConstraints{
			{Date: "2026-07-06", MaxSessionsPerDay: 1},
			{Date: "2026-07-06", MaxSessionsPerDay: 2},
		},
	}
	if err := planning.ValidateWeekConstraints(wc); err == nil {
		t.Error("expected error for duplicate date")
	}
}

func TestValidateWeekConstraints_NaNTargetRejected(t *testing.T) {
	wc := planning.WeekConstraints{
		WeekStartDate:    "2026-07-06",
		WeeklyTargetLoad: ptrF(math.NaN()),
	}
	if err := planning.ValidateWeekConstraints(wc); err == nil {
		t.Error("expected error for NaN weekly_target_load")
	}
}

func TestValidateWeekConstraints_DeterministicFirstError(t *testing.T) {
	// Both targets are NaN; weekly_target_minutes must be reported first.
	wc := planning.WeekConstraints{
		WeekStartDate:       "2026-07-06",
		WeeklyTargetMinutes: ptrF(math.NaN()),
		WeeklyTargetLoad:    ptrF(math.NaN()),
	}
	err := planning.ValidateWeekConstraints(wc)
	if err == nil {
		t.Fatal("expected error for NaN targets")
	}
	if !strings.Contains(err.Error(), "weekly_target_minutes") {
		t.Errorf("expected first error to mention weekly_target_minutes, got: %v", err)
	}
}

func TestValidateWeekConstraints_NilTargets_Valid(t *testing.T) {
	// Nil WeeklyTargetLoad and WeeklyTargetMinutes are valid (opt-out of budget tracking).
	wc := planning.WeekConstraints{WeekStartDate: "2026-07-06"}
	if err := planning.ValidateWeekConstraints(wc); err != nil {
		t.Errorf("nil targets should be valid, got: %v", err)
	}
}

func TestValidateWeekConstraints_ZeroTargets_Valid(t *testing.T) {
	// ptrF(0) is a valid explicit zero budget.
	wc := planning.WeekConstraints{
		WeekStartDate:       "2026-07-06",
		WeeklyTargetMinutes: ptrF(0),
		WeeklyTargetLoad:    ptrF(0),
	}
	if err := planning.ValidateWeekConstraints(wc); err != nil {
		t.Errorf("zero targets should be valid, got: %v", err)
	}
}

func TestValidateWeekConstraints_CompletedFixedOverflowRejected(t *testing.T) {
	tests := []struct {
		name string
		week planning.WeekConstraints
	}{
		{
			name: "load",
			week: planning.WeekConstraints{
				WeekStartDate: "2026-07-06",
				CompletedLoad: math.MaxFloat64,
				FixedLoad:     math.MaxFloat64,
			},
		},
		{
			name: "minutes",
			week: planning.WeekConstraints{
				WeekStartDate:    "2026-07-06",
				CompletedMinutes: math.MaxFloat64,
				FixedMinutes:     math.MaxFloat64,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := planning.ValidateWeekConstraints(tc.week); err == nil {
				t.Error("expected completed plus fixed overflow to be rejected")
			}
		})
	}
}

func TestValidation_CompletedFixedOverflowJSONSafe(t *testing.T) {
	tests := []struct {
		name      string
		week      planning.WeekConstraints
		candidate planning.CandidateSession
	}{
		{
			name: "load",
			week: planning.WeekConstraints{
				WeekStartDate:    "2026-07-06",
				WeeklyTargetLoad: ptrF(0),
				CompletedLoad:    math.MaxFloat64,
				FixedLoad:        math.MaxFloat64,
				AvailableDays:    []planning.DayConstraints{{Date: "2026-07-06", MaxSessionsPerDay: 1}},
			},
			candidate: planning.CandidateSession{Date: "2026-07-06", Load: 1},
		},
		{
			name: "minutes",
			week: planning.WeekConstraints{
				WeekStartDate:       "2026-07-06",
				WeeklyTargetMinutes: ptrF(0),
				CompletedMinutes:    math.MaxFloat64,
				FixedMinutes:        math.MaxFloat64,
				AvailableDays:       []planning.DayConstraints{{Date: "2026-07-06", MaxSessionsPerDay: 1}},
			},
			candidate: planning.CandidateSession{Date: "2026-07-06", DurationMinutes: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			single := planning.ValidateCandidate(tc.week, tc.candidate)
			if !hasViolation(single, planning.ViolationArithmeticOverflow) {
				t.Errorf("single validation should report arithmetic overflow: %v", single.Violations)
			}
			if _, err := json.Marshal(single); err != nil {
				t.Fatalf("single validation result must marshal: %v", err)
			}

			batch := planning.ValidateCandidates(tc.week, []planning.CandidateSession{tc.candidate})
			if !hasViolation(batch.Results[0], planning.ViolationArithmeticOverflow) {
				t.Errorf("batch validation should report arithmetic overflow: %v", batch.Results[0].Violations)
			}
			if _, err := json.Marshal(batch); err != nil {
				t.Fatalf("batch validation result must marshal: %v", err)
			}
		})
	}
}

func TestValidateCandidates_PriorAccumulationOverflowJSONSafe(t *testing.T) {
	tests := []struct {
		name       string
		week       planning.WeekConstraints
		candidates []planning.CandidateSession
	}{
		{
			name: "load",
			week: planning.WeekConstraints{
				WeekStartDate:    "2026-07-06",
				WeeklyTargetLoad: ptrF(math.MaxFloat64),
				AvailableDays:    []planning.DayConstraints{{Date: "2026-07-06", MaxSessionsPerDay: 2}},
			},
			candidates: []planning.CandidateSession{
				{Date: "2026-07-06", Load: math.MaxFloat64},
				{Date: "2026-07-06", Load: math.MaxFloat64},
			},
		},
		{
			name: "minutes",
			week: planning.WeekConstraints{
				WeekStartDate:       "2026-07-06",
				WeeklyTargetMinutes: ptrF(math.MaxFloat64),
				AvailableDays:       []planning.DayConstraints{{Date: "2026-07-06", MaxSessionsPerDay: 2}},
			},
			candidates: []planning.CandidateSession{
				{Date: "2026-07-06", DurationMinutes: math.MaxFloat64},
				{Date: "2026-07-06", DurationMinutes: math.MaxFloat64},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			batch := planning.ValidateCandidates(tc.week, tc.candidates)
			if !hasViolation(batch.Results[1], planning.ViolationArithmeticOverflow) {
				t.Errorf("second candidate should report prior accumulation overflow: %v", batch.Results[1].Violations)
			}
			if _, err := json.Marshal(batch); err != nil {
				t.Fatalf("batch validation result must marshal: %v", err)
			}
		})
	}
}

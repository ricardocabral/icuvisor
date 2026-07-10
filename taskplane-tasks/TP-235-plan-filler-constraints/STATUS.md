# TP-235: Add plan-filler constraint model and validator — Status

**Current Step:** Step 3: Add boundary-focused regression coverage
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 18
**Iteration:** 3
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing planning semantics reviewed

---

### Step 1: Define the constraint contract

**Status:** ✅ Complete

<!-- R014 revision items -->
- [x] R014-1: Protect intermediate arithmetic in validateCore (remaining budget subtraction can produce -Inf when completed+fixed overflow; guard before embedding in Violation.Value/Warning.Value)
- [x] R014-2: Guard priorLoad/priorMinutes accumulation in ValidateCandidates; emit arithmetic_overflow violation when prior totals exceed float64 range
- [x] R014-3: Add ValidateWeekConstraints preflight check for completed+fixed overflow; tests for all three overflow paths

<!-- R013 revision items -->
- [x] R013-1: Detect float64 overflow in Reconcile; change to (Reconciliation, error); in ValidateCandidates use WarnArithmeticOverflow on overflow; test overflow JSON-marshal safety

<!-- R012 revision items -->
- [x] R012-1: ValidateCandidate must also emit ViolationRequestedSessionCountExceeded when RequestedSessionCount is pointer-to-0

<!-- R011 revision items -->
- [x] R011-1: Add WarnRequestedSessionCountUnmet batch warning when validCount < *RequestedSessionCount (underfill detection)
- [x] R011-2: Fix QF1001 lint issues at constraints.go:802,805 (De Morgan's law)

<!-- R010 revision items -->
- [x] R010-1: Reconciliation target fields use *float64 to preserve nil vs 0 distinction; RemainingLoad/RemainingMin are nil when untracked; fix nil+nonzero completed calculation
- [x] R010-2: Fix design-doc batch-validation and boundary-condition sections to distinguish nil (no cap) from pointer-to-0 (hard zero)

<!-- R009 revision items -->
- [x] R009-1: Change WeeklyTargetLoad, WeeklyTargetMinutes, RequestedSessionCount to *float64/*int (nil=no cap; 0=explicit zero)
- [x] R009-2: Write constraints_test.go with boundary tests (zero/nil, matching, reconcile NaN, all Step 3 cases)
- [x] R009-3: Fix Reconciliation table in design doc — CandidateLoad/CandidateMinutes are valid-input only

<!-- R008 revision items -->
- [x] R008-1: Replace greedy first-fit slot assignment with augmenting-path bipartite matching so all feasible day schedules are accepted regardless of candidate order
- [x] R008-2: Fix Reconcile to skip invalid-input (NaN/negative) candidates, same as ValidateCandidates

<!-- R007 revision items -->
- [x] R007-1: ValidateCandidates must treat MaxSessionsPerDay==0 as day_unavailable (same as ValidateCandidate), before calling validateAgainstDay
- [x] R007-2: Fix WarnInfeasibleLoad exported comment and design-doc table entry — say "valid-input candidates only", not "including invalid candidates"

<!-- R006 revision items -->
- [x] R006-1: ValidateCandidate single-candidate path must also sanitize embedded candidate for NaN-safe serialization
- [x] R006-2: ValidateWeekConstraints must parse and validate WeekStartDate (Monday), day dates (YYYY-MM-DD), and each day within the declared week
- [x] R006-3: Update design doc to match R005 isolation: invalid candidates excluded from numeric accumulations (not pessimistic); add invalid_input to codes table; fix "echoed verbatim" note; clean R001 artifact whitespace

<!-- R005 revision items -->
- [x] R005-1: Invalid candidates (NaN/negative) must not pollute priorLoad/priorMinutes or reconciliation; sanitize embedded Candidate to avoid JSON NaN failure
- [x] R005-2: Align docs — zero-budget case fires warning + violation (hard block), not warning-only; update Key Invariant 2, Validation Logic, example
- [x] R005-3: ValidateWeekConstraints — replace map iteration with ordered field checks for deterministic error ordering

<!-- R004 revision items -->
- [x] R004-1: Exhausted budgets (remaining <= 0) must also violate when candidate.Load > 0 / candidate.DurationMinutes > 0 (not just warn)
- [x] R004-2: Add input validation — finite non-negative candidate.Load and DurationMinutes; ValidateWeekConstraints() for constraint struct validation; ViolationInvalidInput code
- [x] R004-3: Fix duplicate-date inconsistency — ValidateCandidates must use first-match (same as findDay); add duplicate detection to ValidateWeekConstraints

<!-- R003 revision items -->
- [x] R003-1: Add ViolationRequestedSessionCountExceeded code; enforce RequestedSessionCount cap in ValidateCandidates (reject excess valid candidates)
- [x] R003-2: Implement slot consumption in ValidateCandidates (each slot holds one session; consumed slots unavailable to subsequent candidates)
- [x] R003-3: Fix slot violation semantics to use universal (all-slots) aggregation; add ViolationNoCompatibleSlot fallback for mixed-reason rejections

<!-- R002 revision items -->
- [x] R002-1: Resolve WarnZeroRemainingLoad trigger inconsistency (code vs design doc boundary table — remove "with Load > 0" qualifier)
- [x] R002-2: Add WarnZeroRemainingTime parallel handling when remainingMin <= 0 (symmetric with load case)
- [x] R002-3: Add "Field semantics / Units" section to design doc documenting float64-minutes rationale
- [x] R002-A: Document batch accumulation behavior (all candidates, not just valid, consume weekly budget)
- [x] R002-B: Note in design doc that WarnInfeasibleLoad uses all candidates including invalid ones
- [x] R002-C: Note in design doc that availableSlotCount is structural capacity, not candidate-filtered
- [x] R002-D: Update STATUS.md Step 2 to reflect implementation already present

- [x] Weekly constraint struct defined: WeekTarget (full-week target, remaining target, completed load, fixed load), RequestedSessionCount, AvailableDays (per-day slot list)
- [x] Daily slot struct defined: date, max sessions per day, slots with per-slot duration cap, indoor/outdoor cap, sport allow-list, mode allow-list
- [x] Candidate session struct defined: sport, mode, indoor/outdoor flag, proposed duration, proposed load
- [x] Result codes defined: ViolationCode enum covering daily-cap, slot-duration-cap, indoor-cap, session-count-cap, weekly-load-overshoot, infeasible-slots
- [x] ReconciliationResult struct defined: completed, fixed, candidate, remaining, projected totals
- [x] Design doc `docs/design/plan-filler-constraints.md` written with field semantics, invariants, result codes, and examples
- [x] Availability/session-count distinction documented (availability = where, requested = how many)

---

### Step 2: Implement validation and reconciliation

**Status:** ✅ Complete

- [x] Daily, session, mode, sport, and weekly constraints validated
- [x] Weekly time/load reconciliation implemented
- [x] Infeasible requests reported explicitly
- [x] Package remains pure and write-free

---

### Step 3: Add boundary-focused regression coverage

**Status:** ✅ Complete

<!-- R016 revision items -->
- [x] R016-1: Ensure the targeted Step 3 command selects every boundary regression group and verify its selection (`go test -list '^(TestConstraint|TestReconciliation)'` selected all eight Step 3 tests)
- [x] R016-2: Assert reconciliation remaining/projected load and minutes in in-progress-week and fixed-event coverage
- [x] R016-3: Exercise batch validation for slots and requested-count/infeasibility, including ordered results and warnings

- [x] In-progress week overshoot covered
- [x] Separate 45-minute slots versus 95-minute session covered
- [x] Indoor versus outdoor cap covered
- [x] Fixed, zero-load, unavailable, and infeasible cases covered

---

### Step 4: Align roadmap and product contract

**Status:** 🟨 In Progress

<!-- R018 revision items -->
- [ ] R018-1: Add explicit v2.0 acceptance criteria for the future Plan Filler constraint validator without claiming its tool/write path is shipped
- [ ] R018-2: Record the PRD/changelog review and no-change rationale for this internal unregistered validator
- [ ] R018-3: Verify roadmap/design boundary wording with a focused documentation diff review

- [ ] v2.0 acceptance criteria updated
- [ ] v2.2 boundary clarified
- [ ] PRD and changelog reviewed for user-visible impact

---

### Step 5: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Race suite passing
- [ ] Lint passing
- [ ] All failures fixed
- [ ] Build passes
- [ ] Formatting and docs diff clean

---

### Step 6: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 13:51 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 13:51 | Step 0 started | Preflight |
| 2026-07-10 15:45 | Worker iter 1 | done in 6871s, tools: 266 |
| 2026-07-10 15:52 | Worker iter 2 | done in 414s, tools: 34 |
| 2026-07-10 15:52 | Step 3 started | Add boundary-focused regression coverage |

## Blockers

*None*

## Notes

*Reserved for execution notes*
| 2026-07-10 13:57 | Review R001 | plan Step 1: APPROVE |
| 2026-07-10 14:10 | Review R002 | code Step 1: REVISE |
| 2026-07-10 14:16 | Review R003 | code Step 1: REVISE |
| 2026-07-10 14:25 | Review R004 | code Step 1: REVISE |
| 2026-07-10 14:32 | Review R005 | code Step 1: REVISE |
| 2026-07-10 14:38 | Review R006 | code Step 1: REVISE |
| 2026-07-10 14:45 | Review R007 | code Step 1: REVISE |
| 2026-07-10 14:48 | Review R008 | code Step 1: REVISE |
| 2026-07-10 14:58 | Review R009 | code Step 1: REVISE |
| 2026-07-10 15:08 | Review R010 | code Step 1: REVISE |
| 2026-07-10 15:14 | Review R011 | code Step 1: REVISE |
| 2026-07-10 15:18 | Review R012 | code Step 1: REVISE |
| 2026-07-10 15:22 | Review R013 | code Step 1: REVISE |
| 2026-07-10 15:29 | Review R014 | code Step 1: REVISE |
| 2026-07-10 16:00 | Review R015 | code Step 1: UNAVAILABLE |
| 2026-07-10 15:54 | Review R016 | plan Step 3: REVISE |
| 2026-07-10 15:59 | Review R017 | code Step 3: APPROVE |
| 2026-07-10 16:01 | Review R018 | plan Step 4: REVISE |

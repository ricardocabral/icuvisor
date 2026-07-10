# TP-235: Add plan-filler constraint model and validator — Status

**Current Step:** Step 1: Define the constraint contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 9
**Iteration:** 1
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing planning semantics reviewed

---

### Step 1: Define the constraint contract

**Status:** 🟨 In Progress

<!-- R009 revision items -->
- [ ] R009-1: Change WeeklyTargetLoad, WeeklyTargetMinutes, RequestedSessionCount to *float64/*int (nil=no cap; 0=explicit zero)
- [ ] R009-2: Write constraints_test.go with boundary tests (zero/nil, matching, reconcile NaN, all Step 3 cases)
- [ ] R009-3: Fix Reconciliation table in design doc — CandidateLoad/CandidateMinutes are valid-input only

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

**Status:** 🟨 In Progress

- [x] Daily, session, mode, sport, and weekly constraints validated
- [x] Weekly time/load reconciliation implemented
- [x] Infeasible requests reported explicitly
- [x] Package remains pure and write-free

---

### Step 3: Add boundary-focused regression coverage

**Status:** ⬜ Not Started

- [ ] In-progress week overshoot covered
- [ ] Separate 45-minute slots versus 95-minute session covered
- [ ] Indoor versus outdoor cap covered
- [ ] Fixed, zero-load, unavailable, and infeasible cases covered

---

### Step 4: Align roadmap and product contract

**Status:** ⬜ Not Started

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

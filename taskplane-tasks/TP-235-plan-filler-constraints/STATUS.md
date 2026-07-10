# TP-235: Add plan-filler constraint model and validator — Status

**Current Step:** Step 1: Define the constraint contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 3
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

<!-- R003 revision items -->
- [ ] R003-1: Add ViolationRequestedSessionCountExceeded code; enforce RequestedSessionCount cap in ValidateCandidates (reject excess valid candidates)
- [ ] R003-2: Implement slot consumption in ValidateCandidates (each slot holds one session; consumed slots unavailable to subsequent candidates)
- [ ] R003-3: Fix slot violation semantics to use universal (all-slots) aggregation; add ViolationNoCompatibleSlot fallback for mixed-reason rejections

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

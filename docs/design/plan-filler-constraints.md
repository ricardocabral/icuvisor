# Plan-Filler Constraint Model and Validator

**Status:** In progress — constraint model implemented, `fill_calendar_from_library` tool not yet shipped.
**Related roadmap item:** v2.0 Plan Filler (`fill_calendar_from_library`)
**Package:** `internal/planning`

---

## Field Semantics and Units

All duration fields (`MaxDurationMinutes`, `MaxIndoorMinutes`, `MaxTotalDailyMinutes`, `WeeklyTargetMinutes`, `CompletedMinutes`, `FixedMinutes`, `CandidateMinutes`, `RemainingMinutes`, `ProjectedMinutes`, and `DurationMinutes` in `CandidateSession`) use **`float64` representing minutes**.

`time.Duration` (nanoseconds) was not used because:
- This domain operates on human-scale planning intervals (30 min, 90 min, 2 h) where `float64` minutes are natural, readable, and free of unit-conversion error when multiplied or compared.
- All JSON-serialisable output fields use `float64` for forward-compatible transport; `time.Duration` (an `int64`) would require custom marshalling.
- The constraint model has no sub-minute precision requirement; `float64` minutes carry sufficient accuracy for the target use case.

The convention is uniform across all four input structs (`SlotConstraint`, `DayConstraints`, `WeekConstraints`, `CandidateSession`) and all `Reconciliation` fields.

Load fields (`WeeklyTargetLoad`, `CompletedLoad`, `FixedLoad`, `CandidateLoad`, `RemainingLoad`, `ProjectedLoad`, and `Load` in `CandidateSession`) use **`float64` representing TSS/ATL load points** — the same unit convention used by `propose_annual_training_plan` and `get_fitness_projection`.

---

## Purpose

The planning constraint package provides a deterministic, pure validator for the plan-filler scheduling domain. It answers one question before any calendar writes happen:

> "Given these week-level targets, daily availability windows, and per-slot caps, is this candidate session placeable — and does placing the full batch leave the weekly load within bounds?"

The package contains no intervals.icu client calls, no calendar writes, no workout selection, no model inference, and no physiology classification. All inputs are caller-supplied numeric structs; free-text instructions are never treated as hard constraints.

---

## Key Invariants

1. **Availability ≠ requested sessions.** `AvailableDays` encodes *where* sessions may be placed. `RequestedSessionCount` encodes *how many* sessions the caller wants placed. Having five available days does not imply five sessions are requested, and requesting three sessions does not create availability on days absent from `AvailableDays`.

2. **Remaining target ≠ full-week target.** For an in-progress week, the validator uses `WeeklyTargetLoad - CompletedLoad - FixedLoad` (the *remaining* budget) as the scheduling cap. The full-week target is preserved for reporting but is never used as the overshoot threshold for new sessions. This prevents filling an in-progress week to its full original target when part of the week has already passed.

3. **Slots do not combine.** A candidate session must fit within exactly one slot. Two 45-minute slots cannot accommodate a 95-minute session. Sessions do not span slot boundaries.

4. **Indoor and outdoor caps are independent.** A slot can declare a `MaxIndoorMinutes` that is shorter than its `MaxDurationMinutes`. An outdoor session of 90 minutes in a slot with `MaxDurationMinutes=120` and `MaxIndoorMinutes=60` is valid; the same duration as an indoor session violates the indoor cap only.

5. **Fixed events reduce remaining budget, not availability.** Races, A-priority events, and unavailable blocks are represented as `FixedMinutes`/`FixedLoad` in `WeekConstraints`. They are subtracted from the weekly target to compute the remaining budget but are not candidates themselves and do not occupy `AvailableDays` slots.

6. **Exhausted budgets block positive work.** If the remaining load or time budget is zero or negative (completed + fixed already meet the target), the validator emits a `zero_remaining_load` / `zero_remaining_time` warning to explain the state, AND emits a `weekly_load_overshoot` / `weekly_time_overshoot` violation for any candidate that would add positive load or duration. A zero-load or zero-duration candidate (e.g. mobility notes) does not receive the violation but still receives the warning. Deficits and overruns are never redistributed or silently absorbed.

---

## Data Model

### `WeekConstraints`

Top-level container for one calendar week.

| Field | Type | Description |
|---|---|---|
| `WeekStartDate` | `string` | Athlete-local Monday, YYYY-MM-DD. |
| `WeeklyTargetMinutes` | `*float64` | Full-week training-time target. Nil means no time budget is tracked. Pointer-to-0 means an explicit zero time budget (all positive-duration candidates are blocked). |
| `WeeklyTargetLoad` | `*float64` | Full-week training-load target. Nil means no load budget is tracked. Pointer-to-0 is an explicit zero load budget. |
| `CompletedMinutes` | `float64` | Already-logged training time (read-only past data). |
| `CompletedLoad` | `float64` | Already-logged training load (read-only past data). |
| `FixedMinutes` | `float64` | Committed future time from locked events (races, etc.). |
| `FixedLoad` | `float64` | Committed future load from locked events. |
| `RequestedSessionCount` | `*int` | How many sessions the caller wants placed. Nil means no session-count cap. Pointer-to-0 means zero sessions (all candidates are excess). |
| `AvailableDays` | `[]DayConstraints` | Days where sessions may be placed; absent days are unavailable. |

### `DayConstraints`

One available calendar day within the week.

| Field | Type | Description |
|---|---|---|
| `Date` | `string` | Athlete-local date, YYYY-MM-DD. |
| `MaxSessionsPerDay` | `int` | Hard cap on sessions placed on this day. Zero means unavailable. |
| `MaxTotalDailyMinutes` | `float64` | Combined duration cap for all sessions on this day. Zero means uncapped. |
| `Slots` | `[]SlotConstraint` | Independent training windows for the day. |

### `SlotConstraint`

One training window within a day. Sessions must fit in exactly one slot.

| Field | Type | Description |
|---|---|---|
| `MaxDurationMinutes` | `float64` | Per-session duration cap. Zero means uncapped. |
| `MaxIndoorMinutes` | `float64` | Indoor-specific duration cap. Zero means no indoor cap. Outdoor sessions are unaffected. |
| `AllowedSports` | `[]string` | Permitted sports. Empty means any sport. |
| `AllowedModes` | `[]string` | Permitted training modes. Empty means any mode. |

### `CandidateSession`

A proposed training session to validate.

| Field | Type | Description |
|---|---|---|
| `Date` | `string` | Proposed athlete-local date, YYYY-MM-DD. |
| `Sport` | `string` | Training discipline (e.g. "Ride", "Run", "Swim"). |
| `Mode` | `string` | Training mode (e.g. "EnduranceRide", "Intervals"). |
| `Indoor` | `bool` | Indoor trainer, treadmill, pool, or similar. |
| `DurationMinutes` | `float64` | Proposed session length. |
| `Load` | `float64` | Proposed training load contribution (e.g. TSS). |

---

## Result Types

### `Reconciliation`

Computed weekly totals — no rounding, no redistribution.

| Field | Description |
|---|---|
| `WeeklyTargetMinutes` | `*float64`: mirrors `WeekConstraints.WeeklyTargetMinutes`. Absent from JSON (`omitempty`) when nil (untracked). Present when set, including a pointer-to-zero. |
| `WeeklyTargetLoad` | `*float64`: mirrors `WeekConstraints.WeeklyTargetLoad`. Same nil/zero semantics. |
| `CompletedMinutes` | Completed time (from input). |
| `CompletedLoad` | Completed load (from input). |
| `FixedMinutes` | Fixed future time (from input). |
| `FixedLoad` | Fixed future load (from input). |
| `CandidateMinutes` | Sum of `DurationMinutes` for valid-input candidates only (NaN/negative excluded). |
| `CandidateLoad` | Sum of `Load` for valid-input candidates only (NaN/negative excluded). |
| `RemainingMinutes` | `*float64`: `*WeeklyTargetMinutes - CompletedMinutes - FixedMinutes`. Nil when time is untracked. Non-nil and possibly negative when tracked. Absent from JSON when nil. |
| `RemainingLoad` | `*float64`: `*WeeklyTargetLoad - CompletedLoad - FixedLoad`. Nil when load is untracked. Absent from JSON when nil. |
| `ProjectedMinutes` | `CompletedMinutes + FixedMinutes + CandidateMinutes`. |
| `ProjectedLoad` | `CompletedLoad + FixedLoad + CandidateLoad`. |

### `CandidateResult`

Outcome for one candidate.

| Field | Description |
|---|---|
| `Candidate` | The input candidate, echoed as-provided. For invalid-input results (`invalid_input` violation), non-finite or negative `DurationMinutes`/`Load` values are replaced by 0 to ensure the result is JSON-serializable. |
| `Valid` | True if and only if `Violations` is empty. |
| `Violations` | Hard constraint breaches (see codes below). |
| `Warnings` | Soft concerns (see codes below). |

### `BatchResult`

Outcome for all candidates in a week.

| Field | Description |
|---|---|
| `Results` | One `CandidateResult` per input candidate, in order. |
| `Warnings` | Week-level soft concerns (e.g. infeasible session count). |
| `Reconciliation` | Computed weekly totals for all candidates combined. |

---

## Result Codes

### Violation Codes (hard — block placement)

| Code | Trigger |
|---|---|
| `invalid_input` | Candidate `DurationMinutes` or `Load` is non-finite (NaN, Inf) or negative. Checked before all other constraints. The invalid value is excluded from budget accumulation and reconciliation totals. The embedded `Candidate` has the invalid field replaced by 0 to ensure JSON-safe output. |
| `day_unavailable` | Candidate date absent from `AvailableDays` or `MaxSessionsPerDay` is zero. |
| `daily_session_count_exceeded` | Adding the candidate would exceed `MaxSessionsPerDay`. |
| `daily_time_exceeded` | Combined daily duration would exceed `MaxTotalDailyMinutes`. |
| `slot_duration_exceeded` | Candidate duration exceeds EVERY available (unconsumed) slot's `MaxDurationMinutes`. Only emitted when the reason is universal across all slots. |
| `indoor_duration_exceeded` | Indoor candidate duration exceeds EVERY available slot's `MaxIndoorMinutes`. Only emitted when universal. |
| `sport_not_allowed` | Candidate sport is excluded by EVERY available slot's `AllowedSports`. Only emitted when universal. |
| `mode_not_allowed` | Candidate mode is excluded by EVERY available slot's `AllowedModes`. Only emitted when universal. |
| `no_compatible_slot` | No available slot fits the candidate and no single constraint is universal across all slots (mixed-reason rejection, e.g. slot A rejects for duration, slot B for sport). |
| `no_available_slot` | All slots for the day have been consumed by prior candidates in a batch pass; `MaxSessionsPerDay` has not yet been reached but no slots remain. |
| `weekly_load_overshoot` | Candidate load exceeds the remaining weekly load budget. |
| `weekly_time_overshoot` | Candidate duration exceeds the remaining weekly time budget. |
| `requested_session_count_exceeded` | This candidate would be the (N+1)th valid session; `RequestedSessionCount` is N. Position in the batch determines priority. |

### Warning Codes (soft — caller attention required)

| Code | Trigger |
|---|---|
| `infeasible_session_count` | `RequestedSessionCount` exceeds total structural slots across all days. Structural capacity only; does not filter by sport or mode. |
| `requested_session_count_unmet` | Batch contains fewer valid sessions than `*RequestedSessionCount` (underfill). Includes `requested` and `accepted` counts in the value field. |
| `infeasible_load` | Total load of valid-input candidates is less than the remaining weekly load target. Invalid-input candidates (NaN/negative) are excluded from this check. |
| `zero_remaining_load` | Remaining load budget is zero or negative. Fires unconditionally when remaining ≤ 0. Accompanied by `weekly_load_overshoot` when `candidate.Load > 0`. |
| `zero_remaining_time` | Remaining time budget is zero or negative. Parallel to `zero_remaining_load`. Accompanied by `weekly_time_overshoot` when `candidate.DurationMinutes > 0`. |

---

## Validation Logic

### Single-candidate validation (`ValidateCandidate`)

1. **Day check:** find the `DayConstraints` for the candidate date; return `day_unavailable` if absent or `MaxSessionsPerDay == 0`.
2. **Daily session count:** if `sessionsAlreadyOnDay >= MaxSessionsPerDay`, emit `daily_session_count_exceeded`.
3. **Combined daily duration:** if `MaxTotalDailyMinutes > 0` and the new total would exceed it, emit `daily_time_exceeded`.
4. **Slot matching:** find the first available slot where all constraints pass (duration, indoor cap, sport, mode). If none found, emit violation codes for constraints that are universally violated (every slot rejects for the same reason). If no reason is universal (slot A rejects for duration, slot B for sport), emit `no_compatible_slot` as the deterministic fallback.
5. **Weekly load:** compute `remainingLoad = WeeklyTargetLoad - CompletedLoad - FixedLoad - priorLoad`. If `remainingLoad ≤ 0`: emit `zero_remaining_load` warning unconditionally; if `candidate.Load > 0`, also emit `weekly_load_overshoot` (hard block — session cannot be placed). If `remainingLoad > 0` and `candidate.Load > remainingLoad`: emit `weekly_load_overshoot`.
6. **Weekly time:** compute `remainingMin = WeeklyTargetMinutes - CompletedMinutes - FixedMinutes - priorMinutes`. If `remainingMin ≤ 0`: emit `zero_remaining_time` warning unconditionally; if `candidate.DurationMinutes > 0`, also emit `weekly_time_overshoot` (hard block). If `remainingMin > 0` and `candidate.DurationMinutes > remainingMin`: emit `weekly_time_overshoot`.

### Batch validation (`ValidateCandidates`)

Processes candidates in order, maintaining per-day session counts, per-day cumulative duration, and accumulated weekly load/time from prior candidates.

**Slot assignment uses maximum bipartite matching.** For each day, a maximum bipartite matching is pre-computed across all valid-input candidates on that day before per-candidate results are emitted. This ensures feasible schedules are accepted regardless of candidate input order: given slots [any-sport 60 min, Ride-only 60 min] and candidates [Ride 60 min, Run 60 min], the matching finds {Ride → Ride-only, Run → any-sport} regardless of which candidate appears first in the input. Matching uses augmenting-path DFS in candidate input order for determinism when multiple maximum matchings exist.

A candidate that is not in the matching gets slot violations:
- `no_available_slot` if the candidate fits at least one slot's constraints but all compatible slots are claimed by other matched candidates (contention).
- Specific constraint codes (`slot_duration_exceeded`, `sport_not_allowed`, etc.) or `no_compatible_slot` if the candidate does not fit any slot at all.

**RequestedSessionCount cap:** `RequestedSessionCount` is a `*int`. When nil, no session-count cap is enforced. When non-nil, once `*RequestedSessionCount` violation-free candidates have been accepted, subsequent valid candidates receive `requested_session_count_exceeded`. A pointer-to-zero means zero sessions are wanted — all candidates are immediately excess regardless of availability.

**Invalid-input candidates are isolated.** A candidate with `invalid_input` (NaN/negative numeric inputs) increments the day session counter (for deterministic positional tracking) but is excluded from all numeric accumulations: its `DurationMinutes` and `Load` are not added to the per-day minute total, `priorLoad`/`priorMinutes`, or reconciliation sums. This prevents non-finite values from poisoning comparisons or serialized output for subsequent candidates. Valid-input candidates (even those with other violations) follow the pessimistic accumulation rules below.

**Accumulation is pessimistic for valid-input candidates.** All valid-input candidates (valid or not) increment the day session counter and the per-day minute total, and consume the weekly `priorLoad`/`priorMinutes` budget. This ensures deterministic, position-based rejection: if sessions 1 and 2 are proposed for a day with `MaxSessionsPerDay: 1`, session 2 receives `daily_session_count_exceeded` regardless of whether session 1 is valid for other reasons. Weekly `priorLoad` and `priorMinutes` are accumulated from all valid-input candidates, so budget checks for candidate N account for the full proposed load of candidates 1…N-1.

The `WarnInfeasibleLoad` check and `Reconciliation` sums use only valid-input candidates. A batch containing only invalid-input sessions will have zero `CandidateLoad` and `CandidateMinutes` in the reconciliation.

**Slot-count feasibility** (`WarnInfeasibleSessionCount`) compares `RequestedSessionCount` against structural slot capacity (sum of `min(MaxSessionsPerDay, len(Slots))` across available days). This count does not filter by sport or mode; a "Swim" candidate against a Ride-only slot would still count as a structural slot.

After individual validation, adds week-level warnings:
- `infeasible_session_count` if `RequestedSessionCount > structuralSlotCount`.
- `infeasible_load` if total candidate load cannot satisfy the remaining load budget.

### Reconciliation (`Reconcile`)

Computes totals arithmetically. Does not validate. Can be called independently of validation for reporting purposes.

---

## Examples

### In-progress week with partial completion

```
WeeklyTargetLoad: 300      CompletedLoad: 120     FixedLoad: 0
→ RemainingLoad = 180

Candidate(Load: 200) → ViolationWeeklyLoadOvershoot (200 > 180)
Candidate(Load: 150) → Valid (150 ≤ 180)
```

### Two separate 45-minute slots

```
Day slots: [{MaxDurationMinutes: 45}, {MaxDurationMinutes: 45}]

Candidate(DurationMinutes: 95) → ViolationSlotDuration  (no single slot holds 95 min)
Candidate(DurationMinutes: 45) → Valid                   (fits in either slot)
```

### Indoor vs outdoor cap

```
Slot: {MaxDurationMinutes: 120, MaxIndoorMinutes: 60}

Candidate(DurationMinutes: 90, Indoor: false) → Valid              (outdoor, 90 ≤ 120)
Candidate(DurationMinutes: 90, Indoor: true)  → ViolationIndoorDuration (indoor, 90 > 60)
```

### Fixed events and remaining budget

```
WeeklyTargetLoad: 400    FixedLoad: 150    CompletedLoad: 0
→ RemainingLoad = 250

Candidate(Load: 300) → ViolationWeeklyLoadOvershoot (300 > 250)
Candidate(Load: 250) → Valid
```

### Zero remaining load

```
WeeklyTargetLoad: 300    CompletedLoad: 200    FixedLoad: 150
→ RemainingLoad = -50

Candidate(Load: 50) → Warning: zero_remaining_load  (budget exhausted)
                   → Violation: weekly_load_overshoot (Load > 0, cannot be placed)
                   Valid: false

Candidate(Load: 0)  → Warning: zero_remaining_load  (budget exhausted)
                   Valid: true  (zero-load session, e.g. mobility notes)
```

### Infeasible session count

```
RequestedSessionCount: 5
AvailableDays: [{MaxSessionsPerDay: 1, Slots: [...]}, {MaxSessionsPerDay: 1, Slots: [...]}]
→ totalAvailableSlots = 2

BatchResult.Warnings: [infeasible_session_count (requested 5, have 2 slots)]
```

---

## Input Validation

### `ValidateWeekConstraints(wc WeekConstraints) error`

Call before `ValidateCandidate` or `ValidateCandidates` to ensure the constraint struct is sound. Returns a non-nil error for:
- `WeekStartDate` absent, non-YYYY-MM-DD, or not a Monday.
- Any `AvailableDays.Date` that is non-YYYY-MM-DD or falls outside the Monday–Sunday week declared by `WeekStartDate`.
- Any `float64` field that is NaN, infinite, or negative (targets, completed, fixed, slot caps, daily caps).
- `RequestedSessionCount < 0`.
- Duplicate `Date` values in `AvailableDays`.

Errors are reported in a fixed deterministic order (date fields first, then numeric fields, then day entries in list order).

Constraint struct errors are returned as `error` rather than `Violation` because they represent a programming error in how the caller built the struct, not a property of the candidate session.

### Candidate input validation

`ValidateCandidate` and `ValidateCandidates` check each candidate for finite, non-negative `DurationMinutes` and `Load` before applying any other constraints. Invalid inputs return `ViolationInvalidInput` immediately. This prevents NaN/negative values from propagating into comparisons, budget accumulations, or `Reconciliation` fields.

---

## Boundary Conditions

| Condition | Behaviour |
|---|---|
| `WeeklyTargetLoad == nil` | No load budget is tracked. `weekly_load_overshoot`, `zero_remaining_load`, and `WarnInfeasibleLoad` are never emitted. `Reconciliation.WeeklyTargetLoad` and `Reconciliation.RemainingLoad` are absent from JSON output. |
| `WeeklyTargetLoad == pointer-to-0` | Explicit zero load budget. All candidates with `Load > 0` receive `zero_remaining_load` warning AND `weekly_load_overshoot` violation. |
| `WeeklyTargetMinutes == nil` | No time budget is tracked. Time overshoot checks are skipped. `Reconciliation.WeeklyTargetMinutes` and `Reconciliation.RemainingMinutes` are absent from JSON output. |
| `WeeklyTargetMinutes == pointer-to-0` | Explicit zero time budget. All candidates with `DurationMinutes > 0` receive `zero_remaining_time` warning AND `weekly_time_overshoot` violation. |
| `RequestedSessionCount == nil` | No session-count cap. All valid candidates are accepted regardless of count. |
| `RequestedSessionCount == pointer-to-0` | Zero sessions requested. Every candidate receives `requested_session_count_exceeded`. |
| `FixedLoad + CompletedLoad > *WeeklyTargetLoad` | RemainingLoad is negative; `zero_remaining_load` warning fires unconditionally. If `candidate.Load > 0`, `weekly_load_overshoot` also fires. |
| `FixedMinutes + CompletedMinutes > *WeeklyTargetMinutes` | RemainingMinutes is negative; `zero_remaining_time` warning fires unconditionally. If `candidate.DurationMinutes > 0`, `weekly_time_overshoot` also fires. |
| `MaxSessionsPerDay == 0` | Day is treated as unavailable; `day_unavailable` fires. |
| `len(Slots) == 0` | Slot constraints are skipped; only day-level and weekly checks apply. Sessions per day are capped only by `MaxSessionsPerDay`. |
| All slots consumed by prior candidates | `no_available_slot` fires even if `MaxSessionsPerDay` not yet reached. |
| `MaxDurationMinutes == 0` on a slot | Duration cap is uncapped for that slot. |
| `MaxIndoorMinutes == 0` on a slot | No indoor cap applies for that slot. |
| Empty `AllowedSports` on a slot | Any sport is allowed for that slot. |
| Empty `AllowedModes` on a slot | Any mode is allowed for that slot. |

---

## Scope Boundaries

This package validates constraints. It does not:
- Select workouts from a library.
- Write calendar events.
- Infer athlete physiology, age, recovery time, or training history.
- Apply ramp-rate or intensity-distribution guardrails (those are v2.2 scope).
- Fetch data from intervals.icu or any upstream service.
- Accept or enforce free-text instructions as hard constraints.

The v2.2 science-backed guardrails (ramp rate, recovery cadence, intensity distribution) are a separate concern that builds on top of this constraint layer. This package deliberately has no knowledge of them.

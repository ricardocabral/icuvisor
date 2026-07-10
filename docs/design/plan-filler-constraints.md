# Plan-Filler Constraint Model and Validator

**Status:** In progress — constraint model implemented, `fill_calendar_from_library` tool not yet shipped.
**Related roadmap item:** v2.0 Plan Filler (`fill_calendar_from_library`)
**Package:** `internal/planning`

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

6. **Deficits are never redistributed silently.** If the remaining load budget is zero or negative (completed + fixed already meet the target), the validator emits a `zero_remaining_load` warning and lets the caller decide. It never invents availability or adjusts the target downward to accommodate a request.

---

## Data Model

### `WeekConstraints`

Top-level container for one calendar week.

| Field | Type | Description |
|---|---|---|
| `WeekStartDate` | `string` | Athlete-local Monday, YYYY-MM-DD. |
| `WeeklyTargetMinutes` | `float64` | Full-week training-time target. |
| `WeeklyTargetLoad` | `float64` | Full-week training-load target (e.g. TSS). |
| `CompletedMinutes` | `float64` | Already-logged training time (read-only past data). |
| `CompletedLoad` | `float64` | Already-logged training load (read-only past data). |
| `FixedMinutes` | `float64` | Committed future time from locked events (races, etc.). |
| `FixedLoad` | `float64` | Committed future load from locked events. |
| `RequestedSessionCount` | `int` | How many sessions the caller wants placed. May exceed available slots (produces a warning). |
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
| `WeeklyTargetMinutes` | Full-week target (from input). |
| `WeeklyTargetLoad` | Full-week load target (from input). |
| `CompletedMinutes` | Completed time (from input). |
| `CompletedLoad` | Completed load (from input). |
| `FixedMinutes` | Fixed future time (from input). |
| `FixedLoad` | Fixed future load (from input). |
| `CandidateMinutes` | Sum of all candidate durations. |
| `CandidateLoad` | Sum of all candidate loads. |
| `RemainingMinutes` | `WeeklyTargetMinutes - CompletedMinutes - FixedMinutes`. Scheduling budget. May be negative. |
| `RemainingLoad` | `WeeklyTargetLoad - CompletedLoad - FixedLoad`. Load budget. May be negative. |
| `ProjectedMinutes` | `CompletedMinutes + FixedMinutes + CandidateMinutes`. |
| `ProjectedLoad` | `CompletedLoad + FixedLoad + CandidateLoad`. |

### `CandidateResult`

Outcome for one candidate.

| Field | Description |
|---|---|
| `Candidate` | The input candidate, echoed verbatim. |
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
| `day_unavailable` | Candidate date absent from `AvailableDays` or `MaxSessionsPerDay` is zero. |
| `daily_session_count_exceeded` | Adding the candidate would exceed `MaxSessionsPerDay`. |
| `daily_time_exceeded` | Combined daily duration would exceed `MaxTotalDailyMinutes`. |
| `slot_duration_exceeded` | Candidate duration exceeds every available slot's `MaxDurationMinutes`. |
| `indoor_duration_exceeded` | Indoor candidate duration exceeds every slot's `MaxIndoorMinutes`. |
| `sport_not_allowed` | Candidate sport is excluded by every slot's `AllowedSports` list. |
| `mode_not_allowed` | Candidate mode is excluded by every slot's `AllowedModes` list. |
| `weekly_load_overshoot` | Candidate load exceeds the remaining weekly load budget. |
| `weekly_time_overshoot` | Candidate duration exceeds the remaining weekly time budget. |

### Warning Codes (soft — caller attention required)

| Code | Trigger |
|---|---|
| `infeasible_session_count` | `RequestedSessionCount` exceeds total available slots across all days. |
| `infeasible_load` | Candidate load total is less than the remaining weekly load target. |
| `zero_remaining_load` | Remaining load budget is zero or negative; completed + fixed events already meet the target. |

---

## Validation Logic

### Single-candidate validation (`ValidateCandidate`)

1. **Day check:** find the `DayConstraints` for the candidate date; return `day_unavailable` if absent or `MaxSessionsPerDay == 0`.
2. **Daily session count:** if `sessionsAlreadyOnDay >= MaxSessionsPerDay`, emit `daily_session_count_exceeded`.
3. **Combined daily duration:** if `MaxTotalDailyMinutes > 0` and the new total would exceed it, emit `daily_time_exceeded`.
4. **Slot matching:** find one slot where all constraints pass. If none found, emit all applicable slot violation codes (`slot_duration_exceeded`, `indoor_duration_exceeded`, `sport_not_allowed`, `mode_not_allowed`).
5. **Weekly load/time:** compute remaining = target − completed − fixed − prior candidate totals. If remaining ≤ 0, emit `zero_remaining_load` warning. If candidate load > remaining, emit `weekly_load_overshoot`. If candidate duration > remaining minutes, emit `weekly_time_overshoot`.

### Batch validation (`ValidateCandidates`)

Processes candidates in order, maintaining per-day session counts, per-day cumulative duration, and accumulated weekly load/time from prior candidates. All proposed candidates (valid or not) increment the day counter — this ensures deterministic violation assignment when the daily cap would be exceeded by position in the batch.

After individual validation, adds week-level warnings:
- `infeasible_session_count` if `RequestedSessionCount > totalAvailableSlots`.
- `infeasible_load` if candidate totals cannot satisfy the remaining load budget.

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

Candidate(Load: 50) → Warning: zero_remaining_load
                      (target already met by completed + fixed)
```

### Infeasible session count

```
RequestedSessionCount: 5
AvailableDays: [{MaxSessionsPerDay: 1, Slots: [...]}, {MaxSessionsPerDay: 1, Slots: [...]}]
→ totalAvailableSlots = 2

BatchResult.Warnings: [infeasible_session_count (requested 5, have 2 slots)]
```

---

## Boundary Conditions

| Condition | Behaviour |
|---|---|
| `WeeklyTargetLoad == 0` | Load overshoot checks are skipped; no load violations or zero-load warnings. |
| `WeeklyTargetMinutes == 0` | Time overshoot checks are skipped. |
| `FixedLoad + CompletedLoad > WeeklyTargetLoad` | RemainingLoad is negative; `zero_remaining_load` warning fires for any candidate with Load > 0. |
| `MaxSessionsPerDay == 0` | Day is treated as unavailable; `day_unavailable` fires. |
| `len(Slots) == 0` | Slot constraints are skipped; only day-level and weekly checks apply. |
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

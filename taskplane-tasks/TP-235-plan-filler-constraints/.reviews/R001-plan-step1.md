# Plan Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** APPROVE

**Reviewed:** 2026-07-10  
**Step:** Step 1: Define the constraint contract  
**Artifacts planned:** `internal/planning/constraints.go` (new), `docs/design/plan-filler-constraints.md` (new)

---

## Summary

The plan is structurally sound. The struct decomposition maps correctly to the domain invariants stated in the PROMPT: availability as a slot list (where) vs requested session count (how many) are cleanly separated; the full-week target vs remaining-week target vs completed load vs fixed load are all modeled as distinct fields; and ViolationCode is the right mechanism for deterministic, caller-inspectable results. The package is correctly placed in `internal/` and declared pure. Step scope is appropriately narrow — type definitions and a design doc only, no implementation logic.

---

## Issues Requiring Attention in This Step

### 1. Duration unit must be decided and locked in the design doc

The plan names "per-slot duration cap" and "proposed duration" but does not specify their Go type or unit. The upstream `intervals` layer stores time in seconds (`*int`, e.g., `MovingTimeSeconds`, `TimeTarget`). Step 3 tests reference "45-minute slots" and "60-minute indoor cap", so minutes are the natural human unit for callers of this package. Seconds produce error-prone arithmetic (a 90-minute cap would be 5400 — easy to confuse). `time.Duration` is the idiomatic Go choice and avoids unit confusion entirely.

**Required action:** pick a concrete representation (recommend `time.Duration` for cap fields and candidate session duration), document it in the design doc under "Field semantics", and keep it consistent across all structs. Do not leave this as an implicit convention.

### 2. `ReconciliationResult` must track both load and duration

The PROMPT says: "Compute completed, fixed, candidate, remaining, and projected weekly **time/load** totals". The plan names five total fields (completed, fixed, candidate, remaining, projected) but omits which dimension each field belongs to. The design doc needs to make explicit that each of these totals exists in both dimensions — e.g., `CompletedDuration`, `CompletedLoad`, `FixedDuration`, `FixedLoad`, etc. — rather than a single undifferentiated numeric. If only one dimension is tracked, the design doc must state the deliberate choice and its rationale.

**Required action:** decide whether `ReconciliationResult` carries parallel load and duration fields for each bucket, or a single `TotalsByBucket` shape, and document this explicitly.

### 3. Violation vs warning distinction needs a concrete representation

The PROMPT distinguishes between hard violations ("a candidate exceeds a cap") and warnings ("requested load is infeasible within available slots"). The STATUS.md plan lists a single `ViolationCode` enum that includes `infeasible-slots`. This conflation is fine if the struct that wraps `ViolationCode` also carries a `Severity` field (`violation` vs `warning`), or if the type is renamed to `ResultCode` and the severity is embedded in the result struct. If `infeasible-slots` is always a warning while `slot-duration-cap` is always a violation, document that mapping.

**Required action:** the design doc must specify the severity of each code, even if the type system collapses them into one enum.

---

## Notes (Non-blocking)

- **`go test ./internal/planning` at end of Step 1** will be a compile-only check because no `_test.go` file exists at this stage. That is fine and expected. The step description could note "verifies package compiles" to avoid confusion.
- **`docs/design/` directory does not exist.** Creating it is fine; just ensure the directory is created alongside the file (a `mkdir -p` equivalent before writing the file).
- **Indoor/outdoor cap scope:** the per-slot placement in the daily slot struct is correct for the Step 3 test case ("60-minute indoor cap that does not constrain a longer allowed outdoor slot"). Confirm the design doc spells out that a `nil` indoor cap means "uncapped" and that outdoor and indoor caps are evaluated independently per slot, not as a combined day-level cap.
- **JSON tags:** since this is a pure internal package consumed by future tools rather than directly serialized to MCP responses, JSON tags on the planning structs are not required. The design doc should note this boundary explicitly so future tooling code doesn't add them redundantly.

---

## What the design doc must cover to unblock Step 2

The `docs/design/plan-filler-constraints.md` artifact in this step is load-bearing because Step 2 implements against it. At minimum it needs:

1. All struct field names with Go types and units
2. The severity of each `ViolationCode` (or `ResultCode`)
3. The ReconciliationResult both-dimension contract
4. The invariant: `RemainingTarget = FullWeekTarget − CompletedLoad − FixedLoad`, and that the validator uses `RemainingTarget`, never `FullWeekTarget`, for infeasibility checks during an in-progress week
5. The indoor/outdoor cap independence rule
6. A worked example showing a two-slot day where combining the slots would violate per-slot caps


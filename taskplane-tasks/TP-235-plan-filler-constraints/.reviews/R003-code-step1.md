# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking findings

### 1. The requested session count does not constrain the candidate batch

`ValidateCandidates` only compares `RequestedSessionCount` with structural capacity (lines 278–287); it never compares it with `len(candidates)`. Consequently, a request for one session with two compatible candidates on a day that permits two sessions returns two `Valid: true` results and no warning. This contradicts the documented contract that `RequestedSessionCount` is “the number of sessions the caller wants placed” and fails the mission’s over-scheduling protection.

Add an explicit deterministic result for a batch exceeding the requested count (and define under-fill semantics), then ensure a caller cannot treat an over-request batch as placeable.

### 2. Slots are counted as capacity but never consumed by validation

`availableSlotCount` treats each declared slot as one unit of capacity (lines 532–545), while the per-candidate path tracks only `MaxSessionsPerDay` (lines 248–274). For example, one 60-minute slot with `MaxSessionsPerDay: 2` accepts two 60-minute candidates as individually valid even though both require the same single independent window. With `RequestedSessionCount: 1`, finding 1 is also bypassed.

Define and implement deterministic per-slot allocation/consumption (including compatible-slot selection), or make the data model explicitly represent only caps rather than independent availability slots. As written, the validator can certify a schedule that cannot fit its available slots.

### 3. Slot violation codes are emitted when their documented condition is false

`checkSlotConstraints` records a code if *any* slot rejects for that reason (lines 456–509), but both code comments and the design doc say these codes mean the candidate is rejected by *every* slot. With a 45-minute Ride slot and a 120-minute Run-only slot, a 90-minute Ride has no complete match; the function returns both `slot_duration_exceeded` and `sport_not_allowed`, although it fits the duration cap of the second slot and Ride is allowed by the first.

Return a truthful deterministic explanation for the no-compatible-slot case (for example, per-slot rejection details or a distinct no-compatible-slot code), or change the aggregation and documented meanings so callers are not told to fix constraints that are not universally breached.

## Verification

- `go test ./internal/planning` passes, but reports `[no test files]`.
- `go vet ./internal/planning` passes.
- `git diff --check 6a30598..HEAD` fails on trailing whitespace and a final blank line in the committed `R001-plan-step1.md` review artifact.

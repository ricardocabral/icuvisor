# Code Review: Step 1 — Model interval-source heuristics

## Verdict

Approved. No blocking findings.

## Review scope

- Compared `1588cca4cf5385399cff6b7db8116d724f631cb8..HEAD`.
- Changed file: `taskplane-tasks/TP-090-auto-lap-disambiguation/STATUS.md`.
- Read the task prompt/status and checked the referenced interval DTO/response code for consistency.

## Findings

None.

## Notes

- This step is planning/modeling only; no implementation code was changed.
- The documented field inventory matches `internal/intervals/activity_details.go` and the current `get_activity_intervals` shaping path.
- The classifier contract now defines source precedence, minimum usable rows, distance/duration targets, tolerances, invalid-data handling, contiguity, unit assumptions, metadata semantics, shared helper placement, and acceptance examples.
- The conservative precedence for structured signals over near-uniform rows directly addresses the main false-positive risk for structured repeats.

## Follow-up for Step 2/4

When implementing, lock down the ambiguous parts with tests: objective edge-row/partial-row dropping, explicit-marker precedence, and that unavailable/Strava-blocked interval responses keep their existing shape unless evaluable rows are present.

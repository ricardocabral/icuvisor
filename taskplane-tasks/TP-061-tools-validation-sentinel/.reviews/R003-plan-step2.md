# R003 — Plan review: Step 2

**Verdict: APPROVE**

The updated Step 2 plan addresses the R002 gaps and is safe to execute for this small scoped change.

## What is good

- It explicitly creates `internal/tools/errors.go` with a package-level `ErrInvalidInput` sentinel before changing call sites.
- It names the five migration targets by function/branch rather than relying on stale line numbers:
  - `validateActivitiesTokenArgs` unsupported token version
  - `decodeDeleteEventsByDateRangeRequest` too-long date range
  - `applyTrainingPlan` no relative-day workout metadata
  - `validateIntRange` wellness integer scale bounds
  - `validateIntMin` wellness integer minimum bounds
- It records the `validateFloatMin` scope decision, avoiding an accidental wider sweep.
- It specifies the desired wrapping pattern (`fmt.Errorf("%w: existing message", ErrInvalidInput)`) and keeps existing `NewUserError(..., err)` wrapping so `errors.Is` works through `UserError.Unwrap()`.
- The proposed test matrix now targets the actual cited branches, including separate wellness tests for `validateIntRange` and `validateIntMin`.

## Execution notes

- In files under package `tools`, use `ErrInvalidInput` directly; reserve `tools.ErrInvalidInput` for external-package tests or callers.
- Keep non-cited `errors.New(...)` validation branches unchanged in this step unless a later review intentionally expands scope.
- For handler-level tests, assert both `errors.Is(err, ErrInvalidInput)` and, where practical, that the public `UserError` message remains the existing tool-level invalid-arguments message. This is not a new requirement for Step 2, but it will protect the Step 3 mapper work.

No further plan revisions are required before implementation.

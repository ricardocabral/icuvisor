# Plan Review: TP-008 Step 1 — typed `Unit` enum

## Verdict: approved

The revised Step 1 plan is concrete enough to implement. It addresses the key requirements from the task prompt and PRD §7.4 #17: a dedicated `internal/units` package, exact upstream enum values, a never-error `ParseUnit(string) (Unit, raw string)` contract, explicit unknown handling, WARN logging via `slog`, and table-driven tests in the same step.

## What looks good

- Keeps the intervals.icu enum isolated from `internal/response/units.go`, preserving the response-layer `UnitSystem` boundary.
- Covers all required unit families and calls out stable exported constants with exact upstream string values.
- Defines the parser semantics clearly enough to test: trim input, case-sensitive known matches, empty raw for known units, `UnitUnknown` plus raw token for unknowns.
- Explicitly notes that callers must preserve the second return value when `UnitUnknown` is returned, which is necessary because a plain enum constant cannot carry raw upstream data by itself.
- Includes Step 1 tests for every enum member plus unknown/mixed/future tokens and raw preservation.
- Logging plan follows project conventions (`slog.Default()` in libraries) and avoids logging response bodies.

## Minor implementation notes

- Ensure exported identifiers have doc notes so `revive` passes.
- Consider suppressing or explicitly testing empty-string warning behavior if it becomes noisy; empty input should still be deterministic.
- At the step boundary, run `gofmt`/`go test ./internal/units` at minimum, even though the full `make test`/`make lint`/`make build` gate is listed later.
- Do not forget the task-level documentation requirement to update `CHANGELOG.md` before TP-008 is completed.


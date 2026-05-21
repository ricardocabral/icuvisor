# R002 Code Review — Step 1: Define projection assumptions

**Verdict:** Request changes

## Findings

### 1. `recovery_week_cadence` schema permits a value the decoder rejects

- **Severity:** Medium
- **File:** `internal/tools/get_fitness_projection.go:168`

The JSON Schema advertises `recovery_week_cadence` as any integer from `0` through `12`, but `decodeFitnessProjectionRequest` rejects `1` (`0` or `2-12` only). MCP clients and LLMs rely on the schema to decide what to send, so this public contract currently says `1` is valid and then returns a user error.

Please make the schema match the decoder, e.g. `oneOf`/`anyOf` for `{const: 0}` or `{minimum: 2, maximum: 12}`, or change the decoder if cadence `1` is intentionally supported.

### 2. Horizon contract is internally inconsistent

- **Severity:** Medium
- **File:** `internal/tools/get_fitness_projection.go:14`, `internal/tools/get_fitness_projection.go:80`

The user-facing invalid-arguments message says callers must provide “exactly one `horizon_date` or `horizon_days`”, but the decoder and schema allow callers to provide neither and silently default to `42` days. Step 1 is intended to lock down the request contract before engine work starts, so this should be made unambiguous now.

Either require exactly one horizon field in both schema/decoder, or keep the default and update the error text/schema descriptions to say both fields are optional and that omitting both uses the default horizon.

### 3. Planned-load dates are validated after trimming but returned unnormalized

- **Severity:** Low
- **File:** `internal/tools/get_fitness_projection.go:137`

`validateProjectionPlannedLoads` trims `load.Date` for validation and duplicate detection, but `decodeFitnessProjectionRequest` returns the original slice with the untrimmed date values. If Step 2 builds the override map from `args.PlannedDailyLoads` directly, a date like `" 2026-06-02 "` passes validation but will not match the athlete-local projection date.

Please normalize planned load dates in the decoded request, or ensure the projection engine uses the same trimmed representation. Since other request date fields are normalized in the decoder, doing it here would keep the contract consistent.

## Task/status tracking note

`STATUS.md` records `Review R001 | plan Step 1: APPROVE`, but the checked-in `R001-plan-step1.md` verdict is “Needs changes / not yet reviewable”. Please correct the status/review table so the task history is accurate.

## Verification

Ran:

```sh
go test ./internal/analysis ./internal/tools
```

Result: passed.

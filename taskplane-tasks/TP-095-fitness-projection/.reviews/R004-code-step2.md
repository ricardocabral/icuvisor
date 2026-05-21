# R004 Code Review — Step 2: Implement projection engine

**Verdict:** Request changes

## Findings

### 1. Full curve omits explicit zero-load days

- **Severity:** Medium
- **File:** `internal/tools/get_fitness_projection.go:81`

`fitnessProjectionPoint.TrainingLoad` is tagged `json:"training_load,omitempty"`. The request schema explicitly allows `planned_daily_loads[].training_load: 0`, and `recovery_week_load_pct: 0` can also produce projected zero-load recovery days. In `include_full:true` responses those days lose the `training_load` field entirely, so clients cannot distinguish an explicit rest/zero-load projection from a missing load value. The projected curve is the only place load source and daily modeled load are exposed, so this breaks the output contract for valid inputs.

Please remove `omitempty` for projected points, or use a pointer/special start-point shape if day 0 should omit load while projected days preserve `0`.

### 2. Previously flagged request-schema contract issues are still present

- **Severity:** Medium
- **File:** `internal/tools/get_fitness_projection.go:19`, `internal/tools/get_fitness_projection.go:238`, `internal/tools/get_fitness_projection.go:329`

The Step 2 handler now exposes the projection tool behavior, but the Step 1 contract mismatches from R002 remain unresolved:

- The public invalid-arguments message says callers must provide “exactly one `horizon_date` or `horizon_days`”, while the decoder still accepts neither and defaults to 42 days.
- The JSON Schema advertises `recovery_week_cadence: 1` as valid (`minimum: 0`, `maximum: 12`), while the decoder rejects it (`0` or `2-12`).

Because MCP clients and LLMs rely on the schema/description to construct calls, this should be fixed before the engine is considered complete. Either make the decoder match the public schema/descriptions, or make the schema/descriptions match the decoder.

### 3. Status/review tracking is inaccurate

- **Severity:** Low
- **File:** `taskplane-tasks/TP-095-fitness-projection/STATUS.md:24`, `taskplane-tasks/TP-095-fitness-projection/STATUS.md:82`, `taskplane-tasks/TP-095-fitness-projection/STATUS.md:115`

`STATUS.md` marks Step 1 complete and records R001/R002/R003 as approved in the notes, but the checked-in review files say otherwise (`R001`: Needs changes, `R002`: Request changes, `R003`: Needs changes / not yet reviewable). The Reviews table is also empty. This makes the task history misleading for later reviewers and implementers.

Please update the status file to reflect the actual review verdicts and any remaining blockers/follow-ups.

## Verification

Ran:

```sh
go test ./internal/analysis ./internal/tools
```

Result: passed.

# Plan Review — TP-080 Step 1

Verdict: **changes requested**.

I only found the high-level Step 1 checklist in `STATUS.md`; there is not yet an implementable plan to review. Before starting the extraction, expand the plan so it is clear what will change, what will deliberately remain unchanged, and how the power-curve API will be protected.

## Required amendments

1. **Acknowledge the current baseline instead of assuming client support is missing.**
   `internal/intervals/fitness.go` already has `ListAthleteHRCurves`, `ListAthletePaceCurves`, `CurveParams.DurationSeconds`, and `CurveParams.DistanceMeters`. The Step 1 plan should say whether client work is limited to adding regression coverage, or identify a specific upstream/client gap that still needs code.

2. **Separate HR duration curves from pace distance curves.**
   Do not blindly genericize everything around `duration_seconds`. Existing `get_best_efforts` treats power/HR as duration curves using `secs`, but pace as a distance curve using `distance`/`distance_meters`. The extraction plan should name the two axes and preserve that distinction so Step 2 does not accidentally ship a pace tool with the wrong request/response contract.

3. **Define the exact reusable plumbing boundary.**
   The plan should specify which pieces will be shared, e.g. date-range validation, curve-spec construction, positive bucket normalization, bucket lookup, activity-id lookup, missing-bucket metadata, and shaped encoding. Keep metric-specific typed structs/field names where they affect the public API (`watts`, future HR bpm field, future pace unit field). Avoid replacing typed responses with broad `map[string]any` plumbing.

4. **Lock `get_power_curves` behavior before/while refactoring.**
   Add or strengthen regression tests that prove the existing tool still emits the same terse shape, omits `full` by default, includes upstream raw payload only with `include_full`, preserves missing-bucket metadata, and uses the same default sport/date/bucket handling. The plan should also state that the existing tool name, description/schema semantics, tier, and output field names are not changed in Step 1.

5. **Add interval-client endpoint tests for HR and pace.**
   `internal/intervals/fitness_test.go` currently covers the power curve endpoint but not HR/pace. Step 1 should include tests for `/hr-curves.json` and `/pace-curves.json`, including query construction (`curves`, `secs` for HR, `distances` for pace, and `type` when a sport is supplied). If sport omission is intentional for HR/pace because `requireType` is false, cover or document that behavior.

## Suggested Step 1 test command

At this step boundary, run at least:

```sh
go test ./internal/intervals ./internal/tools
```

Full `make test`, `make build`, and `make lint` can remain in the later verification steps, but targeted tests are needed here because this step is explicitly a refactor/regression-safety step.

With those amendments, the step will be appropriately scoped: extract only the reusable curve mechanics, document any upstream uncertainty, and leave public catalog/tool additions for Step 2.

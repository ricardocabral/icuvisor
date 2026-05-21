# Plan Review — TP-080 Step 3

Verdict: **APPROVE**

The Step 3 plan is aligned with the task requirements and the current implementation state. It explicitly targets the remaining coverage gap from Step 2: symmetry across power/HR/pace response shaping, terse-by-default vs `include_full`, missing-bucket metadata, axis-specific fields, pace unit conversion, unknown-unit fallback, and targeted verification.

## Notes

- The plan correctly treats pagination as not applicable for these upstream curve endpoints and substitutes missing-bucket metadata as the relevant terse-boundary behavior to test.
- Consolidating shared assertions in table-driven tests is a good fit here, as long as metric-specific point fields remain explicit: `watts`, `heart_rate_bpm`, and `elapsed_seconds` plus the preferred pace field.
- For the pace-specific table, include assertions on `_meta.units` and on the absence of the non-preferred pace field so metric/imperial behavior is unambiguous.
- When practical, have the fake client record the `CurveParams` it receives so the tests also lock in normalized bucket forwarding and the intentional HR/pace optional-sport behavior. This is not a blocker for the plan, but it would make the symmetry coverage more robust.

## Verification expectations for the step

Targeted runs should include at least:

```sh
go test ./internal/tools ./internal/intervals ./internal/toolcatalog ./internal/toolchecks ./internal/safety
go run ./scripts/check_schema_stability.go -baseline-dir internal/tools/schema_snapshot -require-baseline
```

Full `make test`, `make build`, and `make lint` remain appropriately deferred to the later verification/documentation steps unless Step 3 changes generated artifacts unexpectedly.

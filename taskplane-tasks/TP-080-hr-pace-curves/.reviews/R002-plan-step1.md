# Plan Review — TP-080 Step 1

Verdict: **APPROVED**.

The revised Step 1 plan in `STATUS.md` addresses the prior review concerns and is now scoped well for an extraction/regression-safety step rather than prematurely adding the public HR/pace tools.

## What looks good

- It correctly acknowledges the existing intervals client baseline: `ListAthleteHRCurves`, `ListAthletePaceCurves`, `DurationSeconds`, and `DistanceMeters` already exist, so Step 1 client work is limited to regression coverage unless implementation reveals a concrete gap.
- It preserves the important axis split: power/HR are duration-second curves (`secs` / `duration_seconds`), while pace is distance-meter based (`distances` / `distance_meters`).
- It defines a sensible reusable plumbing boundary: date validation, curve-spec construction, positive bucket normalization, lookup helpers, activity-id lookup, missing-bucket metadata, and shaped encoding.
- It explicitly keeps metric-specific typed response fields instead of collapsing public responses into generic maps.
- It locks `get_power_curves` public behavior during refactoring, including terse/full behavior, missing-bucket metadata, defaults, tool name/schema/tier, and output field names.
- It adds the missing HR/pace interval-client endpoint coverage called out in R001, including query construction and the intentional sport-omission behavior.

## Non-blocking notes for implementation

- At the Step 1 boundary, run at least:

  ```sh
  go test ./internal/intervals ./internal/tools
  ```

- Keep any newly shared helpers unexported and narrow unless Step 2 proves they need a broader interface.
- When refactoring `get_power_curves`, prefer tests that assert observable JSON fields over implementation details so the public contract remains the guardrail.

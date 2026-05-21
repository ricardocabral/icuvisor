# Plan Review — Step 4: Extract row helpers

Verdict: APPROVE

No blocking issues with the Step 4 plan. The current row-shaping block in `internal/tools/get_activities.go` is clearly above the 30 LOC threshold (`activityRow`, distance/pace/speed helpers, run-like detection, and scalar helpers), so extracting it to `internal/tools/get_activities_row.go` matches the task prompt and should keep the main file focused on handler/schema/glue.

Notes for implementation:

- Keep this behaviour-preserving: move code as-is first; avoid changing unit conversion, pace eligibility, rounding precision, Strava unavailable row handling, or `include_full` raw preservation.
- Be careful with helpers that are package-wide today (`stringValue`, `intValue`, `anyString`, and possibly `round`). Several other `internal/tools` files reference them, so if they move into `get_activities_row.go`, keep their names/signatures unchanged and package-private.
- Existing handler tests cover the important row-shaping contracts (`TestGetActivitiesDoesNotEmitPaceForCycling`, `TestGetActivitiesShapesStravaFullAndUnits`, timezone fallback, Strava rows). Keeping those is acceptable for this refactor; adding a small direct row-shaping test is optional, not required for the plan.
- After the move, run at least `go test ./internal/tools` (or the project `make test`) to catch any package-wide helper fallout and import cleanup issues.

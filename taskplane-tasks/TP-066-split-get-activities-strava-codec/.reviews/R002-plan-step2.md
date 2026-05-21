# Plan Review — Step 2: Extract Strava heuristic

Result: **Approved with one required adjustment**

The proposed Step 2 is appropriately scoped: moving `isStravaBlocked` into `internal/tools/get_activities_strava.go` and moving its table-driven golden test into `get_activities_strava_test.go` is a mechanical split that matches TP-066 and preserves the PRD contract for Strava-imported activity labelling.

## Required adjustment

- **Do not narrow the verification below the task's Step 2 requirement without saying so.** `PROMPT.md` says Step 2 should “Run all checks,” while `STATUS.md` currently says only “Run targeted tool tests for the Strava split.” Either update the Step 2 plan/status to run the requested broader check (`make test` or the repo's agreed equivalent), or explicitly defer full-suite verification to Step 5 and at least run `go test ./internal/tools` after the extraction.

## Implementation notes to keep the plan safe

- Keep the extraction behavior-preserving: copy the current decision order in `isStravaBlocked` exactly, including the `source`/`_note` marker handling, meaningful-field short-circuit, empty/raw-stub handling, and the non-stub fallback for additional fields such as power-only `N/A` HR/cadence cases.
- Keep `isStravaBlocked` package-private. No new exported API is needed.
- Be careful with “helper data”: avoid turning function-local maps/slices into mutable package-level state unless there is a clear reason. A straight move with local literals is lowest risk.
- Move only the pure heuristic golden test if the goal is separation. Existing handler/response integration tests that assert Strava rows are surfaced can remain in `get_activities_test.go` or move without behavior change, but they should not be rewritten as part of this step.
- After moving, clean up imports in both files (`get_activities.go` may no longer need `fmt` only if no other helpers still use it; currently it is still used elsewhere) and run `gofmt`.

No schema, output shape, or heuristic-rule changes are needed for this step.

# Plan Review — TP-085 Step 5

**Verdict:** APPROVE

The Step 5 plan covers the required verification gate for this task: targeted tests, full unit suite, build, lint, and explicit handling of any failures. This is sufficient for a small user-facing wording change as long as the command results are recorded before the step is marked complete.

## What is covered

- Targeted tests for the Strava/unavailable behavior changed in Steps 2–3.
- Full test suite via `make test`.
- Binary build via `make build`.
- Lint via `make lint`.
- Failure disposition: fix failures or document clearly unrelated pre-existing failures in `STATUS.md`.

## Implementation notes

- Do not rely only on the Step 4 “full quality gate” checkbox unless the exact command outputs are already recorded and still apply to the current tree. Prefer re-running the gates for Step 5 and logging the results.
- Use a targeted tools-package command that exercises the updated Strava/unavailable paths, for example:

  ```sh
  go test ./internal/tools -run 'Test(IsStravaBlocked|StravaBlockedWorkaround|GetActivities.*Strava|GetActivity(Streams|Splits|Intervals|Messages).*Unavailable|GetExtendedMetrics.*Strava)'
  ```

  The exact regex may vary, but it should cover the list-row, messages, intervals, streams/splits fallback, and extended-metrics assertions added for this task.
- Run the full gates from the repository root:

  ```sh
  make test
  make build
  make lint
  ```

- Update `STATUS.md` with the commands run and their pass/fail outcomes. If a failure is claimed as pre-existing/unrelated, include enough detail to make that claim reviewable.

No plan changes are required before executing Step 5.

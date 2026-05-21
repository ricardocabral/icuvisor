# Plan Review: TP-082 Step 4 — Verify full write cluster

## Verdict: APPROVE

The Step 4 plan covers the required verification gate for this task: rerun the write-tool cluster, run the full repository checks, and record the user-visible behavior change in `CHANGELOG.md`. Given Step 3 already landed the only implementation fix and R010 confirmed the focused tests pass, this is the right next step.

## Expectations for execution

- Use the existing targeted write-cluster command from STATUS.md so every audited write tool is covered, not only the custom-item regressions:
  ```sh
  go test ./internal/tools -run 'Test(AddOrUpdateEvent|LinkActivityToEvent|AddActivityMessage|UpdateWellness|UpdateSportSettings|ApplyTrainingPlan|CreateWorkout|UpdateWorkout|CreateCustomItem|UpdateCustomItem)'
  ```
- Then run the required full gates exactly as planned:
  ```sh
  make test
  make build
  make lint
  ```
- Update `CHANGELOG.md` under `[Unreleased]` with a concise `Changed` entry for write-tool echo responses stripping upstream null keys by default while retaining full/raw detail where explicitly supported.
- Record command outcomes in `STATUS.md`. If any gate fails, fix in scope; only document a failure as pre-existing/unrelated if there is clear evidence.

## Non-blocking note

Because Step 3 adjusted custom-item write output schema descriptions, remember in Step 6 to perform the “Check If Affected” docs/generated-tool-reference review. That does not block Step 4, but it should not be forgotten before delivery.

# Plan Review: TP-082 Step 5 — Testing & Verification

## Verdict: APPROVE

The Step 5 plan is appropriate for a verification-only gate. It matches the task completion criteria: rerun the affected write-tool tests, run the full suite, build, lint, and either fix failures or document clearly evidenced pre-existing/unrelated failures in `STATUS.md`.

## Expectations for execution

- Do not rely solely on the Step 4 results; Step 5 should record its own fresh command outcomes in `STATUS.md`.
- Run the targeted write-tool cluster from the task audit so every affected write path remains covered:
  ```sh
  go test ./internal/tools -run 'Test(AddOrUpdateEvent|LinkActivityToEvent|AddActivityMessage|UpdateWellness|UpdateSportSettings|ApplyTrainingPlan|CreateWorkout|UpdateWorkout|CreateCustomItem|UpdateCustomItem)'
  ```
  Using `-count=1` is acceptable if you want to avoid cached output for this final verification pass.
- Run the required repository gates:
  ```sh
  make test
  make build
  make lint
  ```
- If any gate fails, fix in-scope regressions before advancing. Only mark a failure as pre-existing/unrelated when the evidence is explicit and recorded in `STATUS.md` with enough detail for the next reviewer.
- Avoid new behavior changes in this step unless a verification failure requires a focused fix; if a fix is needed, rerun the impacted targeted command plus the full gates.

## Non-blocking note

Step 6 still needs the documentation/delivery review, especially checking whether the custom-item write schema description changes require generated reference/docs updates. That does not block this testing plan.

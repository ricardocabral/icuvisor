# Plan Review — Step 1: Audit write/delete guidance

**Verdict: changes requested before executing Step 1.**

## Findings

1. **Missing create-workout audit target.** The step says to inspect create/update/delete workout and event guidance, but the Step 1 artifacts omit `internal/tools/create_workout.go` (and its tests). Because the risk is specifically “delete and recreate instead of edit in place,” the create side of the unsafe path must be audited explicitly. Add `internal/tools/create_workout.go` / `internal/tools/create_workout_test.go` to the Step 1 scope and record whether `create_workout` should warn against modifying existing templates.

2. **Safety-contract audit should include registration-time gating tests.** `go test ./internal/tools` is useful, but the existing catalog/safe-mode contract is also covered in `internal/safety/adversarial_test.go` (delete-mode registration matrix and no `confirm` schemas). If Step 1 is documenting the current safety contract, the plan should inspect that file and either run `go test ./internal/safety` or explicitly defer it with rationale.

## Suggested Step 1 plan adjustment

- Audit `add_or_update_event`, `create_workout`, `update_workout`, `delete_event`, and `delete_workout` descriptions, input schemas, and relevant tests.
- Include existing adversarial/safety coverage from `internal/safety/adversarial_test.go` and `docs/safety/adversarial-prompts.md` in the discovery notes.
- Run `go test ./internal/tools` plus `go test ./internal/safety` if safety-contract findings are changed or relied upon.

# Plan Review — Step 5

Verdict: **REVISE**

`STATUS.md` marks Step 5 as in progress, but it does not yet contain an implementation plan for the verification step beyond the checklist. For a final verification step, the plan should be explicit enough that failures are reproducible and the Step 5 checkboxes can be updated consistently.

Required plan additions:

1. Name the targeted regression command(s) to run before the full suite. At minimum, rerun the affected packages from this task: `go test ./internal/workoutdoc ./internal/units ./internal/response ./internal/tools`.
2. Specify the full verification sequence and recording expectations: `make test`, `make build`, then `make lint` (or another deliberate order), with each command/result added to the execution log.
3. Define how failures will be handled: fix task-related failures before marking the step complete; if a failure is believed pre-existing/unrelated, document the exact command, failure excerpt, and rationale in `STATUS.md` while leaving the corresponding checkbox accurate.
4. Clarify that no network-dependent tests or external services are expected for this step, consistent with the task prompt and repository rules.

Once those details are added to the Step 5 notes/status, the plan should be ready to execute.

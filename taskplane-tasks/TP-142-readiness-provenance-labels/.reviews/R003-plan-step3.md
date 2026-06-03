# Plan Review — Step 3: Testing & Verification

**Verdict:** Approve.

The Step 3 plan matches the task’s quality gate: run the full suite with `make test`, run lint with `make lint`, address any failures or document exact unrelated failure output, and verify `make build` passes. This is sufficient for a Level 1 plan-only review.

## Execution guidance

- Do not substitute targeted package tests for `make test`; Step 3 is the full-repo gate.
- If any command fails, fix task-related failures and rerun the same command to green. Only document a failure as pre-existing/unrelated when the exact command output is captured in `STATUS.md` and the cause is clearly outside this task.
- Record the final successful commands in `STATUS.md` before moving to Step 4.

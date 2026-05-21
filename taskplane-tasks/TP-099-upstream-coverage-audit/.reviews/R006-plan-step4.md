# Review R006 — Plan Review for Step 4

**Verdict:** REVISE

The Step 4 plan is too thin for the current state of the task. Step 2 added a Go audit script and Step 3 added a durable upstream-gap document, so verification needs to name the exact commands and the changelog decision rather than only restating the checklist.

## Required plan changes

- Add an explicit reproducibility check for the audit output:
  - `go run scripts/audit_zone_time_coverage.go`
  - Compare the summary against `STATUS.md` and `docs/upstream-gaps/zone-time-coverage.md` (`0` precomputed, `0` fallback, `6` unknown for each metric family; 36 skipped).
- Because code/scripts changed, include the quality gate required by the prompt, or explicitly explain how Step 4 and Step 5 will avoid duplicating responsibility. The safest Step 4 plan is to run:
  - `make test`
  - `make build`
  - `make lint`
- Include how failures will be handled: fix task-caused failures, or record clearly unrelated/pre-existing failures with command output in `STATUS.md`.
- Make the CHANGELOG decision concrete. This task added a user-facing upstream-gap document, and the task prompt says `CHANGELOG.md` must record user-visible docs behavior changes under `[Unreleased]`. The revised plan should either update `CHANGELOG.md` or document a specific rationale for not doing so.
- Record the verification commands and results in `STATUS.md` before moving to the next step.

No additional analyzer behavior changes are needed for this step.

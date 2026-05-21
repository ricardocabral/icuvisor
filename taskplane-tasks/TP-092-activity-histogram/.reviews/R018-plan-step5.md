# R018 plan review — Step 5: Testing & Verification

Verdict: REVISE

## Findings

1. **[P1] Step 5 starts from an inconsistent/unresolved Step 4 review state.**  
   `STATUS.md` marks Step 4 complete and lists R017 as approved, but `.reviews/R017-code-step4.md` has `Verdict: REVISE` with two P2 coverage gaps: strict input/schema coverage and stronger fixed-width raw-edge metadata coverage. A Step 5 plan that only runs the gates can produce green commands while still leaving the approved Step 4 test plan unimplemented. Before proceeding, reconcile this state explicitly: either resolve the R017 findings and get/record a follow-up approval, or correct the stale review file/status if it was superseded elsewhere.

2. **[P2] The targeted-test portion is too underspecified for traceability.**  
   The Step 5 checklist says “Targeted tests passing” but does not name the command(s). Given the approved Step 4 plan and the generated catalog/docs surface, record the exact targeted command set in `STATUS.md` before execution, e.g. `go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./cmd/gendocs` after the R017 fixes. If generated tool docs/catalog artifacts are part of the current diff, also include the canonical generation/check step (`make docs-tools` or the repo-equivalent command plus diff inspection) so stale generated artifacts are not hidden until a later CI job.

## What is otherwise acceptable

- The full-gate checklist itself is the right set for this step: targeted tests, `make test`, `make build`, `make lint`, and fixing or documenting any unrelated pre-existing failures.
- Deferring the full suite/build/lint from Step 4 to Step 5 is consistent with the approved Step 4 plan, as long as Step 4’s unresolved review findings are handled first.

## Suggested revised Step 5 plan

1. Resolve/reconcile R017 and update the review table/status before quality gates.
2. Run and log the targeted package tests for the affected surfaces.
3. Run `make test`.
4. Run `make build`.
5. Run `make lint`.
6. For every failure, either fix it in this task or document evidence that it is pre-existing and unrelated in `STATUS.md`.

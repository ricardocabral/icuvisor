# R001 plan review — Step 1

Verdict: **REVISE**

The Step 1 checklist is directionally right, but the checkpoint asks to confirm the exact one-write/readback wording and no proposed wording is recorded yet. Please add the actual prompt text to `STATUS.md` (or the plan) before updating the golden.

Required revisions:

- Resolve the TP-109 dependency mismatch. `TP-109` is still not started, so do not document a concrete new `_meta.warnings` contract from that task. Either wait for TP-109, or word this against current behavior: `validate_workout` diagnostics, write `_meta` fields that exist today (for example `workout_doc_warning` when present), and readback `workout_doc_summary`.
- Scope the parallel-write warning narrowly: bulk workout/calendar writes where schema wording, warning metadata, or description/`workout_doc` preservation semantics are uncertain. Do not imply a universal ban on parallel writes.
- If adding a custom `Guardrails` slice to `WeeklyPlanningPrompt`, include the two default guardrails too; `renderSpec` replaces defaults when `Guardrails` is non-empty.
- Keep the prompt terse enough for `TestPromptResourceCitationsStayTerse`'s line-count/verbosity guard.

Suggested wording to adapt:

> Before bulk calendar/workout writes, validate or preview one representative structured payload (use `validate_workout` for `workout_doc`/DSL when uncertain), perform one representative write, read it back, and inspect validation warnings, write `_meta` warning fields when present, and `workout_doc_summary`/stored description to confirm structured steps were preserved before writing the rest. Avoid parallel bulk writes while schema wording, warning metadata, or preservation semantics are ambiguous.

After revising the wording, updating `internal/prompts/testdata/weekly_planning.md` and running `go test ./internal/prompts` is the right targeted verification.

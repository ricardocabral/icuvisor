# R001 — Plan review: Step 1

**Verdict: REVISE**

The current Step 1 plan/status does not yet contain the required decision; `STATUS.md` still says “Pending decision.” This step is specifically to read TP-048, decide the sentinel home, and record that decision before wiring any sites, so the plan is not actionable as written.

## Required plan revision

Record a concrete decision in `STATUS.md`, preferably:

- Add `var ErrInvalidInput = errors.New("invalid input")` in `internal/tools/errors.go` with an exported doc note.
- Do **not** add `IsInvalidInput()` to `UserError` for this task.
- Keep `UserError` as the public-message wrapper; handlers can continue returning `NewUserError(invalid...Message, err)`, and `errors.Is` will still see `ErrInvalidInput` through `UserError.Unwrap()` once the underlying validation error wraps the sentinel.

## Rationale

TP-048 is complete and did not introduce validation categorization. The existing `UserError` in `internal/tools/registry.go` is a generic LLM-facing error wrapper used for invalid arguments, fetch failures, missing clients, write/delete failures, etc. Treating `UserError` itself as “invalid input” would conflate unrelated user-visible failures, or require adding a new tag/state to the type. A single exported sentinel is a cleaner stable contract and matches the task acceptance criteria (`errors.Is(err, tools.ErrInvalidInput)`).

The sentinel also works for lower-level validators such as `validateActivitiesTokenArgs`, `decodeDeleteEventsByDateRangeRequest`, `decodeApplyTrainingPlanRequest`, and the wellness range validators: they can return `fmt.Errorf("%w: ...", ErrInvalidInput)` directly, while the outer handler preserves the existing short public message via `NewUserError`.

## Guardrails to include in the plan

- Scope the new sentinel to `internal/tools`; do not introduce sentinels in other packages.
- Preserve current public error-message wording and wire shape; do not surface the full wrapped chain to the LLM.
- Use `%w` so tests and downstream mapping can use `errors.Is`.
- If Step 3 updates `publicToolErrorMessage`, keep `PublicErrorMessage` precedence or otherwise ensure invalid-argument `UserError` messages remain short and unchanged.

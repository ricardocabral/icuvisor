# Plan Review: Step 1 — Add `weekly_review` prompt registration

Verdict: **Approved for implementation**

I reviewed `PROMPT.md`, `STATUS.md`, the prior plan reviews, and the existing prompt registry/test patterns. The current Step 1 plan now covers the previously blocking items:

- registration work includes `catalog.go`, `registry.go`, and the non-golden `catalog_test.go` expectations;
- targeted verification is explicitly `go test ./internal/prompts`;
- the planned prompt text includes athlete-local date/timezone guidance, planned-vs-completed review, no write/delete without explicit approval, and advanced-capability fallback guidance;
- wellness is now included in the tool sequence, with `_meta.stale`/missing/provenance cautions so the completion criteria are covered in Step 1 rather than deferred to golden tests.

Implementation notes, not blockers:

- Use exact tool names in the rendered prompt/tool list where possible, especially `compute_load_balance` rather than loose “load-balance equivalent” wording.
- Keep `include_next_week` consistent with the current prompt argument model: arguments are string-valued, so describe accepted boolean text clearly if it is included.
- Watch `TestPromptResourceCitationsStayTerse`; the weekly-review guidance is necessarily dense, but it should still fit the existing terse prompt style.

The plan is sufficiently specific to proceed.

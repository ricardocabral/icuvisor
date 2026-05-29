# Code Review — Step 2

Verdict: **REVISE**

## Findings

1. **P2 — Weekly planning still does not gather available write tools before proposal.**  
   `internal/prompts/catalog.go:91-93` only calls `icuvisor_list_advanced_capabilities` when read helpers are unavailable, then drafts the proposal. The task requires season/week planning to gather available write tools before proposing plans. Please add explicit guidance to check visible/write capabilities (and surface unavailable gated tools) before the proposal, and lock it with a test assertion.

2. **P2 — Existing bulk-write and `workout_doc` preservation guardrails were dropped.**  
   `internal/prompts/catalog.go:93` replaces the prior detailed approved-write guidance with a much weaker sentence. For season planning, an approved set of calendar edits can still be a bulk write; losing the representative validate/write/read-back step and the reminder that `description` can replace structured DSL makes it easier for an assistant to corrupt or overwrite structured workouts. Please restore this guidance in a compressed form that still satisfies the terse prompt test.

## Verification

- Ran `go test ./internal/prompts -count=1` — pass.

# Review R003 — Plan review for Step 2: Shared prompt set

Verdict: Request changes

## Findings

1. **The prompt set includes an out-of-scope destructive/catalog prompt for a KR5 efficiency benchmark.** `KR5-10` asks to delete `[EVENT_ID]` if deletion is available and otherwise reason about catalog absence/confirmation workarounds. Step 2 is supposed to pin the “10 most common forum prompts” used to measure response-byte efficiency, not safety-gate behavior from TP-029. This prompt is MCP/catalog-aware rather than athlete-question-shaped, can bias median response bytes with short refusal/error paths, and is unsafe for later live/reference runs unless every mapping is carefully sandboxed. Replace it with a non-destructive common forum shape from the dogfood/read corpus, or move destructive refusal coverage to a separate safety benchmark that is not part of KR5 response-byte median.

2. **The shared set does not yet demonstrate the forum-shape provenance it claims.** The JSON has useful `forum_shape` labels and references the v0.2/v0.3 dogfood files, but it does not record which dogfood prompts or PRD/forum issues each KR5 scenario came from. Add a lightweight `source_prompt_ids`/`prd_anchor` field or a short note in `STATUS.md` so future reviewers can see that the set was fixed from representative prompts before measurements were inspected, rather than curated after the fact.

## Recommendation

Keep the current vendor-neutral structure (`version`, redaction conventions, `required_intents`, prompt text). Revise `KR5-10` into a non-destructive, common athlete/coach analysis prompt and add source/provenance metadata for the ten scenarios before proceeding with measurement.

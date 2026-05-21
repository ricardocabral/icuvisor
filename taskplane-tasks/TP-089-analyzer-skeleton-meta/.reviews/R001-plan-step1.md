# Plan Review R001 — Step 1: Design shared meta structs

Verdict: **Changes requested**

I could not find a step-specific design plan beyond the generic Step 1 checkboxes in `STATUS.md`. For this foundational contract, the implementation should not proceed until the plan records the concrete struct/helper shape and ownership decisions below.

## Required plan clarifications

1. **Define the exact analyzer `_meta` contract as typed, non-omitempty fields.**
   The plan should name the struct and fields and state that the mandatory fields are always emitted, including zero/false values:
   - `method string`
   - `source_tools []string`
   - `n int`
   - `missing_days int`
   - `missing_action string` defaulting to `"skip"`
   - `insufficient_sample bool`
   - optional `formula_ref string,omitempty` for formulas backed by `icuvisor://analysis-formulas`

   Do not use `omitempty` on mandatory fields. Also require `source_tools` to be initialized to a non-nil slice; this repo's response JSON encoder turns nil slices into `null`, which the response shaper can then strip in terse mode, violating the mandatory contract.

2. **State package ownership and response-shaping interaction.**
   Prefer a small `internal/analysis/meta.go` for analysis-domain types/helpers, with any tool-specific wrapper in `internal/tools` only if needed. The plan should explicitly account for `response.Shape`: analyzer keys inside `_meta` are preserved while common response-owned keys (`server_version`, `catalog_hash`, `delete_mode`, `toolset`, `units`) are added at the response boundary. Avoid adding `ServerVersion` to the analyzer meta struct itself unless there is a specific reason.

3. **Specify `n`, missing-day, and insufficient-sample semantics.**
   The plan should make these definitions unambiguous for downstream tools:
   - `n` is the number of usable samples after skipped missing values; for pairwise tools, it is usable pairs.
   - `missing_days` is a count of athlete-local daily buckets skipped; non-daily analyzers still emit `0`.
   - `missing_action` defaults to `skip`; there should be no forward-fill default.
   - Provide a helper such as `InsufficientSample(n, minN int) bool` / `SampleStatus(n, minN int)` and constants or call-site guidance for PRD minimums (`7` baseline, `14` correlation; stricter tool rules allowed).

4. **Avoid duplicating formula-ref strings.**
   The plan should say whether analyzer code imports the exported constants from `internal/resources/analysis_formulas.go` or centralizes them elsewhere. Do not hand-copy URI fragments in multiple packages; definition drift is explicitly called out as a roadmap risk.

5. **Determinism for `source_tools`.**
   The plan should decide whether source tools are preserved in execution order or sorted/deduped. Either is acceptable if documented, but golden files will be much less noisy if the helper normalizes deterministically.

## Suggested acceptance criteria for Step 1

Before marking Step 1 complete, update `STATUS.md` Notes/Discoveries with the selected contract, including mandatory JSON tags and helper names. That gives TP-090+ a stable target and gives Step 2 implementation something reviewable.


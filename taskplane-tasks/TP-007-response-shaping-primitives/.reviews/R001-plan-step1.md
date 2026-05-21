# Plan Review — TP-007 Step 1

Verdict: **Not approved yet**. The current `STATUS.md` does not contain an implementation plan for Step 1; it only repeats the prompt checklist and leaves `Step 1 decisions` as `Pending`. Since this step's deliverable is explicitly to record response-shaping decisions and tradeoffs before coding, there is nothing concrete enough to approve.

## Blocking findings

1. **No design decisions are recorded.**
   - `STATUS.md` still says `Pending` under `Step 1 decisions`.
   - The plan must specify the actual pipeline, ownership boundaries, and tradeoffs before Step 2 implementation starts.

2. **The plan must address `omitempty` before committing to `typed struct → marshal-to-map → strip nulls`.**
   - Existing response structs use `omitempty` heavily in `internal/tools/get_athlete_profile.go`.
   - If a struct field has `omitempty`, `json.Marshal` removes `nil`, `0`, `""`, and `false` before the null-stripper ever sees them. That prevents accurate `_meta.missing_fields` for nulls and can violate the “do not strip `0`, `""`, or `false`” invariant if future shaped structs use `omitempty` on meaningful fields.
   - Step 1 should decide one of: avoid `omitempty` on shaped nullable/value fields, use pointer fields for optional values, build the map with reflection that can distinguish zero values from JSON null, or maintain expected-field metadata separately.

3. **There is no chosen field-renaming strategy.**
   - The prompt asks to decide where canonical renames live and prefers per-tool struct tags.
   - The plan should explicitly choose struct tags or a central rename map, and explain how dynamic unit-dependent keys like `distance_km` / `distance_mi` and pace fields are handled without a brittle global rename pass.

4. **The single response chokepoint is unspecified.**
   - Existing handlers currently build a typed response, `json.Marshal` it for text content, and also set `StructuredContent` directly.
   - To guarantee `_meta.server_version` on every response, the plan must identify the chokepoint/API that every read tool will call, and clarify whether it returns `map[string]any`, typed structured content, JSON text, or both.

## Required additions before approval

Please update `STATUS.md` with a concrete Step 1 design covering at least:

- Pipeline order and exact API shape for `internal/response`.
- How null stripping collects `_meta.fields_present` and `_meta.missing_fields`, including nested objects/arrays and deterministic ordering.
- How `include_full: true` interacts with null stripping, `_meta.server_version`, `_meta.units`, scales, and debug metadata.
- How `_meta` is represented for a single object vs a list of top-level rows, and how `_meta` key collisions are handled.
- The field-renaming decision, preferably per-tool JSON tags, plus the approach for unit-dependent field keys.
- How the design avoids `omitempty` hiding nulls or meaningful zero/empty/false values before the shaping pipeline.
- Confirmation that existing `config.NormalizeAthleteID` remains the central athlete-ID normalizer unless a specific gap is found.
- Where timezone rendering and unit-system helpers live, and whether Step 1 is only documenting conventions or also defining interfaces for later steps.

Once those decisions are written down, this step should be reviewable.

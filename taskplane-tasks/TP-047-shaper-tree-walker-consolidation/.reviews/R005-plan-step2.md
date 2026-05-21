# Plan Review — Step 2: Pick the approach

**Decision: Changes requested before implementation proceeds.**

`STATUS.md` still has Step 2 as placeholders only: no approach has been selected, the Decisions section is empty, and there is no typed-struct enumeration or visitor/predicate sketch. That is not enough planning for the highest-risk design choice in this task.

## Blocking concerns

1. **No approach decision is recorded.**
   Step 2 requires choosing typed shaping vs. a single visitor-based tree walker and justifying the choice by diff size, blast radius, and `include_full` fit. The current status only repeats the checklist. Please update `STATUS.md` with an explicit decision and rationale before Step 3 starts.

2. **The planned mechanism for eliminating the marshal round-trip is unspecified.**
   `Shape` currently relies on `json.Marshal`/`json.Unmarshal` to apply JSON tags, `omitempty`, number normalization, and map/array conversion before the shaper walks `map[string]any`. The Step 2 plan must say what replaces that on the happy path and which current call paths it covers. A fast path that only handles already-generic `map[string]any` values would not satisfy the task for representative tool responses like `get_activities` and `get_fitness`, which pass typed DTOs.

3. **If choosing typed shaping, the plan must account for package boundaries and contract preservation.**
   `internal/tools` imports `internal/response`, so `internal/response` cannot import tool-specific response types without creating cycles. If typed shaping is chosen, enumerate the envelope/row structs or interfaces to introduce, where they live, which call sites must change, and exactly where `json.RawMessage` sits for `include_full` passthrough. Also explain how JSON tag/`omitempty` behavior remains byte-identical to the Step 1 goldens.

4. **If choosing the fallback single walker, the visitor shape and predicate set need to be concrete.**
   The plan should sketch the recursive helper signature and the predicates/actions that replace the existing walkers, including the special provenance behavior:
   - ordinary `fetched_at` / `query_type` debug fields are dropped when debug metadata is disabled,
   - `_meta.provenance.<field>.fetched_at` is preserved and not filtered from missing fields,
   - null stripping still records the same dotted/indexed paths, and
   - scales still ignore nested `_meta` content.

## What a sufficient Step 2 update should contain

Add a short Decisions entry in `STATUS.md` with one of these shapes:

- **Typed shaping:** list the new envelope abstractions/structs, their package location, changed call sites, `json.RawMessage` passthrough fields, and any narrow fallback case that still uses marshal/unmarshal.
- **Single visitor walker:** give the visitor function signature, the path representation, the predicate/action names, and the fast-path strategy for avoiding the marshal round-trip for typed DTOs or an explicit rationale for any remaining fallback.

Until that is recorded, the plan does not give Step 3 enough guardrails to preserve byte-identical output while removing the marshal round-trip.

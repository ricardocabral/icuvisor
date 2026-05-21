# Plan Review — TP-007 Step 2

Verdict: **Not approved yet**. The current Step 2 "plan" in `STATUS.md` only restates the prompt checklist. It does not define the implementation approach for the null-stripper, and Step 1 still has unresolved design blockers that directly affect Step 2 correctness.

## Blocking findings

1. **There is no concrete Step 2 implementation plan.**
   - `STATUS.md` marks Step 2 in progress, but the section only contains the four prompt bullets.
   - A reviewable plan needs to specify the API in `internal/response`, the data types it accepts/returns, how callers identify top-level rows, and how the shaped value is used for both MCP text and structured content.

2. **Step 1 blockers still block Step 2.**
   - The recorded pipeline still relies on JSON marshal/unmarshal semantics where `omitempty` remains authoritative.
   - That makes Step 2 unable to distinguish fields that were JSON `null` from fields removed before shaping, and it prevents `include_full: true` from surfacing raw nulls for `omitempty` pointer fields.
   - Before implementing null stripping, the plan must decide how shaped response structs avoid `omitempty` on fields whose nulls must be reported, or introduce another explicit expected-field/null-capture mechanism.

3. **Per-row metadata semantics are undefined.**
   - The task requires `_meta.fields_present` and `_meta.missing_fields` per top-level row, emitted only when at least one strip happened.
   - The plan does not define what counts as a row for a single object, a bare array, or a wrapper object containing row arrays, nor whether the helper mutates each row's `_meta` or returns a sidecar result.
   - It also does not define whether `fields_present` includes `_meta`, whether nested keys are represented as local names or paths, or how duplicate missing fields from nested arrays are deduplicated.

4. **Determinism and collision handling are unspecified.**
   - Tests will need stable ordering for `fields_present` and `missing_fields`; the plan should require sorted output.
   - The plan should define behavior when an input row already has `_meta`: merge into it, reserve/overwrite only the null-strip keys, and reject or preserve existing non-conflicting metadata.

5. **Recursive behavior needs sharper rules.**
   - The plan says nested objects and arrays-of-objects, but does not cover arrays containing scalars, `null` elements, mixed values, empty objects after stripping, or maps nested inside arrays.
   - It should explicitly state that only map/object keys with value `nil` are removed; `0`, `""`, `false`, empty arrays, empty objects, and non-object array elements are preserved unless a different behavior is intentionally chosen.

6. **`include_full: true` behavior is incomplete.**
   - The plan must state that `include_full` skips null stripping and therefore does not emit null-strip `_meta.fields_present`/`_meta.missing_fields`.
   - It should also clarify that later common metadata (`server_version`, units, debug metadata, scales) still goes through the common response chokepoint even when null stripping is disabled.

## Required additions before approval

Please update `STATUS.md` with a real Step 2 plan covering at least:

- Proposed `internal/response` function names and signatures for shaping a single row and multi-row response.
- The representation of top-level rows and how wrappers identify row collections.
- Exact recursive null-strip rules for maps, arrays, scalar arrays, `null` array elements, empty containers, and mixed arrays.
- Exact `_meta.fields_present` / `_meta.missing_fields` semantics, including nested path format, deduplication, deterministic sorting, and whether `_meta` itself is excluded from `fields_present`.
- How existing row `_meta` is merged without clobbering future Step 3/4/5 metadata.
- How the implementation avoids the `omitempty`/raw-null problem identified in Step 1.
- Tests planned in this step, including preservation of `0`, `""`, and `false`, nested strips, arrays-of-objects, multi-row independence, pre-existing `_meta`, and `include_full` bypass.

Once those decisions are recorded, Step 2 should be straightforward to implement and review.

# Plan Review — Step 2: Pick the approach

**Decision: Approved.**

The updated `STATUS.md` now makes a concrete Step 2 decision: use the fallback single visitor walker, and replace the current marshal/unmarshal conversion with an `internal/response` reflection-based JSON-value builder for normal typed DTOs. It also gives the requested rationale for not pursuing full typed shaping: importing tool DTOs would create package cycles, while mirroring the tool/intervals response types would expand the diff beyond the intended M-sized refactor.

The plan now covers the items requested in R005:

- It names the chosen approach and justifies it by diff size, blast radius, package boundaries, and `include_full` handling.
- It sketches the visitor shape and predicate/action set (`stripNullVisitor`, `dropDebugVisitor`, scale collection, and explicit debug/provenance predicates).
- It states how the marshal round-trip will be removed on the happy path while allowing a narrow documented fallback for custom/unsupported reflection cases.
- It calls out the provenance/debug edge cases that must remain byte-identical, including preserving `_meta.provenance.<field>.fetched_at` while dropping ordinary debug metadata.

## Implementation cautions for Step 3

These are not blockers to the Step 2 plan, but they should be treated as guardrails during implementation:

1. **Preserve encoding/json semantics where Shape currently relies on them.** The reflection builder needs targeted tests for JSON tag names, `json:"-"`, `omitempty`, pointers/interfaces, nested structs, maps/slices, and any custom `json.Marshaler`/`json.RawMessage` fallback behavior that remains. If a case falls back to marshal/unmarshal, document the exact type/class in `STATUS.md` as planned.

2. **Do not introduce caller-visible mutation.** The current marshal/unmarshal path deep-copies the payload before `dropDebugMetadata`, `stripNulls`, and metadata additions mutate shaped maps. The replacement should build fresh `map[string]any` / `[]any` values, including for `include_full` maps/slices, or otherwise prove that the shaper cannot mutate caller-owned data.

3. **Keep the fallback narrow and measurable.** The normal typed tool responses covered by the Step 1 goldens (`get_activities`, `get_fitness`, wrapper rows, provenance metadata) should not use the marshal/unmarshal fallback. If they do, the acceptance criterion for eliminating the happy-path round-trip is not met.

4. **Use the goldens as the contract, not just existing unit assertions.** After the visitor/reflection changes, re-run the Step 1 fixtures and require an empty diff before moving on to cleanup or lint/build verification.

No additional Step 2 plan changes are required before implementation proceeds.

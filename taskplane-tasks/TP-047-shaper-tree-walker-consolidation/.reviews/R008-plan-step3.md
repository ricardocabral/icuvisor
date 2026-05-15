# Plan Review — Step 3: Implement

**Decision: Changes requested before implementation proceeds.**

The Step 2 decision gives a reasonable overall direction (single visitor walker plus a reflection-based JSON-value builder), but the Step 3 implementation plan still leaves a few contract-critical details ambiguous. These are the places most likely to break the “byte-identical output” requirement or introduce mutation that the current marshal/unmarshal round-trip accidentally prevents.

## Blocking concerns

1. **Define the `toJSONValue` copy/ownership contract explicitly.**
   The current marshal/unmarshal path deep-copies the entire payload before the shaper mutates maps to strip nulls/debug fields and add `_meta`. The plan currently says maps/slices “recurse directly” and that `include_full` maps remain ordinary map/slice values, which is ambiguous: if the new fast path reuses caller-owned maps or slices, Step 3 can mutate tool payloads or upstream raw/full payloads.

   Update the Step 3 plan to require that `toJSONValue` always returns a fresh JSON-shaped tree for maps, slices, arrays, and struct fields, including nested `Full`/`include_full` payloads. Add a targeted no-mutation test or explicitly include it in the implementation checklist.

2. **Close the gap with `encoding/json` semantics before replacing the round-trip.**
   `Shape` currently relies on `encoding/json` for more than field tags: `omitempty`, anonymous/embedded field resolution, `json.Marshaler` / `encoding.TextMarshaler`, `json.RawMessage`, unsupported values, and number behavior after unmarshalling into `any`. The status says “primitives remain primitives,” which is not the same as the current marshal/unmarshal conversion and could change internal shaped values or, for edge numeric cases, output bytes.

   The Step 3 plan should state the supported fast-path scope and fallback order concretely. At minimum:
   - custom `json.Marshaler` / `encoding.TextMarshaler` values either use a narrow marshal fallback or are proven absent from normal shaped DTOs;
   - `json.RawMessage` is not treated as a raw `[]byte`/array of numbers;
   - unsupported JSON values fail with wrapped errors consistent with the current `json.Marshal` failure path;
   - numeric handling is either intentionally byte-equivalent to the old path or any internal type differences are documented and covered by tests.

3. **Add Step 3-specific tests for the new converter/walker, not only the five goldens.**
   The goldens are necessary but do not cover enough reflection edge cases. Before implementation proceeds, record that Step 3 will add focused table-driven tests for JSON tags/renames, `json:"-"`, `omitempty` for pointers/slices/maps/scalars, embedded fields or the chosen fallback for them, map/slice deep-copy behavior, custom marshaler/raw-message fallback behavior if retained, and the provenance/debug edge cases handled by the single walker.

4. **Make fallback accounting auditable.**
   The plan allows a narrow marshal/unmarshal fallback but does not say how the implementer will know whether representative tool responses used it. Add a checklist item to document any fallback class in `STATUS.md` after implementation and to ensure the Step 1 golden cases (`get_activities`, `get_fitness`, wrapper rows, provenance metadata) do not depend on the old full-response marshal/unmarshal path.

## Suggested minimal `STATUS.md` update

Add a short Step 3 implementation note under Decisions/Notes with bullets like:

- `toJSONValue` allocates fresh maps/slices for every container; shaper never mutates caller-owned input, including `include_full` raw maps.
- Fast path covers plain structs/maps/slices/pointers/interfaces/primitives used by tool DTOs; values implementing `json.Marshaler`/`encoding.TextMarshaler` use a narrow per-value fallback, with `json.RawMessage` decoded as JSON rather than reflected as bytes.
- Add converter tests for tag/omitempty/deep-copy/fallback behavior and walker tests for null stripping, debug removal, provenance `fetched_at` preservation, and scale collection skipping nested `_meta`.
- After Step 3, record any remaining fallback case in `STATUS.md`; normal golden fixtures must not rely on the old whole-response marshal round-trip.

Once those guardrails are recorded, the implementation direction should be sound.

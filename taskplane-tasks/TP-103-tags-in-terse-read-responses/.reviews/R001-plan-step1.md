# Plan Review: Step 1 — Implement event tag read shaping

**Verdict: Needs revision before implementation**

I could not find a concrete Step 1 plan artifact beyond the high-level checklist in `STATUS.md`. The direction is right, especially because all named tools flow through shared `eventRow`, but the plan needs a few explicit design choices to avoid response-shape regressions.

## Required clarifications

1. **Define tag extraction semantics before coding.**
   - Prefer extracting from `intervals.Event.Raw["tags"]` in `eventRow`, or use a custom tolerant decode type. A plain `[]string json:"tags"` on `intervals.Event` can make the whole event unmarshal fail if upstream ever returns `tags: null`, a non-array, or mixed/non-string elements.
   - State that only an upstream JSON array of strings is emitted; missing, null, non-array, or non-string values should not produce guessed tags.

2. **Preserve explicit empty arrays.**
   - `getEventsRow` must distinguish missing/null from `tags: []`. A `[]string \`json:"tags,omitempty"\`` field will omit explicit empty upstream lists. Use a pointer to a copied slice (or equivalent) so `tags: []` round-trips while missing/null/non-array stays absent.

3. **Make the shared-row impact explicit.**
   - Updating `eventRow` will cover `get_events`, `get_event_by_id`, `add_or_update_event`, and `get_today`, but it also affects any other callers such as delete/apply-training-plan response helpers. The plan should either accept that inherited behavior or list any paths that must remain unchanged.

4. **Move key edge-case tests into Step 1.**
   - Do not defer all null/malformed handling to Step 3. Step 1 should at least add targeted event tests for: tags present in order, explicit empty array emitted, missing/null absent, non-array absent without decode/shape failure, and `include_full` still preserving raw upstream payloads.

## Suggested implementation shape

- Add a small helper near `eventRow`, e.g. `eventTags(raw map[string]any) *[]string`, that copies only valid string arrays and preserves order.
- Add `Tags *[]string \`json:"tags,omitempty"\`` to `getEventsRow` and populate it from that helper.
- If `events.go` is modified, add interval-client tests proving malformed or null `tags` do not break event decoding and raw payload preservation still works.
- Run targeted tests for `internal/tools` event read/write response paths and `internal/intervals` event decoding.

With those details added, the step should be low risk and well-contained.

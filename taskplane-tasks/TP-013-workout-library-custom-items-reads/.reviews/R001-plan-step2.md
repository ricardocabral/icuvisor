# R001 plan review — Step 2

Decision: **APPROVE**

The Step 2 plan is scoped correctly to custom-item reads. It uses the documented read endpoints, keeps custom-item writes out of scope, separates the terse list tool from the full-by-ID tool, includes registry/docs/change-log wiring, and calls out both multiple `item_type` coverage and the v0.4 move from inline schema guidance to `icuvisor://custom-item-schemas`.

No blocking findings.

## Follow-up notes for implementation

- Preserve `content` truly verbatim for `get_custom_item_by_id`. The shared response shaper strips nulls unless `IncludeFull` is true, so the detail response should either call `encodeShaped` with full-preservation semantics or otherwise protect `content` from null stripping. Add a fixture with nested nulls/arrays in `content` so this cannot regress.
- Keep `get_custom_items` terse-by-default. Even if the list endpoint returns `content`, the list rows should omit it and expose only `id`, `name`, `item_type`, and small metadata such as visibility/order/usage/index fields. Full payloads belong in `get_custom_item_by_id`.
- Normalize item IDs defensively in the intervals model, as Step 1 did for folder/workout IDs. Upstream IDs may arrive as strings or numbers; the tool argument should reject an empty `item_id` with a short `UserError` and pass the string path segment unchanged otherwise.
- Avoid overfitting the typed model. Use typed top-level fields for stable row metadata, but retain `Raw map[string]any` and `Content any` so unknown or newly added per-`item_type` shapes round-trip through the detail read.
- Make the inline v0.2 schema guidance useful but not enormous. The first sentence of `get_custom_items` and `get_custom_item_by_id` descriptions should be distinct, and the long-form guidance should identify known `item_type` families from the public docs while stating that the raw `content` object is preserved as returned by intervals.icu.
- Include both layers of tests: `httptest.Server`/fixture coverage for `GET /athlete/{id}/custom-item` and `GET /athlete/{id}/custom-item/{itemId}`, plus tool-shaping tests proving the list omits `content`, the detail includes it, and multiple `item_type` variants preserve nested content shape.
- Remember that `DefaultAPIBaseURL` already includes `/api/v1`; intervals client methods should use existing `doJSON(ctx, ..., "athlete", c.athleteID, "custom-item", ...)` path parts rather than adding another `/api/v1` prefix.

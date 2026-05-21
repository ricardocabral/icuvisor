# R009 Plan Review — Step 4: Hugo rendering

**Verdict:** Approve.

I read `PROMPT.md`, `STATUS.md`, the prior R008 review, and the current `web/` site shape. The revised Step 4 plan in `STATUS.md` addresses the blockers from R008 well enough to proceed.

## What is now resolved

- **TP-050 / Hextra absence is explicit.** The plan records that this worktree does not have the expected Hextra/reference structure and chooses the acceptable fallback: add the minimal non-Hextra reference layout/navigation needed for `/reference/tools/`.
- **Shared rendering approach is appropriate.** Moving the real HTML into shared partials, with a shortcode wrapper for Markdown, avoids trying to call a shortcode from `web/layouts/index.html` and keeps the landing page and reference page on the same `site.Data.tools` source.
- **Landing-page drift guard is planned.** Keeping featured tool names in `web/hugo.toml` params is fine as long as the template validates them with Hugo `errorf` when a configured name is not present in `web/data/tools.json`.
- **Badge/link dependencies are accounted for.** The plan includes adding the minimal tier explanation page/link target and CSS for catalog tables plus tier/safety badges.
- **Readable grouping is covered.** Adding group labels in the rendering layer is the right way to avoid exposing raw slugs such as `workout-library` unless intentionally styled that way.

## Implementation guardrails

These are not blockers, but please keep them in mind while coding Step 4:

1. Add a normal page layout that renders `.Content`; otherwise `web/content/reference/tools.md` and the shortcode will exist but `/reference/tools/` may not display anything useful in the current non-Hextra site.
2. Preserve the generator's deterministic order. Prefer ranging over the already-sorted `site.Data.tools` and emitting a new section when `.group` changes, or explicitly sort groups/names if using Hugo grouping helpers.
3. Make `tool-catalog.html` the thin shortcode wrapper and keep reusable logic in partials so future Hextra migration can move templates without duplicating catalog rendering.
4. Ensure featured-tool validation fails the Hugo build clearly, e.g. with a message telling contributors to update `web/hugo.toml` or rerun `make docs-tools` as appropriate.
5. Keep the new toolset-tier explanation minimal and documentation-focused; broader website documentation remains out of scope for this step/TP.

Proceed with Step 4 implementation.

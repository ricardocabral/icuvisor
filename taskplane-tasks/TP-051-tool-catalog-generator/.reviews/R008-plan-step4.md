# R008 Plan Review — Step 4: Hugo rendering

**Verdict:** Request changes before implementation.

I read `PROMPT.md`, `STATUS.md`, and the current website files under `web/`. The Step 4 checklist in `STATUS.md` restates the prompt, but it does not yet account for the current Hugo site shape or specify how the shared rendering will actually work. There are a few blockers to settle before coding this step.

## Blocking issues

1. **The TP-050 / Hextra precondition is not present in this worktree.**  
   The prompt says TP-050 must have landed and expects a Hextra `reference/` section. The current site is still a custom single-page Hugo setup: `web/layouts/index.html`, `web/content/_index.md`, no theme/module config, no `reference/` content section, and no `_default/single.html` layout. If Step 4 only adds `web/content/reference/tools.md` plus a shortcode, Hugo will not necessarily produce a usable `/reference/tools/` page in this site. The plan needs to either:
   - stop and land/rebase TP-050 first, then implement against the Hextra structure; or
   - explicitly expand Step 4 to add the minimal non-Hextra page layout/navigation/CSS needed for `/reference/tools/` in the current site.  

   Do not silently mix these approaches; the task dependency should be made explicit in `STATUS.md` before implementation continues.

2. **A shortcode cannot be the only shared renderer if the landing page is a layout template.**  
   `{{< tool-catalog >}}` works inside Markdown content, not directly inside `web/layouts/index.html`. The plan says the reference page and landing page should use the same data source/shortcode, but the landing page is a Go template. Use a shared partial for the actual HTML, for example:
   - `web/layouts/partials/tool-catalog.html` for grouped tables,
   - `web/layouts/partials/tool-chips.html` for the featured subset,
   - `web/layouts/shortcodes/tool-catalog.html` as a thin wrapper that calls the partial from Markdown.

   This preserves one rendering implementation where possible and avoids trying to invoke a shortcode from a layout.

3. **The featured-chip source must be chosen and validated.**  
   The current `ToolDescriptor` JSON has no `featured` field, so the safest Step 4 choice is a curated list in `web/hugo.toml` params. The plan should list the intended source and require the template to fail the build with `errorf` if a configured featured tool name is not found in `site.Data.tools`. Otherwise the landing page can still drift by keeping stale names in params.

4. **The badge/link plan depends on pages and styles that do not exist yet.**  
   Step 4 requires tier badges, safety badges, and the `full`/advanced badge linking to an explanation page about toolset tiers. In the current website there is no such explanation page and `web/static/css/style.css` only styles homepage chips, not catalog tables or badges. The plan should identify the target URL for the toolset-tier explanation and the CSS/theme mechanism for amber/write and red/delete badges. If TP-050 provides this, reference that; if not, add the minimal page/style work to this step or defer the link requirement explicitly with reviewer agreement.

5. **Grouped rendering needs deterministic ordering and display labels.**  
   `web/data/tools.json` is sorted by `group` then `name`; the Hugo plan should preserve that determinism and define user-facing group labels such as `workout-library` → `Workout library` and `custom-items` → `Custom items`. Avoid an ad hoc grouping expression that reorders sections unpredictably or renders raw slugs as headings unless that is intentional.

## Suggested plan adjustment

Update `STATUS.md` or a short design note for Step 4 with the concrete rendering approach before implementation:

- Confirm whether TP-050/Hextra is available; if not, decide whether this step is blocked or will add the needed non-Hextra reference layout.
- Implement shared Hugo rendering via partials, with shortcode wrappers only for Markdown content.
- Use `site.Data.tools` as the only source for catalog rows and validate missing data with Hugo `errorf`.
- Configure the landing-page featured subset in one place, ideally `web/hugo.toml`, and fail the build if any configured name is absent from the generated catalog.
- Add or identify the CSS/classes and link target for tier/safety badges.
- Preserve stable group/name ordering and add readable group labels.

Once those details are recorded, Step 4 should be straightforward to implement.

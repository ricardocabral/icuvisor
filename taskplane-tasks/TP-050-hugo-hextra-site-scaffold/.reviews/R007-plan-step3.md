# Plan Review — TP-050 Step 3: Configure navigation, theme params, Pagefind

## Verdict: Approve

The revised Step 3 plan addresses the blockers from R006. It is now explicit about staying on Hextra `v0.9.7` for Hugo `0.139.4` compatibility while adding Pagefind via local partial overrides rather than silently falling back to FlexSearch, and it correctly accounts for local Hugo layout precedence by removing the generic `_default` templates so docs pages can use Hextra.

## What looks good

- Keeps the Hextra/Hugo version decision stable instead of reopening the Step 2 compatibility issue.
- Uses Hextra `v0.9.7`'s actual navbar mechanism: a `menu.main` pseudo-entry with `params.type = "search"`.
- Plans to override the narrow search surface (`partials/search.html` and `partials/scripts/search.html`) rather than forking larger Hextra templates.
- Explicitly removes local `_default/list.html` and `_default/single.html`, which is necessary for Hextra nav/search/footer to render on docs pages.
- Uses `partials/custom/footer.html` and footer params instead of replacing the full Hextra footer.
- Correctly calls out `disableKinds` cleanup and docs-page verification, not just home-page verification.

## Non-blocking implementation notes

1. **Handle both Hextra search partial call sites.** In Hextra `v0.9.7`, `partials/search.html` is called from the navbar with a dict and from the mobile sidebar with no context. Make the override tolerant of both, and avoid duplicate fixed IDs if both header and mobile sidebar render on the same page. A class-based Pagefind UI container plus script initialization over all matching elements is safer than a single `id="search"`.

2. **Avoid unsupported-search warnings.** If `params.search.type = "pagefind"` is added, the stock Hextra script partial would warn that the type is unsupported. Since the plan overrides `partials/scripts/search.html`, ensure the override fully replaces that branch or configure params so the build remains warning-free.

3. **Primary color likely needs CSS, not only TOML.** Hextra `v0.9.7` uses CSS variables such as `--primary-hue`, `--primary-saturation`, and `--primary-lightness`. If there is no theme param for this version, add/override `web/assets/css/custom.css` rather than trying to express the primary colour only in `hugo.toml`.

4. **Keep verification on a docs page.** The bespoke `web/layouts/index.html` will not exercise Hextra's head/nav/footer/search path. Verify an actual section/doc page includes the menu, Pagefind container/assets, and custom footer.

5. **Record Pagefind evidence after implementation.** `STATUS.md` should note that the UI is wired in Step 3 and that the static index generation is completed/verified in Step 5.

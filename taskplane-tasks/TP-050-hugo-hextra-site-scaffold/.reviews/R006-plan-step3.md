# Plan Review — TP-050 Step 3: Configure navigation, theme params, Pagefind

## Verdict: Request changes

The Step 3 checklist identifies the right areas, but it is not yet concrete enough to safely proceed. Two current-repo facts create blockers for this step: the pinned Hextra version does not appear to provide Pagefind, and existing local default layouts will prevent Hextra's navbar/search/footer from rendering on documentation pages.

## Required changes

1. **Resolve the Pagefind vs. Hextra-version conflict before implementation.**
   - `web/go.mod` currently pins `the pinned Hextra module v0.9.7` to stay compatible with the workflow-pinned Hugo `0.139.4`.
   - In the installed `v0.9.7` module, Hextra search is FlexSearch-based (`site.Params.search.type | default "flexsearch"`), and grep finds no `pagefind` integration. Setting `[params.search] enable = true` will not create `public/pagefind/`, so it will not satisfy the task acceptance criteria.
   - The plan needs an explicit path: either upgrade to a Hextra/Pagefind-capable release and update the Hugo workflow pin if required, or add an explicit Pagefind indexing/search integration while keeping Hextra `v0.9.7`. Do not silently substitute FlexSearch for Pagefind unless the task owner amends the acceptance criteria.

2. **Account for local layout precedence.**
   - `web/layouts/_default/list.html` and `web/layouts/_default/single.html` currently override Hextra's `_default` layouts for regular docs pages. If they remain as-is, Hextra menu/search/theme/footer params will not show up on those pages.
   - The plan should explicitly remove or rename those generic local `_default` templates, or otherwise route documentation sections to Hextra's docs layouts (for example with an appropriate `type`/cascade) while keeping only the bespoke home-page override from `web/layouts/index.html`.
   - Verification should include inspecting a rendered docs page, not just `/`, for the Hextra header and search UI.

3. **Specify the Hextra search/menu mechanism.**
   - For Hextra `v0.9.7`, the header renders the search input from a `menu.main` entry whose `params.type = "search"`; `params.search.enable` only controls script loading. The plan should call out the needed search menu item if this version remains in use.
   - Keep the visible top-level menu order required by the prompt: Install, Connect, Tutorials, Guides, Reference, Explain, GitHub. The search pseudo-entry should not replace any required menu item.

4. **Make the footer approach version-aware.**
   - Hextra `v0.9.7` calls `partials/custom/footer.html` from its footer. Prefer adding a small `web/layouts/partials/custom/footer.html` and disabling the default powered-by/copyright bits through params, rather than overriding the full Hextra footer.
   - Preserve the required text and links: `MIT-licensed · not affiliated with intervals.icu`, GitHub, SECURITY.md, and CONTRIBUTING.md.

5. **Revisit `disableKinds`.**
   - `web/hugo.toml` currently disables `sitemap` in addition to taxonomy/term/RSS. The Step 3 plan should either remove `sitemap` from `disableKinds` or document why disabling it remains acceptable for a docs site. Keep `enableRobotsTXT = true`.

## Suggested verification for this step

- Build with the same Hugo version as CI (`0.139.4`) if Hextra remains pinned to `v0.9.7`, or record and test any Hugo version bump.
- Confirm a section/documentation page is rendered by Hextra and includes the configured main menu, footer, and search UI.
- Confirm the search implementation selected for this task actually emits `web/public/pagefind/` if Pagefind remains an acceptance criterion.
- Record the Pagefind decision and evidence in `STATUS.md`.

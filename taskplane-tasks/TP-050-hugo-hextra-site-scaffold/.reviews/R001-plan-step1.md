# Plan Review — TP-050 Step 1: Decide on landing-page strategy

## Verdict: Request changes

The current `STATUS.md` does not yet contain an actual Step 1 plan or decision; it only restates the prompt checklist. For this step, the plan should explicitly choose a landing strategy and state the rationale before proceeding to Hextra installation.

## Required changes

1. **Record the strategy now.** Based on the current files, choose **(A) preserve the bespoke landing page** unless a later Hextra build test proves it impossible.
   - `web/layouts/index.html` is a complete custom Hugo home template with the project’s current branding, download cards, tool catalog partial, footer, and release fetch logic.
   - `web/static/css/style.css` contains substantial custom design-system styling. Rebuilding it with Hextra shortcodes would be design work and is outside this scaffold task.

2. **Be precise about the Hugo mechanism.** The plan should say that keeping `web/layouts/index.html` should continue to override only the Hugo home page, because project layouts take precedence over module/theme layouts. Avoid adding `type`/`layout` frontmatter to `web/content/_index.md` unless testing shows it is necessary; an unnecessary `type` change could affect template lookup in surprising ways.

3. **List the landing dependencies that must be preserved.** The bespoke landing depends on:
   - `web/static/css/style.css`
   - `web/layouts/partials/brand.html`
   - `web/layouts/partials/tool-list.html`
   - `web/data/tools.json` / `params.featuredTools`
   - `params.github`, `params.intervals`, and the favicon/static assets
   The Step 1 plan should call out that these stay in place when Hextra is introduced.

4. **Add a verification note for the next steps.** After installing Hextra, verify that `/` still renders via the bespoke layout and that section/documentation pages render via Hextra. This is the key risk of strategy (A).

## Notes

- The working tree already contains content under `web/content/{install,connect,guides,reference,explain,tutorials}` beyond the “single-page plus stub” described in the task prompt. Step 1 does not need to resolve that, but later steps should avoid deleting or emptying those files unless the task owner confirms that is intended.
- Preserving the landing page means the home page header will remain the custom landing header, while Hextra navigation will appear on Hextra-rendered documentation pages. That is consistent with strategy (A), but it should be stated explicitly so acceptance review does not treat it as accidental.

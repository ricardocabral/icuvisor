# Plan Review: Step 3 — Render docs and refine UX

**Verdict:** Approved with implementation notes

The Step 3 plan is aligned with the accepted data shape from Steps 1–2: keep `tools.json` as the summary catalog, read per-tool details from `web/data/tool_schemas.json`, and render arguments/examples without dumping raw schemas into the always-visible table.

## Notes to carry into implementation

- In `tool-catalog.html`, fail loudly if `site.Data.tool_schemas` is missing, and preferably error if a row in `site.Data.tools` has no matching schema entry. Silent omission would defeat the stale-docs goal.
- Keep the main table compact. A good shape is an always-visible summary row plus collapsed `<details>` for arguments and input examples (or a nested second row), not new wide columns for every schema detail.
- Render examples from the generated `examples` field but label them as input examples. Keep them collapsed by default and JSON-formatted/escaped; do not inline large payloads into the summary cell.
- Cover tools with no arguments or no examples with explicit, tidy fallback text so the docs do not look broken.
- If UX needs styling, use the existing catalog styles in `web/assets/css/custom.css` and record the extra touched file in `STATUS.md`; avoid creating a new partial such as `tool-toc.html` unless there is a concrete need.
- Verification should include `make docs-tools` and `make web-build` (available in the Makefile). If Hugo/Pagefind tooling is unavailable locally, record the exact caveat in `STATUS.md`.

No plan blockers found for Step 3.

# Review R014 — Plan Step 5: Pages workflow

## Verdict: Approved

The Step 5 plan covers the required workflow changes: resolve Hugo Modules before the build, run explicit Pagefind indexing after Hugo, add a pre-upload smoke check, keep the deploy job unchanged, and verify the commands locally. This matches the task scope and the Step 3 decision that Hextra `v0.9.7` needs an explicit Pagefind CLI run.

## What looks good

- It keeps the focus on `.github/workflows/pages.yml`; no deploy target or CNAME changes are planned.
- It preserves the existing Hugo version strategy instead of reopening the Hextra/Hugo compatibility decision from Step 2.
- It accounts for Hugo Modules before `hugo --minify --gc`, which is required now that `web/go.mod` imports Hextra.
- It adds Pagefind after the Hugo build, which is the right order because Pagefind indexes the generated `public/` HTML.
- It adds a smoke check before artifact upload so missing landing output or search output fails CI before deployment.
- It includes local verification of the workflow commands.

## Implementation notes

1. **Do not use update flags in CI.** Use a dependency-resolution command that respects the committed `web/go.mod` / `web/go.sum` pin. Avoid `hugo mod get -u` in the Pages workflow because it can silently move Hextra beyond the compatibility-tested `v0.9.7`.

2. **Be explicit about working directories.** If the build/index/smoke commands run with `working-directory: web`, the smoke paths should be `public/index.html` and `public/pagefind/...`. If the smoke runs from the repository root, use `web/public/...`. Avoid mixing the two forms.

3. **Smoke-check the actual Pagefind assets the site references.** A directory check is acceptable per the prompt, but this site's partials load `/pagefind/pagefind-ui.js` and `/pagefind/pagefind-ui.css`, so prefer checking those files as well as `pagefind.js` or `pagefind-entry.json`.

4. **Consider pinning the Pagefind CLI invocation.** Step 3 verification used `npx --yes pagefind --site public`, which currently works, but it downloads the latest Pagefind on every CI run. A pinned command such as `npx --yes pagefind@<known-good-version> --site public` would make the deploy pipeline less likely to break unexpectedly.

5. **Keep the deploy job unchanged.** The existing `actions/upload-pages-artifact` path should remain `web/public`, and the `deploy` job should continue to depend only on `build`.

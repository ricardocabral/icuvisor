# R008 code review — Step 3

Result: REQUEST_CHANGES

## Findings

1. **Project docs pages still publish Hextra's default SVG/webmanifest favicons.** In `web/hugo.toml:27-31` the navbar logo is pointed at `favicon-32.png`, but Hextra's docs-page head still uses its own `layouts/partials/favicons.html`. A clean build copies the theme's `favicon.svg`, `favicon-dark.svg`, `favicon-16x16.png`, `favicon-32x32.png`, and `site.webmanifest`; because the SVG favicon is listed after `favicon.ico`, browsers that prefer SVG will show Hextra's favicon instead of icuvisor's. This misses the Step 3 requirement to keep the favicon/brand consistent. Add a local `layouts/partials/favicons.html` that references the existing icuvisor assets (`favicon.ico`, `favicon-32.png`, `favicon-192.png`, `apple-touch-icon.png`) or provide icuvisor-branded files for all Hextra favicon/webmanifest names.

2. **Pagefind assets are linked before they exist in a clean build.** `web/layouts/partials/custom/head-end.html:1-3` and `web/layouts/partials/scripts/search.html:1-3` emit `/pagefind/pagefind-ui.css` and `/pagefind/pagefind-ui.js` whenever search is enabled, but `cd web && hugo --minify --gc --destination /tmp/icuvisor-review-public` does not produce a `pagefind/` directory; the generated docs pages have 404s for both assets. If Step 5 is guaranteed to land before this branch is merged, this can be resolved there, but Step 3 currently marks Pagefind as enabled while the rendered search widget cannot load or initialize from a clean Hugo build. Either add the Pagefind indexing/copy step now or keep the status unchecked until the workflow step produces those assets.

## Verification

- Ran `git diff 9fbbbcc..HEAD --name-only` and reviewed the full diff.
- Ran `cd web && hugo --minify --gc` successfully with local Hugo 0.161.1; it reports upstream Hextra deprecation warnings as noted in `STATUS.md`.
- Ran a clean destination build to `/tmp/icuvisor-review-public` and confirmed the generated docs pages include Pagefind links but no `/pagefind/pagefind-ui.{css,js}` output.

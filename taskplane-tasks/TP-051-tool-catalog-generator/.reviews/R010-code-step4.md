# R010 Code Review — Step 4: Hugo rendering

**Verdict:** Request changes.

I reviewed the Step 4 diff against `5c94b77`, read the task prompt/status, and built the site locally. The implementation generally follows the approved non-Hextra plan, but it currently breaks the deployed Pages build.

## Blocking issue

1. **Pages workflow Hugo version cannot render `hugo.Data.tools`.**  
   `web/layouts/partials/tool-catalog.html:1` and `web/layouts/partials/tool-chips.html:1` read the generated catalog with `hugo.Data.tools`. This works with my local Hugo `0.161.1`, but the repository's Pages workflow is pinned to Hugo `0.139.4` (`.github/workflows/pages.yml:24`). Running the same build with that pinned version fails before deployment:

   ```text
   executing "partials/tool-catalog.html" at <hugo>: can't evaluate field Data in type interface {}
   executing "partials/tool-chips.html" at <hugo>: can't evaluate field Data in type interface {}
   ```

   Repro command used:

   ```sh
   tmp=$(mktemp -d)
   curl -sSL -o "$tmp/hugo.tar.gz" the relevant upstream project documentation
   tar -xzf "$tmp/hugo.tar.gz" -C "$tmp" hugo
   (cd web && "$tmp/hugo" --minify --gc)
   ```

   Please switch the partials to the compatibility form already called for in the task text, e.g. `site.Data.tools` or `.Site.Data.tools`, or update and validate the Pages workflow Hugo version as part of this step. Given the repo currently pins `0.139.4`, the safer fix is to use `site.Data.tools` in both partials and rerun `cd web && hugo --minify --gc` with the pinned version.

## Notes

- `cd web && hugo --minify --gc` succeeds with local Hugo `0.161.1`.
- The rendered `/reference/tools/` page contains 40 table rows, matching the 40 entries in `web/data/tools.json`, after building with the newer local Hugo.

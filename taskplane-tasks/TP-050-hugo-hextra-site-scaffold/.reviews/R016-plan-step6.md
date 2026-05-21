# Review R016 — Plan Step 6: Local preview docs

## Verdict: Approved

The Step 6 plan is scoped correctly for the task: update only the website README/local-preview documentation, cover the Hugo Modules workflow, mention Pagefind indexing, and document the Hugo/Hextra versioning details. This satisfies the Step 6 requirements without reopening earlier scaffold, workflow, or content-migration decisions.

## What looks good

- It starts by reviewing the existing `web/README.md`, which is currently stale: it still says `hugo server -D` only and recommends Hugo `0.128+`, while the scaffold now depends on Hugo Modules, Hextra, and Pagefind.
- It plans to document Hugo Modules setup, which is required now that `web/go.mod` imports `the pinned Hextra module`.
- It includes Pagefind indexing in the local docs, matching the Step 3/5 decision to generate Pagefind explicitly with the CLI rather than relying on Hextra's built-in FlexSearch path.
- It plans to record the minimum Hugo extended version, the current Hextra pin, and the upgrade command, which are the important operational details for contributors.

## Implementation notes

1. **Distinguish quick preview from full search preview.** `hugo server -D` is the right fast path for previewing pages/navigation, but the Pagefind assets are generated after a static build. The README should make clear that search requires something like:
   ```bash
   hugo --minify --gc
   npx --yes pagefind --site public
   ```
   followed by serving `public/` with a static file server if the contributor wants to test search locally.

2. **Avoid making routine preview commands silently upgrade Hextra.** The prompt mentions `hugo mod get -u`, but this repo intentionally pinned Hextra to `v0.9.7` for Hugo `0.139.4` compatibility. If `-u` is documented, label it as an intentional dependency refresh/upgrade path, not the safest everyday preview command. For normal setup, prefer commands that respect the committed `go.mod`/`go.sum` pin.

3. **State the version guidance precisely.** The README should say Hugo **extended** is required; Hextra needs extended Hugo >= `0.134`, and this repo's Pages workflow is tested/pinned at Hugo `0.139.4`. Also note that the Hextra pin lives in `web/go.mod`.

4. **Document the intentional upgrade flow.** Include the requested tagged upgrade command, for example:
   ```bash
   hugo mod get -u the pinned Hextra module@<tag>
   hugo mod tidy
   ```
   and remind contributors to run the Hugo build plus Pagefind indexing after changing the pin.

5. **Keep the README scaffold-focused.** Do not add migrated end-user documentation content in this step; this is only local preview/deploy guidance for the Hextra scaffold.

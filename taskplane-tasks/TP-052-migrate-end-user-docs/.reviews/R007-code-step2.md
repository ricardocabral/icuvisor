# Code Review — Step 2: Reference section first

Verdict: **request changes**

Note: the requested baseline hash `d9453e4884b8774090a03436bbe0fc3c10b2d48c` is not present in this worktree. I reviewed against the matching in-branch Step 2 baseline commit `d9453e4183ffcbf41e84ee49a9cc49d407a73b71` (`d9453e4`).

## Finding

### 1. Deleted `toolset-tiers` page leaves generated tool catalog links broken

- **Severity:** Blocking
- **Files:** `web/content/reference/toolset-tiers.md` (deleted), `web/layouts/partials/tool-catalog.html:38`

Step 2 correctly consolidates toolset-tier content into `web/content/reference/safety-modes.md`, but the generated tool catalog partial still links every `full` tier badge to `/reference/toolset-tiers/`:

```go-html-template
<a class="badge badge-tier badge-tier-full" href="{{ "/reference/toolset-tiers/" | relURL }}">full ...</a>
```

On a clean build after deleting `web/content/reference/toolset-tiers.md`, that page no longer exists, so `/reference/tools/` contains broken links for every full-tier tool. I verified with:

```sh
rm -rf web/public && cd web && hugo --minify --gc
grep -R "href=/reference/toolset-tiers/" -n public/reference/tools/index.html
test -e public/reference/toolset-tiers/index.html # exits 1
```

This also violates the R006 implementation note to ensure no existing internal link still points at `toolset-tiers.md` after removing it.

**Suggested fix:** update `web/layouts/partials/tool-catalog.html` to link full-tier badges to the new canonical page, e.g. `/reference/safety-modes/#toolset-tier`, ideally via Hugo `relref` if possible so future moves are validated. Alternatively keep an alias/redirect stub for `/reference/toolset-tiers/`, but the canonical in-site link should point at `safety-modes`.

## Other checks

- `cd web && hugo --minify --gc` otherwise builds successfully.
- The new CLI page's verbatim help block matches `internal/app/testdata/help.golden`.
- The safety-mode defaults and invalid-value fallback documented in `reference/safety-modes.md` match `internal/safety/mode.go` and `internal/safety/toolset.go`.

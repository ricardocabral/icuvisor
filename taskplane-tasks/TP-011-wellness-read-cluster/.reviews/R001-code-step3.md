# Code Review: TP-011 Step 3 — Provenance and staleness `_meta` assembly

**Verdict: Revise.**

The wellness provenance assembly is directionally aligned with the approved Step 3 plan, but the response shaper change does not actually preserve `_meta.provenance.*.fetched_at` for the `get_wellness_data` response shape. Because wellness rows are shaped as a row collection, the wrapper-level debug stripping walks into already-shaped rows with a prefixed path and deletes the provenance timestamp before the response is returned.

## Findings

### 1. `_meta.provenance.*.fetched_at` is stripped from wellness row collections

- **Where:** `internal/response/shaper.go:134-138`, `internal/response/shaper.go:325-351`
- **Severity:** Blocking

`shapeRows` first shapes each wellness row relative to the row root, where `_meta.provenance.<field>.fetched_at` is preserved. Then `shapeWrapperRow` calls `dropDebugMetadata(out, "")` on the whole wrapper. During that second pass, the same provenance object is visited at a path like `wellness[0]._meta.provenance.sleepScore`, but `isProvenancePath` only recognizes paths that start exactly with `_meta.provenance`:

```go
func isProvenancePath(path string) bool {
    return path == "_meta.provenance" || strings.HasPrefix(path, "_meta.provenance.")
}
```

So for the real `get_wellness_data` wrapper (`RowCollections: []string{"wellness"}`), `dropDebugMetadata` treats the provenance entry as ordinary nested data and deletes its `fetched_at` key. That violates the Step 3 contract that every provenance entry is `{ source, native_scale, fetched_at }`, and it specifically defeats the approved shaper requirement to preserve provenance timestamps in default mode.

I confirmed this with an ad-hoc response package test: shaping a wrapper with `wellness[0]._meta.provenance.sleepScore.fetched_at` returned the row provenance with only `source` and `native_scale`; `fetched_at` was missing.

**Suggested fix:** Make provenance-path detection row-collection aware (for example, recognize `._meta.provenance` after an array/index prefix as well as root-relative `_meta.provenance`), or avoid running recursive debug stripping over row collections after they have already been shaped. Add a regression test that uses `Options{RowCollections: []string{"wellness"}}` and asserts `wellness[0]._meta.provenance.<field>.fetched_at` survives while ordinary top-level/row-level debug `fetched_at` still gets stripped.

## What I checked

- Ran `git diff 7d511c7d0af68a1a767c9caa270b5608cce39075..HEAD --name-only`.
- Ran `git diff 7d511c7d0af68a1a767c9caa270b5608cce39075..HEAD`.
- Read `PROMPT.md`, `STATUS.md`, `internal/tools/get_wellness_data.go`, `internal/response/shaper.go`, and related wellness/response context.
- Ran `go test ./...` — it passes.
- Ran an ad-hoc local regression test for provenance `fetched_at` inside a row collection — it fails as described above; the temporary test file was removed.

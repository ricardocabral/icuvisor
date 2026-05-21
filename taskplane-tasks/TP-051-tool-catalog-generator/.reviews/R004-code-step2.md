# R004 Code Review — Step 2: Wire `Catalog()` into the registry

**Verdict:** Request changes.

I reviewed `git diff c9448c0..HEAD`, including `internal/tools/catalog.go`, `internal/tools/catalog_test.go`, and the updated task status.

## Findings

1. **Lint is currently failing.**  
   `golangci-lint run ./internal/tools` fails with:
   ```text
   internal/tools/catalog.go:44:2: Consider preallocating tools with capacity 40 (prealloc)
   ```
   Please fix the preallocation warning, or otherwise make the package lint-clean, before marking Step 2 complete.

2. **`Catalog()` duplicates the registry registration list instead of sharing the registry path.**  
   `internal/tools/catalog.go:40-85` hand-recreates the same constructor sequence that lives in `defaultRegistry.Register` (`internal/tools/registry.go:117-236`). This undermines the task's central goal that the in-code registry be the source of truth: future tool additions/removals now require editing both the registry and `catalogTools()`. The parity test catches some drift after the fact, but the implementation itself is still a duplicated catalog. Prefer extracting a shared unexported enumeration helper used by both `Register` and `Catalog()` so there is a single registration sequence.

3. **Missing group metadata can silently pass as `other`.**  
   `toolCatalogGroup` falls back to `"other"` (`internal/tools/catalog.go:118-119`), and `TestCatalogDescriptors` only checks that `Group` is non-empty (`internal/tools/catalog_test.go:35`). A newly registered tool can therefore satisfy registry parity while being emitted to the generated website under an unintended catch-all group. Since the prompt calls for deliberate groups and this step is establishing metadata quality, make the group mapping exhaustive: return an error/panic-free sentinel that tests reject, or have the test assert every descriptor group is in an allowlist and none are `"other"`.

## Checks run

- `go test ./internal/tools` — pass
- `go test ./...` — pass
- `golangci-lint run ./internal/tools` — fail, see finding #1

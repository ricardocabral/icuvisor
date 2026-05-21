# Code Review — TP-079 Step 2

**Verdict:** REVISE

The `get_gear_list` implementation and cache behavior are mostly on the right track, but the change leaves repository tests failing because catalog-related golden/static safety fixtures were not updated for the newly registered tool.

## Findings

### 1. Static safety catalog was not updated for `get_gear_list` — blocks tests

- **Where:** `internal/tools/catalog.go:106` (new registered tool); missing corresponding update in `internal/safety/adversarial_test.go:23`
- **Severity:** Blocking

Registering `get_gear_list` adds one read tool in every delete mode, but `internal/safety/adversarial_test.go` still uses the old static catalog/counts. As a result, `go test ./internal/safety` fails in all modes with one extra registered tool:

```text
registered tool count in mode none = 22, want 21; tools=[... get_gear_list ...]
registered tool count in mode safe = 32, want 31; tools=[... get_gear_list ...]
registered tool count in mode full = 39, want 38; tools=[... get_gear_list ...]
```

Please add `get_gear_list` to the static catalog with `RequirementRead` so the safety matrix continues to assert that the new read-only tool is available independently of delete mode.

### 2. Generated tool catalog golden is stale — blocks full test suite

- **Where:** `internal/tools/catalog.go:106` / `internal/tools/catalog.go:123`; missing corresponding update in `cmd/gendocs/testdata/tools.golden.json`
- **Severity:** Blocking

Because `get_gear_list` is now part of `tools.Catalog()`, `cmd/gendocs` generates a catalog containing the new settings/full/read descriptor, but the checked-in golden still omits it. `go test ./...` fails in `cmd/gendocs`:

```text
--- FAIL: TestRunWritesToolsCatalogGolden
    main_test.go:27: generated catalog differs from golden
```

Please regenerate/update the catalog golden (and any generated docs expected by the repo workflow) so the registered catalog and generated artifacts stay in sync.

## Tests run

- `go test ./internal/tools ./internal/toolcatalog ./internal/mcp` — passed
- `go test ./internal/safety` — failed as described above
- `go test ./...` — failed in `cmd/gendocs` and `internal/safety` as described above

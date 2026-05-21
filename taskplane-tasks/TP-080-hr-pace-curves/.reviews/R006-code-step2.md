# Code Review — TP-080 Step 2

Verdict: **REVISE**

The HR/pace tool implementations follow the existing `get_power_curves` shape closely, and the targeted packages compile. However, the change registers new public tools without refreshing the generated catalog/schema artifacts that this repo gates in tests/CI.

## Findings

### 1. Generated tool catalog golden is stale, so `go test ./...` fails

- **Where:** `internal/tools/catalog.go` adds `get_hr_curves` and `get_pace_curves`; missing corresponding update to `cmd/gendocs/testdata/tools.golden.json` (and likely the generated web catalog artifact in the docs step).
- **Severity:** Blocking

I ran `go test ./...`; it fails in `cmd/gendocs` with `TestRunWritesToolsCatalogGolden` because the live catalog now includes:

- `get_hr_curves`
- `get_pace_curves`

but `cmd/gendocs/testdata/tools.golden.json` still reflects the old catalog. Please regenerate/update the generated catalog golden (and keep generated docs data in sync when that step is in scope) so the full test suite is green.

### 2. Schema stability CI will fail because the new tools have no committed schema snapshots

- **Where:** `internal/toolchecks/schema_stability.go` adds `get_hr_curves` and `get_pace_curves` to the schema catalog; missing `internal/tools/schema_snapshot/get_hr_curves.json` and `internal/tools/schema_snapshot/get_pace_curves.json`.
- **Severity:** Blocking

I ran the repository schema guard:

```sh
go run ./scripts/check_schema_stability.go -baseline-dir internal/tools/schema_snapshot -require-baseline
```

It fails with:

```text
snapshot freshness: FAIL (2 issue(s))
tool=get_hr_curves kind=missing-current-snapshot current=internal/tools/schema_snapshot/get_hr_curves.json generated live registry schema has no committed snapshot
tool=get_pace_curves kind=missing-current-snapshot current=internal/tools/schema_snapshot/get_pace_curves.json generated live registry schema has no committed snapshot
```

Since these are newly registered MCP tools with new input schemas, please run the schema snapshot generator and commit the two new snapshot files.

## Verification run

- `git diff befa55e3ad1f74349574b7648acdbc78f1cf476a..HEAD --name-only`
- `git diff befa55e3ad1f74349574b7648acdbc78f1cf476a..HEAD`
- `go test ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./internal/safety` — pass
- `go test ./...` — fail: stale `cmd/gendocs/testdata/tools.golden.json`
- `go run ./scripts/check_schema_stability.go -baseline-dir internal/tools/schema_snapshot -require-baseline` — fail: missing HR/pace schema snapshots
- `go run ./scripts/check_confusable_names.go` — pass

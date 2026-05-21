# Code review — Step 3: Register tool and tests

**Verdict: Changes requested**

## Findings

### P1 — Full test suite fails because the static safety catalog matrix was not updated

`internal/safety/adversarial_test.go:23`

Registering `compute_activity_segment_stats` adds one read tool to the actual registry, but the adversarial static catalog fixture still has the old tool list. As a result `go test ./...` fails in `TestAdversarialStaticCatalogMatrix` for all delete modes with the registered count one higher than expected. This test is specifically guarding the capability/delete-mode registration matrix, so the new read tool should be added to `v03ToolCatalog` with `RequirementRead` (or the fixture should otherwise be updated consistently) so the matrix validates the new registered catalog.

Observed failure excerpt:

```text
--- FAIL: TestAdversarialStaticCatalogMatrix/none
registered tool count in mode none = 25, want 24; tools=[compute_activity_segment_stats ...]
--- FAIL: TestAdversarialStaticCatalogMatrix/full
registered tool count in mode full = 42, want 41; tools=[... compute_activity_segment_stats ...]
--- FAIL: TestAdversarialStaticCatalogMatrix/safe
registered tool count in mode safe = 35, want 34; tools=[... compute_activity_segment_stats ...]
```

### P1 — Generated tool catalog golden is stale after adding the analyzer group/tool

`internal/tools/catalog.go:104`

The new registration changes `tools.Catalog()` output by adding `compute_activity_segment_stats` under the new `analyzers` group, but `cmd/gendocs/testdata/tools.golden.json` was not regenerated/updated. This makes `go test ./...` fail in `cmd/gendocs` (`TestRunWritesToolsCatalogGolden`). Even if the human-facing reference docs are planned for Step 4, the repository test suite is already red at this step because the code-level generator golden no longer matches the catalog source of truth.

Observed failure excerpt:

```text
--- FAIL: TestRunWritesToolsCatalogGolden
    main_test.go:27: generated catalog differs from golden
```

## Tests run

- `go test ./internal/tools ./internal/toolcatalog` — passed
- `go test ./...` — failed due the two stale catalog fixtures above (`internal/safety`, `cmd/gendocs`)

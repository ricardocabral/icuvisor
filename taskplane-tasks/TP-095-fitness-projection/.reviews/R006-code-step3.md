# R006 Code Review — Step 3: Register and test

**Verdict:** Request changes

## Findings

### 1. Full test suite is broken after registering the new public tool

- **Severity:** High
- **File:** `internal/tools/catalog.go:72`

`get_fitness_projection` is now part of the registered/cataloged tool set, but the dependent static catalog fixtures were not updated. Running `go test ./...` fails in two places:

- `cmd/gendocs`: `TestRunWritesToolsCatalogGolden` reports `testdata/tools.golden.json` is stale and missing the new descriptor.
- `internal/safety`: `TestAdversarialStaticCatalogMatrix` still expects 24/34/41 tools for none/safe/full, but registration now yields 25/35/42.

This means the current Step 3 commit cannot pass the repository test suite/CI. Please update the generated-docs golden and the safety static catalog matrix/count expectations as part of the registration change, or otherwise keep the tool unregistered until the dependent fixtures are updated.

### 2. Step 3 tests still do not cover the known zero-load full-series bug

- **Severity:** Medium
- **File:** `internal/tools/get_fitness_projection.go:81`

The full projection point still serializes `TrainingLoad` as `json:"training_load,omitempty"`. Valid inputs can produce `training_load: 0` (`planned_daily_loads[].training_load: 0` or `recovery_week_load_pct: 0`), but those projected days will omit `training_load` entirely in `include_full:true` responses. Clients then cannot distinguish an explicit zero/rest load from missing data.

The new full-series golden uses `recovery_week_load_pct: 50`, so it never exercises this edge case. Please remove `omitempty` (or use a shape that preserves zero for projected days) and add a test/golden with an explicit zero planned load or 0% recovery week.

### 3. Request schema/user-message contract mismatches remain untested and unresolved

- **Severity:** Medium
- **File:** `internal/tools/get_fitness_projection.go:19`, `internal/tools/get_fitness_projection.go:238`, `internal/tools/get_fitness_projection.go:329`

The public invalid-arguments message tells callers to provide “exactly one `horizon_date` or `horizon_days`”, but the decoder still accepts neither and defaults to 42 days. Also, the JSON Schema advertises `recovery_week_cadence` with `minimum: 0`, which makes `1` look valid to MCP clients, while the decoder rejects anything except `0` or `2-12`.

The Step 3 invalid-input tests cover unsupported model, both horizon fields, and ramp bounds only, so these MCP-facing inconsistencies can ship unnoticed. Please either align the decoder with the public schema/descriptions or align the schema/message with the decoder, and add tests for the chosen contract.

### 4. STATUS records rejected reviews as approved

- **Severity:** Low
- **File:** `taskplane-tasks/TP-095-fitness-projection/STATUS.md:115`

`STATUS.md` still appends review outcomes under `## Notes` as `APPROVE`, including R004 and R005, even though the checked-in review files say `Request changes` / `Needs changes`. The `## Reviews` table is also empty. This makes the task history misleading and hides the unresolved blockers above. Please move these entries into the Reviews table with the actual verdicts and update blockers/discoveries accordingly.

## Verification

Ran:

```sh
go test ./internal/analysis ./internal/tools ./internal/toolcatalog
go test ./...
```

Targeted tests passed. `go test ./...` failed in `cmd/gendocs` and `internal/safety` as described in Finding 1.

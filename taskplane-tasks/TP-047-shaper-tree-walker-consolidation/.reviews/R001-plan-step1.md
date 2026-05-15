# Plan Review — Step 1: Snapshot pre-refactor output

**Decision: Changes requested before implementation proceeds.**

The Step 1 checklist in `STATUS.md` matches the prompt at a high level, but it is not yet a concrete implementation plan. For this task, Step 1 is the safety net for the entire refactor; if the snapshot is not deterministic and reproducible, Step 5 cannot prove byte-identical behavior.

## Blocking concerns

1. **No reproducible snapshot mechanism is specified.**
   The plan says to capture golden fixtures, but not how they will be regenerated or compared after the refactor. Please make Step 1 add an automated golden test or generator alongside the fixtures, not just static output files. Otherwise the fixtures are orphaned snapshots and Step 5's “re-run snapshot fixtures” is ambiguous.

   A good plan should state, for each fixture case:
   - the input value/payload used to call `response.Shape` or the tool-level shaper,
   - the exact `response.Options` / shaping config,
   - the expected golden output file path under `internal/response/testdata/`, and
   - the test command that compares current output to the golden bytes.

2. **The exact five representative cases are not chosen.**
   Step 1 needs to lock down the fixture set before any shaper changes. The current status still says “Pick ~5”. Please name the cases explicitly. At minimum, cover:
   - `get_activities` terse,
   - `get_activities` with `include_full: true` and raw/full payload nulls preserved,
   - `get_fitness`,
   - a wrapper-row / multi-row-collection response such as `get_events`, `get_workout_library`, or `apply_training_plan`, and
   - a provenance case, preferably wellness-style `_meta.provenance.<field>.fetched_at`, because `dropDebugMetadata` has special handling for provenance `fetched_at` paths.

3. **Determinism hazards need to be planned away.**
   Golden outputs will include response-owned metadata added by `internal/response/shaper.go`: `catalog_hash`, `delete_mode`, `toolset`, optional `units`, and possibly debug metadata. The plan should require deterministic setup per fixture:
   - call/reset catalog test metadata so `schema_changed` does not depend on prior test order,
   - set a stable `ServerVersion`, `DeleteMode`, `Toolset`, and `UnitSystem`,
   - if `DebugMetadata` is true, set a fixed `FetchedAt` instead of relying on `time.Now()`, and
   - avoid parallelizing cases that mutate package-level catalog runtime unless each case fully isolates/reset it.

4. **Golden comparison format is not defined.**
   “Byte-identical” should mean canonical JSON bytes for the shaped response, not Go map iteration artifacts or pretty-print differences. The Step 1 plan should specify whether fixtures are compact JSON or indented JSON and how the test canonicalizes output before comparison.

## Non-blocking recommendations

- Use synthetic fixtures or existing test fixtures/fake clients only; do not hit intervals.icu and do not include real athlete data.
- Keep Step 1 commits limited to snapshot infrastructure/fixtures plus `STATUS.md`. Do not touch `internal/response/shaper.go` until after the fixture commit is made.
- If the snapshot test lives in `internal/response`, synthetic DTOs that mirror representative tool response shapes may be simpler and avoid importing `internal/tools` (which already imports `response`). If the intent is to exercise actual tool shapers, place the test carefully to avoid import cycles and still keep the golden files under `internal/response/testdata/` as required.
- Include a fixture that has caller-supplied `_meta` fields as well as response-owned keys, so the refactor cannot accidentally preserve/overwrite the wrong metadata.

## Suggested update to `STATUS.md`

Before implementing, add a short “Step 1 plan” note naming the fixture files and the generation/comparison strategy. Example shape:

- `get_activities_terse.golden.json` — synthetic activities response, `include_full=false`, row collection `activities`.
- `get_activities_full.golden.json` — same family, `include_full=true`, raw/full payload with nulls.
- `get_fitness.golden.json` — fitness row collection with missing CTL/ATL/TSB fields.
- `get_events_wrapper.golden.json` — wrapper-row collection case.
- `wellness_provenance.golden.json` — `_meta.provenance.*.fetched_at` preservation case.

Once those details are recorded and the fixtures are backed by a deterministic test, Step 1 is a sound plan.

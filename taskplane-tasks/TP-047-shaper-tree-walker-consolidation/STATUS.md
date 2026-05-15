# TP-047-shaper-tree-walker-consolidation — Status

**Current Step:** Step 1: Snapshot pre-refactor output
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 4
**Iteration:** 1
**Size:** M

---

### Step 1: Snapshot pre-refactor output

**Status:** ✅ Complete

- [x] Lock deterministic fixture plan: `get_activities_terse.golden.json`, `get_activities_full.golden.json`, `get_fitness.golden.json`, `get_events_wrapper.golden.json`, `wellness_provenance.golden.json`; each uses synthetic DTO input, stable `response.Options`, reset catalog metadata, and canonical indented JSON bytes
- [x] Add automated golden snapshot test/generator that maps each case to its input, exact shaping options, fixture path, and comparison command without hitting the network
- [x] Capture `_meta`-shaped output as golden fixtures under `internal/response/testdata/`
- [x] Commit fixtures before touching shaper code
- [x] Fix R003 catalog runtime isolation so golden snapshots restore default metadata under shuffle/update/early returns
- [x] Convert representative golden inputs (`get_activities` terse/full and `get_fitness`) from maps to typed local DTO structs with JSON tags/omitempty matching tool shapes, then regenerate fixtures

### Step 2: Pick the approach

**Status:** ⏳ Not started

- [ ] Decide typed-shape vs single visitor walker
- [ ] Justify in Decisions below (diff size, blast radius, `include_full` fit)
- [ ] Sketch struct set or visitor signature

### Step 3: Implement

**Status:** ⏳ Not started

- [ ] Remove marshal round-trip from `marshalToJSONValue` on happy path
- [ ] Collapse five near-duplicate recursive walkers
- [ ] Preserve every existing path predicate's semantics

### Step 4: Adjacent P2 cleanups

**Status:** ⏳ Not started

- [ ] Move `defaultScaleLabels` to `internal/response/scales.go`
- [ ] Extract common helper shared by `shapeRow` / `shapeWrapperRow`

### Step 5: Verify byte-identical output

**Status:** ⏳ Not started

- [ ] Re-run snapshot fixtures; diff must be empty
- [ ] If diff non-empty, stop and resolve

### Step 6: Build / test / lint

**Status:** ⏳ Not started

- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] Eyeball-benchmark large `include_full` response; must not regress

---

## Decisions

_Record typed-shape vs single-walker choice in Step 2, with rationale (diff size, `include_full` handling, mirror risk with `internal/intervals/`)._

_Record any narrow case where the marshal round-trip survives, with rationale._

## Notes

_Add notes as work progresses._

| 2026-05-15 17:49 | Code review R003 | Added blocking revision items: isolate catalog runtime state and use typed DTO inputs for activity/fitness snapshots so JSON tags and omitempty are locked. |
| 2026-05-15 17:44 | Plan review R001 | Added blocking Step 1 plan items: deterministic named fixtures, automated regeneration/comparison, stable metadata setup, canonical JSON; using synthetic fixtures to avoid network/tool import cycles. |
| 2026-05-15 17:43 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 17:43 | Step 1 started | Snapshot pre-refactor output |
| 2026-05-15 17:47 | Review R001 | plan Step 1: REVISE |
| 2026-05-15 17:49 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 17:56 | Review R003 | code Step 1: REVISE |
| 2026-05-15 18:00 | Review R004 | code Step 1: APPROVE |

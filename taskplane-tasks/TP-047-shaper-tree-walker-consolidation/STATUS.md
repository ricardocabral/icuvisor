# TP-047-shaper-tree-walker-consolidation — Status

**Current Step:** Step 1: Snapshot pre-refactor output
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

---

### Step 1: Snapshot pre-refactor output

**Status:** ⏳ Not started

- [ ] Pick ~5 representative tool responses (terse + `include_full` + wrapper-row + provenance)
- [ ] Capture `_meta`-shaped output as golden fixtures under `internal/response/testdata/`
- [ ] Commit fixtures before touching shaper code

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

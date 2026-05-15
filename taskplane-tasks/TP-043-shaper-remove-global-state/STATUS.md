# TP-043-shaper-remove-global-state — Status

**Current Step:** Step 1: Audit reads
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** S

---

### Step 1: Audit reads

**Status:** ⏳ Not started

- [ ] Grep all readers/writers
- [ ] Decide `Options` construction site

### Step 2: Refactor

**Status:** ⏳ Not started

- [ ] Add fields to `Options`
- [ ] Update `addCommonMeta`
- [ ] Delete globals, `init()`, setters
- [ ] Update call sites

### Step 3: Tests

**Status:** ⏳ Not started

- [ ] Existing tests pass without `Set*`
- [ ] Add divergent-`Options` regression test

### Step 4: Verify

**Status:** ⏳ Not started

- [ ] Build / test / race / lint
- [ ] No `init()` left in `internal/response/`
- [ ] `_meta` byte-identical

---

## Decisions

_Record `Options` construction site in Step 1._

## Notes

_Add notes as work progresses._

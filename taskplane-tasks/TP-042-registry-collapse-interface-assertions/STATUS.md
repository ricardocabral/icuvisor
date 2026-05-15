# TP-042-registry-collapse-interface-assertions — Status

**Current Step:** Step 1: Map the current assertion chain
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

---

### Step 1: Map the current assertion chain

**Status:** ⏳ Not started

- [ ] Enumerate every `XxxClient` interface in `internal/tools/`
- [ ] Verify all are satisfied by `*intervals.Client`
- [ ] Identify existing unit-test fakes
- [ ] Decide direct-dep vs `Deps` struct

### Step 2: Refactor `Register`

**Status:** ⏳ Not started

- [ ] Change `Register` signature to typed dep
- [ ] Replace assertion blocks with direct constructor calls
- [ ] Preserve delete-mode / toolset / capability gating
- [ ] Fix hardcoded `getAthleteProfileName` error message

### Step 3: Collapse `schemaCatalogClient`

**Status:** ⏳ Not started

- [ ] Replace with minimal fake or real client
- [ ] Snapshot output byte-identical

### Step 4: Tests

**Status:** ⏳ Not started

- [ ] `make test` + `make test-race`
- [ ] Schema-stability snapshot unchanged
- [ ] Add regression guard test for full registration

### Step 5: Verify

**Status:** ⏳ Not started

- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] Manual `list_tools` parity check

---

## Decisions

_Record dep-shape decision (direct vs `Deps` struct) in Step 1._

## Notes

_Add notes as work progresses._

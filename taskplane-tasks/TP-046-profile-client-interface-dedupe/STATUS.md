# TP-046-profile-client-interface-dedupe — Status

**Current Step:** Step 1: Confirm the duplication is exact
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** S

---

### Step 1: Confirm the duplication is exact

**Status:** ⏳ Not started

- [ ] Diff the two `ProfileClient` declarations; confirm or normalize
- [ ] Confirm `*intervals.Client` satisfies the merged interface
- [ ] Locate and verify test fakes for both consumers

### Step 2: Create the shared declaration

**Status:** ⏳ Not started

- [ ] Create shared interface in chosen home (default: `internal/clients/profile.go`)
- [ ] Remove the two duplicate declarations
- [ ] Update imports in both consumers

### Step 3: Tests

**Status:** ⏳ Not started

- [ ] `go build ./...` — fakes still satisfy
- [ ] `make test` + `make test-race`

### Step 4: Verify

**Status:** ⏳ Not started

- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] `grep -rn "type ProfileClient interface" internal/` returns one hit
- [ ] `git diff --stat` sanity check

---

## Decisions

_Record placement decision (`internal/clients` vs `internal/intervals`) in Step 2, with a one-paragraph justification._

## Notes

_Add notes as work progresses. If TP-042 lands first, note any textual merge resolved here._

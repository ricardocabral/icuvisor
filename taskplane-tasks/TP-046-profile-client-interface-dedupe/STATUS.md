# TP-046-profile-client-interface-dedupe — Status

**Current Step:** Step 1: Confirm the duplication is exact
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 1
**Review Counter:** 1
**Iteration:** 1
**Size:** S

---

### Step 1: Confirm the duplication is exact

**Status:** 🟨 In Progress

- [ ] Diff the two `ProfileClient` declarations; confirm or normalize
- [ ] Inventory all `ProfileClient` usages under `internal/` and record notable non-declaration consumers
- [ ] Confirm `*intervals.Client` satisfies the merged interface
- [ ] Locate and verify test fakes/stubs across tools, resources, MCP tests, and toolchecks

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

| 2026-05-15 13:18 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 13:18 | Step 1 started | Confirm the duplication is exact |
| 2026-05-15 13:21 | Review R001 | plan Step 1: REVISE |

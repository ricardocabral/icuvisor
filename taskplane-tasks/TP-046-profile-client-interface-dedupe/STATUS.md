# TP-046-profile-client-interface-dedupe — Status

**Current Step:** Step 4: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-15
**Review Level:** 1
**Review Counter:** 4
**Iteration:** 1
**Size:** S

---

### Step 1: Confirm the duplication is exact

**Status:** ✅ Complete

- [x] Diff the two `ProfileClient` declarations; confirm or normalize
- [x] Inventory all `ProfileClient` usages under `internal/` and record notable non-declaration consumers
- [x] Confirm `*intervals.Client` satisfies the merged interface
- [x] Locate and verify test fakes/stubs across tools, resources, MCP tests, and toolchecks

### Step 2: Create the shared declaration

**Status:** ✅ Complete

- [x] Create shared interface in chosen home (default: `internal/clients/profile.go`) with compile-time producer assertion
- [x] Remove the two duplicate declarations while preserving a deliberate `tools.ProfileClient` compatibility alias for internal tools consumers
- [x] Update resource consumer imports to use the shared interface
- [x] Record placement decision in `STATUS.md` and add `CHANGELOG.md` `[Unreleased]` entry

### Step 3: Tests

**Status:** ✅ Complete

- [x] `go build ./...` — fakes still satisfy
- [x] `make test` + `make test-race`

### Step 4: Verify

**Status:** ✅ Complete

- [x] `make build` / `test` / `test-race` / `lint`
- [x] `grep -rn "type ProfileClient interface" internal/` returns one hit
- [x] `git diff --stat` sanity check

---

## Decisions

- Step 2 placement: chose `internal/clients/profile.go` instead of `internal/intervals` because `ProfileClient` is a small consumer-facing interface shared by tools and resources, and a neutral internal package avoids cross-importing one consumer package from another while giving likely future shared client interfaces a clear home. The package remains internal-only, imports the concrete intervals response type, and includes a compile-time assertion that `*intervals.Client` satisfies the shared contract.

## Notes

- Step 1 exactness: the two declarations have the same single method set and signature, `GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error)`. The only difference is consumer-specific doc wording (`for tools` vs `for resources`); the shared declaration should preserve both consumers in its doc note.
- Usage inventory (`git grep '\bProfileClient\b' internal`): beyond the two declarations, `internal/tools` uses the type throughout registry constructors and tool handlers, `internal/resources/registry.go` exposes it in resource registry construction, and `internal/toolchecks/schema_stability.go` asserts `schemaCatalogClient` satisfies `tools.ProfileClient`. Step 2 should keep `tools.ProfileClient` as a compatibility alias or update every internal tools reference plus the schema check deliberately.
- Producer confirmation: `internal/intervals/client.go` defines `func (c *Client) GetAthleteProfile(ctx context.Context) (AthleteWithSportSettings, error)`, matching the merged interface exactly. A Step 2 compile-time assertion can confirm `*intervals.Client` against the shared package without introducing cycles.
- Fake/stub inventory: `internal/tools/get_athlete_profile_test.go` has `fakeProfileClient` with the exact method and several tools tests embed it (for example `fakeActivitiesProfileClient`); resources tests define `fakeAthleteProfileClient`, `blockingAthleteProfileClient`, and `failingBlockingAthleteProfileClient` with the exact method; `internal/mcp/protocol_test.go` defines `testProfileClient`; `internal/toolchecks/schema_stability.go` has `schemaCatalogClient` plus a compile-time assertion currently pointed at `tools.ProfileClient`.

_Add notes as work progresses. If TP-042 lands first, note any textual merge resolved here._

| 2026-05-15 13:18 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 13:18 | Step 1 started | Confirm the duplication is exact |
| 2026-05-15 13:21 | Review R001 | plan Step 1: REVISE |
| 2026-05-15 13:23 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 13:27 | Review R003 | plan Step 2: APPROVE |
| 2026-05-15 13:31 | Review R004 | plan Step 3: APPROVE |

| 2026-05-15 13:33 | Worker iter 1 | done in 927s, tools: 75 |
| 2026-05-15 13:33 | Task complete | .DONE created |
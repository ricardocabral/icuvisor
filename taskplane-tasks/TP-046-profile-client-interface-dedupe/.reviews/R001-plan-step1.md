# Review R001 — Step 1 plan

Decision: changes requested before marking Step 1 complete.

The Step 1 checklist covers the core duplication, but it is under-scoped for this repo layout. `ProfileClient` is declared in `internal/tools/get_athlete_profile.go`, but it is not only consumed by that file: it is the package-level profile interface used by the tools registry and many other tool constructors/handlers. Removing or moving it in Step 2 will affect references across `internal/tools`, plus at least `internal/toolchecks/schema_stability.go` (`tools.ProfileClient`) and MCP tests that pass profile-capable fakes into the registries. Step 1 should explicitly inventory those references, not just the two declaration files and their nearest tests.

Required Step 1 plan amendments:

- Add a usage inventory before implementation, e.g. `git grep '\bProfileClient\b' internal` (or equivalent), and record the notable non-declaration consumers in `STATUS.md`.
- Add a fake/stub inventory broader than the two profile-specific test files. At minimum account for:
  - `internal/tools/get_athlete_profile_test.go` `fakeProfileClient`, which is embedded by many tool test fakes.
  - `internal/resources/athlete_profile_test.go` `fakeAthleteProfileClient`, `blockingAthleteProfileClient`, and `failingBlockingAthleteProfileClient`.
  - `internal/mcp/protocol_test.go` `testProfileClient`.
  - `internal/toolchecks/schema_stability.go` compile-time assertion against `tools.ProfileClient`, which will need a Step 2 decision if `tools.ProfileClient` disappears rather than becomes an alias.
- Record the current exactness result: the method set/signature is identical, while the doc notes differ only by consumer wording (`tools` vs `resources`). The shared doc note should preserve the fact that both tools and resources consume it.
- Record how `*intervals.Client` satisfaction was confirmed. Its `GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error)` method in `internal/intervals/client.go` matches the interface; a compile-time assertion in a non-cyclic test/package would be a good follow-up in Step 2, but Step 1 should at least note the confirmation.

Without these additions, the plan risks treating Step 2 as a two-file import update when it may require either a deliberate `tools.ProfileClient` compatibility alias or updates across all tools-package references and schema checks. The task is still small, but the discovery step should make that choice explicit before edits begin.

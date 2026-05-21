# Review R002 — Step 1 plan

Decision: approved for Step 1 discovery.

The Step 1 plan has been broadened enough to address the R001 concern. It now explicitly includes a repo-wide `ProfileClient` usage inventory, confirmation of the concrete `*intervals.Client` method, and verification of fakes/stubs beyond the two nearest profile files. That is the right scope before choosing the Step 2 migration shape, because `tools.ProfileClient` is package-level infrastructure used by the registry and many tool handlers, not only by `get_athlete_profile.go`.

A couple of execution notes for Step 1, not blockers:

- Use both `git grep -n '\bProfileClient\b' internal` and a test-fake search such as `git grep -n 'GetAthleteProfile' -- 'internal/*_test.go' 'internal/**/*_test.go'` so the inventory catches fakes that satisfy the interface without naming `ProfileClient`. In this tree that includes the MCP protocol fake and tool catalog/capability test fakes in addition to the profile-specific tests.
- Record the exactness result in `STATUS.md`: the two declarations have the same single method, `GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error)`, and differ only in doc-note wording (`tools` vs `resources`).
- Record the Step 2 compatibility implication from the inventory, especially `internal/toolchecks/schema_stability.go` asserting `tools.ProfileClient`. The later implementation can either update that reference or deliberately keep a `tools.ProfileClient` alias, but Step 1 should make the choice visible.

With those notes followed during execution, the plan is appropriately scoped for this small dedupe task.

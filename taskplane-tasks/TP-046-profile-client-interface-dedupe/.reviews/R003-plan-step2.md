# Review R003 — Step 2 plan

Decision: approved for Step 2 implementation.

The Step 2 plan now reflects the important discovery from Step 1: `tools.ProfileClient` is package-level infrastructure used throughout `internal/tools`, not just by `get_athlete_profile.go`. Choosing the default `internal/clients` home plus a deliberate `tools.ProfileClient` type alias is a reasonable small-diff migration shape: it gives the interface one canonical method-set declaration while avoiding a noisy sweep across every tool constructor/handler and the schema stability check.

Execution notes to follow while implementing:

- The compatibility type in `internal/tools/get_athlete_profile.go` must be an alias, not another interface declaration, e.g. `type ProfileClient = clients.ProfileClient`. Re-declaring `interface { ... }` would keep the original drift risk and fail the spirit of the task even if tests pass. Keep an exported doc comment starting with `ProfileClient` so `revive` stays clean.
- Include `internal/resources/registry.go` in the edit scope. Once the resource-local interface is removed, `NewRegistryWithOptions(profileClient ProfileClient, ...)` needs to use the shared type directly (or otherwise be covered by an explicitly justified alias). The current plan says to update resource consumers; this file is one of them.
- The new `internal/clients/profile.go` should preserve the exact method signature and include the producer assertion there if desired: `var _ ProfileClient = (*intervals.Client)(nil)`. That package already needs `internal/intervals` for the return type, so this should not introduce a cycle.
- Record the placement decision in `STATUS.md` before closing the step. The justification can be short: `internal/clients` is a neutral internal home for a cross-consumer interface and avoids cross-importing tools/resources or promoting anything to `pkg/`.
- Add a `CHANGELOG.md` entry under `[Unreleased]` / `Changed` (creating the subsection if necessary), not under the long `Added` list.

The final implementation should still satisfy the task acceptance checks: exactly one `type ProfileClient interface` under `internal/`, both tools/resources reaching the shared interface, unchanged fakes continuing to compile, and the planned build/test/lint verification in later steps.

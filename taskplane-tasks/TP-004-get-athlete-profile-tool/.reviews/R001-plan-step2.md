# Plan Review: TP-004 Step 2 — implement the typed tool

## Verdict

**Approved for coding Step 2.** The updated `STATUS.md` Step 2 design now addresses the earlier planning gaps: concrete files/constructor, fakeable intervals dependency, version propagation, strict argument validation, sanitized user errors, and the Step 2/Step 3 boundary are all explicit enough to implement against.

## Findings

No blocking findings.

## Notes for implementation

- Treat absent or empty MCP arguments as `{}` so `get_athlete_profile` can be called with no arguments and still default `include_full` to `false`. This is consistent with the Step 1 contract’s “one optional argument only” behavior.
- Ensure the concrete registry/tool constructor cannot lead to a nil dependency panic. Either return a registration error when `profileClient` is nil or have `NewRegistry` return an error-bearing constructor variant. This preserves the repository rule of no `panic` outside `main`.
- When using `json.Decoder.DisallowUnknownFields()`, also reject trailing JSON tokens if practical, so runtime validation matches `additionalProperties: false` as closely as possible.
- Keep the first sentence of the registered description aligned with the Step 1 contract exactly enough to distinguish profile/thresholds/zones from activities, wellness, fitness trends, events, and workouts.
- It is acceptable that app-level intervals-client instantiation and MCP stdio end-to-end wiring are deferred to Step 5, as long as Step 2 leaves the planned `tools.NewRegistry(profileClient, version)` hook usable by tests and later wiring.

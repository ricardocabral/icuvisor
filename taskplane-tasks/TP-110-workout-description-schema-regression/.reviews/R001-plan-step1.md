# Plan Review R001 — Step 1

Verdict: approved with guardrails.

Recommended invariant location: add a focused test in `internal/tools` (for example `catalog_test.go` or a small new `metadata_invariants_test.go`) that registers the real tool catalog via `NewRegistryWithOptions(... ModeFull, ToolsetFull ...)` and inspects the collected `Tool` metadata. This is preferable to testing only snapshot files because it covers live registered tool descriptions and input schemas.

Required coverage for the test:
- Tools: `add_or_update_event`, `create_workout`, `update_workout` exactly.
- Text surfaces: `Tool.Description` plus the top-level input-schema property descriptions for `description` and `workout_doc`.
- Negative assertion: reject normalized/case-insensitive contradictory phrases such as `mutually exclusive`, `cannot both`, `can't both`, `not both`, `only one of`, and `either description or workout_doc`. Avoid overly broad checks like banning the word `only`, since current valid guidance says structured-step `description` is a label/comment only.
- Positive assertion: require coexistence/merge guidance without exact sentence snapshots, e.g. `when both are supplied` / `merged with description`, and/or the `icuvisor:steps` sentinel guidance on the relevant fields.

Do not change runtime write behavior in this step. Snapshot/docs refresh should remain Step 2 only if the invariant exposes stale committed text.

Targeted verification remains: `go test ./internal/tools ./internal/toolchecks`.

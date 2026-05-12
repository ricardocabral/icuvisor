# Plan Review — Step 2: Snapshot every tool's argument schema

Decision: **APPROVE**

## Summary

The Step 2 plan is aligned with the prompt: snapshots are generated from the live `tools.Registry`, written one JSON file per tool under `internal/tools/schema_snapshot/`, canonicalized with `encoding/json` indentation/key ordering, and the additive-only snapshot workflow is documented in `CONTRIBUTING.md`.

## Non-blocking notes for implementation

- Make the generator fail loudly on duplicate or empty tool names, and sort tool names before writing files so the script's behavior is deterministic even though each file's JSON is independently canonicalized.
- The fake all-tools client should have compile-time assertions for every current tool client interface, or the script should use a dummy `internal/intervals.Client`, to avoid silently omitting a conditionally registered tool when a new interface is added later.
- Document the exact regeneration command in `CONTRIBUTING.md` (for example, `go run ./scripts/generate_schema_snapshots.go`) and state that only new optional arguments may be added for stable tools; removals, renames, and newly required arguments require a new tool name.
- Consider having the generator clean or report stale files in `internal/tools/schema_snapshot/` so a future removed/renamed tool cannot leave an orphaned snapshot unnoticed before the Step 3 guard is in place.

No blocking issues found for this step's plan.

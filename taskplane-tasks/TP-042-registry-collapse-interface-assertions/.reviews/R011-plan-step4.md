# Plan Review — Step 4: Tests

**Verdict: approve.**

The Step 4 plan covers the right verification targets for this refactor: run the normal and race test suites, prove schema snapshots remain unchanged, and add a registry-level regression guard for the new typed `*intervals.Client` registration path. No blocking plan changes are required before implementation.

## Guardrails for the new registry test

- Add the test in `internal/tools`, not in `internal/toolchecks`, so it exercises `tools.NewRegistry` / `NewRegistryWithOptions` directly with the no-network real client fixture (`newNoNetworkIntervalsClient`).
- Assert the **full production registry surface**, not just the 30-tool schema snapshot allow-list. With a real `*intervals.Client`, the registry should advertise all 38 tools, including the 8 tools intentionally omitted from schema snapshots:
  - `create_custom_item`
  - `update_custom_item`
  - `delete_activity`
  - `delete_custom_item`
  - `delete_event`
  - `delete_events_by_date_range`
  - `delete_gear`
  - `delete_sport_settings`
- Check both exact names and count, and fail on duplicates. Sorting the collected names before comparison will keep the failure readable while still catching dropped/renamed tools.
- Use full-mode/full-toolset options for clarity (`Capability: safety.NewCapability(safety.ModeFull)`, `Toolset: safety.ToolsetFull`), even though the plain `collectingRegistrar` does not perform capability filtering. This makes the test’s intent match “registers every advertised tool.”
- Keep the no-network guarantee: registration must not execute handlers or make HTTP calls. The existing panic `RoundTripper` fixture is the right dependency.
- Do not route this test through the Step 3 schema catalog helper, because that helper deliberately filters to 30 snapshot tools and would miss the exact regression this Step 4 test is meant to catch.

## Verification commands expected

- `make test`
- `make test-race`
- `go run ./scripts/check_schema_stability.go` (or an equivalent existing schema-stability test/script that proves committed snapshots are byte-identical)
- Optionally rerun `go run ./scripts/check_confusable_names.go` if the registry test or catalog names are touched; Step 3 already relied on the same generated catalog path.

If schema snapshot files change during Step 4, treat that as a regression unless there is a separate compatibility plan. This task is still intended to be wiring-only.

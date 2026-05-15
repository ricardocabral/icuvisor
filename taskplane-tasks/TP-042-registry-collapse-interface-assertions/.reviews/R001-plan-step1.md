# Plan Review — Step 1: Map the current assertion chain

**Verdict: request changes to the Step 1 plan before implementation.**

The Step 1 checklist points in the right direction, but it is too shallow for the risks in this refactor. The mapping step needs to produce an explicit registration/test/schema inventory, not just a list of `XxxClient` interfaces, otherwise Step 2 will likely break tests or schema snapshot parity.

## Blocking gaps

1. **The plan does not reconcile `*intervals.Client` with existing registry-level fakes.**
   - Many tests currently call `NewRegistry` / `NewRegistryWithOptions` with fake clients, e.g. `internal/tools/get_activities_test.go:63`, `internal/tools/catalog_tiers_test.go:26`, `internal/tools/list_advanced_capabilities_test.go:20`, and many write/delete tool tests.
   - If Step 1 chooses the default direct `*intervals.Client` dependency, those registry-level tests cannot keep working unchanged. The plan should distinguish:
     - per-tool constructor fakes that should remain unchanged, and
     - registry/catalog fakes that must be migrated to either constructor-direct tests or a dummy real `*intervals.Client` used only for registration.

2. **Schema snapshot parity has an immediate catalog-shape risk.**
   - `internal/toolchecks/schema_stability.go:43` currently registers with `schemaCatalogClient{}`.
   - That fake intentionally implements only a subset of the interfaces. Current snapshots omit several tools that production `*intervals.Client` can register, including `create_custom_item`, `update_custom_item`, and multiple delete tools.
   - Replacing the fake with a real `*intervals.Client` under the current ungated `schemaRegistrar` would change generated snapshot membership, violating the task’s byte-identical acceptance criterion. Step 1 must inventory the current generated schema catalog vs the full production catalog and decide how Step 3 will preserve snapshot output.

3. **The mapping should include constructor coupling and conditional registration semantics.**
   - Some registry blocks are not simple one-interface-to-one-tool mappings:
     - `customItemsClient` is captured and then passed into create/update custom item tools.
     - `get_activity_splits` is only registered when both streams and intervals clients are available.
     - `link_activity_to_event` currently passes optional details/event clients that may be nil depending on assertions.
   - With a direct `*intervals.Client`, these dependencies become non-nil and unconditional unless explicitly preserved. Step 1 should record these cases before changing code.

4. **The dependency-shape decision needs criteria tied to tests and toolchecks.**
   - “Default: direct `*intervals.Client`” is probably fine for production call sites, but Step 1 must document how tests and `toolchecks` will construct a no-network real client, or why a small deps struct is necessary.
   - A dummy real client is feasible because registration should not call the network, but the plan should state the config/HTTP-client guard to ensure accidental handler execution cannot leak to the network.

## Recommended Step 1 deliverable

Before moving to Step 2, produce a table with at least these columns:

- `XxxClient` interface name and source file
- required methods
- `*intervals.Client` method(s) satisfying it
- tool constructor(s) that consume it
- current registry condition/order
- whether `schemaCatalogClient` implements it today
- tests/fakes that currently depend on that interface via `NewRegistry` rather than direct constructor use
- migration action for those tests/fakes

Also include the final dependency decision with rationale, especially how schema snapshot membership remains byte-identical.

## Non-blocking suggestions

- Verification can be made compile-time by adding temporary or committed assertions that `*intervals.Client` satisfies every per-tool interface. If committed, prefer a focused `_test.go` or a small internal assertion block that does not add new production behavior.
- Include `NewRegistry` and `NewRegistryWithOptions` call-site inventory in Step 1, not Step 2, because the public internal constructor signature drives most of the blast radius.


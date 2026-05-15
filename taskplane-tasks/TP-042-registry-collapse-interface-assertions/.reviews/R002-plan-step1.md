# Plan Review — Step 1: Map the current assertion chain

**Verdict: approved to proceed with Step 1.**

The revised Step 1 plan in `STATUS.md` addresses the blocking gaps from R001. In particular, it now explicitly covers:

- the full `XxxClient` interface inventory, including required methods, constructor usage, registry order/conditions, and `schemaCatalogClient` coverage;
- verification that `*intervals.Client` satisfies the interfaces;
- separation of per-tool fakes from registry-level `NewRegistry` / `NewRegistryWithOptions` fake call sites;
- schema snapshot catalog membership versus full production registration, which is the key parity risk when replacing `schemaCatalogClient`;
- special constructor/optional-dependency cases (`customItemsClient`, activity splits, `link_activity_to_event`); and
- the direct-client versus `Deps` decision with test/toolcheck migration and no-network dummy-client rationale.

That is enough detail for the mapping step to discover the real blast radius before Step 2 changes signatures or registration behavior.

## Required Step 1 output before moving to Step 2

Please make sure the mapping is recorded in `STATUS.md` notes or a linked artifact before implementation starts. The deliverable should be concrete enough to drive the refactor, not just checked boxes. At minimum, keep the R001 table shape:

- interface name and source file;
- required methods and the corresponding `*intervals.Client` methods;
- constructor(s) consuming the interface;
- current registry condition/order;
- `schemaCatalogClient` coverage today;
- registry-level tests/fakes affected by a typed registry dependency; and
- migration action for each affected test/toolcheck.

## Non-blocking additions

1. Include each tool's `Requirement` and `Toolset` in the inventory. Even though filtering happens downstream in the MCP registrar, changing which constructors are called can change the catalog collected for `list_advanced_capabilities` and schema generation.
2. Flag capability-sensitive schemas/descriptions, especially `apply_training_plan` and `update_sport_settings`, because the registry passes `Capability` into those constructors and schema snapshots must remain byte-identical.

With those additions folded into the mapping artifact, the plan is ready for Step 1 execution.

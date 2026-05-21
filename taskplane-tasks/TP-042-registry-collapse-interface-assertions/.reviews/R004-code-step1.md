# Code Review — Step 1: Map the current assertion chain

**Verdict: approve Step 1.**

The updated `STATUS.md` resolves the factual gaps from R003 and is now detailed enough to drive the registry refactor. The inventory covers all 33 tool client interfaces, records the structural `ApplyTrainingPlanClient` coverage, calls out the schema snapshot/full-production catalog mismatch, and includes the previously missed toolchecks and MCP protocol dispatch migrations.

## Non-blocking follow-up

- `STATUS.md:119` still says Step 3 can either keep a smaller schema fake/adapter or use a schema-only registrar filter. Given the decision at `STATUS.md:62` to make `NewRegistry` / `NewRegistryWithOptions` accept a direct `*intervals.Client`, a fake/adapter that intentionally omits interfaces will no longer be type-compatible with the registry. In practice the viable schema-parity path is a no-network `*intervals.Client` plus an allow-list/filtering registrar (shared by `GenerateSchemaSnapshots` and `GenerateToolCatalog`). Please tighten that wording before or during Step 3 to avoid reintroducing an interface-shaped registry seam.

Also consider marking Step 1 complete in `STATUS.md` when moving on; all Step 1 boxes are checked.

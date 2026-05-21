# Plan Review — Step 2: Refactor `Register`

**Verdict: approve.**

The updated Step 2 plan resolves the blocking ambiguity from R005. It now keeps the `Registry.Register(context.Context, Registrar) error` interface stable, moves the typed dependency to `NewRegistry` / `NewRegistryWithOptions` and `defaultRegistry`, calls out the optional collaborator/custom-item couplings, and explicitly covers per-tool error wrapping and compile migration after the constructor signature change.

## Implementation guardrails

- Change only the registry constructors/field to `*intervals.Client`; do not add a client parameter to `Register` or introduce a DI/service-locator abstraction.
- Keep registration order byte-for-byte equivalent to the current production order. With a real `*intervals.Client`, every current assertion branch is effectively taken today, so the direct calls should mirror that sequence.
- Use a small helper for registration errors, e.g. wrapping as `registering <tool name>: %w`, and use it for `icuvisor_list_advanced_capabilities` too while still adding that tool directly to `collector.downstream` so it is not included in its own catalog.
- Preserve the current collector semantics: normal tools should still be collected before/while they pass through the downstream registrar, because `icuvisor_list_advanced_capabilities` depends on the full metadata catalog even when the downstream registrar filters by toolset/capability.
- For special couplings, pass the same real client for:
  - custom-item create/update schema validation collaborator,
  - activity details collaborators for intervals/messages/link,
  - event collaborator for link,
  - intervals collaborator for splits.
- Do not add registry-side delete/toolset filtering. Keep capability/toolset enforcement in the downstream registrar and per-tool metadata/schema as it is now.
- Keep Step 2 buildable after the signature change. Registry-level fake users should either move to a no-network `*intervals.Client` for catalog/list-tools coverage or instantiate narrow per-tool constructors directly for handler behavior. The MCP profile dispatch test should use an `httptest.Server`-backed real intervals client as noted in `STATUS.md`.

## Watch for Step 2 / Step 3 boundary

The typed constructor change will force `internal/toolchecks` call sites to be touched even before the larger `schemaCatalogClient` collapse. Avoid replacing the schema snapshot path with an unrestricted real client unless it is paired with an allow-list/filtering registrar; otherwise the catalog expands from the current 30 snapshots to the full 38 production tools and violates the byte-identical schema requirement. Leaving the old `schemaCatalogClient` declarations in place until Step 3 is fine as long as the tree compiles.

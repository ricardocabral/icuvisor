# Plan Review — Step 2: Refactor `Register`

**Verdict: request changes before implementing Step 2.**

The Step 1 inventory is strong, but the Step 2 plan in `STATUS.md` is still too ambiguous for the highest-risk part of this refactor. In particular, it says to “Change `Register` signature to typed dep”, while the recorded Step 1 decision says the typed dependency should be on `tools.NewRegistry` / `tools.NewRegistryWithOptions`. Please tighten the Step 2 plan before changing code.

## Blocking clarifications needed

1. **Keep the public `Registry` interface shape stable unless there is a deliberate reason not to.**
   - The existing interface is `Register(context.Context, Registrar) error`, and `internal/mcp` consumes registries through that shape.
   - The Step 1 decision chose direct `*intervals.Client` for `NewRegistry` / `NewRegistryWithOptions`; Step 2 should explicitly say: change those constructors and `defaultRegistry` storage to `*intervals.Client`, not add a client parameter to `Register`.
   - This avoids an unnecessary blast-radius expansion across `mcp.Options`, tests, and non-tool registries.

2. **Spell out the direct wiring pattern, including optional collaborators.**
   - Replace `profileClient` with a typed `client *intervals.Client` field.
   - Keep the nil-client guard, but update the error to refer to the registry/client rather than hardcoding `get_athlete_profile`.
   - Register constructors in the exact current order, passing the same `client` anywhere the constructor expects a narrow interface.
   - Explicitly preserve the special couplings from Step 1: pass `client` as the `CustomItemsClient` collaborator for create/update custom item tools; pass it as the optional activity/event/details collaborators for intervals/messages/link/splits; and keep `icuvisor_list_advanced_capabilities` added directly to `collector.downstream` after the catalog is collected.

3. **Define the AddTool error-wrapping approach before the mechanical edit.**
   - The prompt requires the failing tool name in wiring errors. The plan should call for a small helper such as `add(tool Tool) error` that wraps downstream failures as `fmt.Errorf("registering %s: %w", tool.Name, err)`.
   - Use the same wrapping for `icuvisor_list_advanced_capabilities` while still avoiding adding that tool to the collected catalog.

4. **State how Step 2 will keep the tree buildable after the signature change.**
   - Changing `NewRegistry` / `NewRegistryWithOptions` to `*intervals.Client` will immediately break the registry-level fake call sites inventoried in Step 1 (`catalog_tiers_test.go`, `list_advanced_capabilities_test.go`, `internal/mcp/protocol_test.go`, many per-tool registration tests, and toolchecks until Step 3).
   - The Step 2 plan should either include the minimal compile migration for those call sites (dummy no-network `*intervals.Client` where only catalog registration is needed, direct `newXxxTool` constructors or an `httptest.Server` for handler-dispatch tests), or explicitly state that Step 2 and Step 3 are intended to be implemented in one non-compiling work segment. Prefer keeping each step compileable.

## Non-blocking implementation notes

- Do not add registry-side safety filtering. Continue to rely on the downstream registrar for delete-mode/capability/toolset gating; the registry should only construct the same catalog metadata in the same order.
- Add `internal/intervals` to `registry.go` only for the typed constructor/field; do not introduce a new DI/service-locator abstraction.
- The Step 3 schema-parity plan should use a real no-network `*intervals.Client` plus a shared allow-list/filtering registrar for schema snapshots and confusable-name checks; an interface-shaped fake/adapter will no longer be type-compatible after Step 2.

Once these details are folded into `STATUS.md`, the refactor itself is straightforward and ready to proceed.

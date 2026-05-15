# Plan Review: TP-043 Step 2 Refactor

Verdict: Approved with implementation watch-outs.

The Step 2 plan covers the necessary refactor path: move delete-mode/toolset onto `response.Options`, read those values in `addCommonMeta`, remove the `atomic.Value` globals/`init()`/setters, update the app registry wiring, handle the athlete-profile resource path, and replace the standalone `response.Toolset()` reader in `icuvisor_list_advanced_capabilities` with the captured active toolset. That matches the prompt and Step 1 audit notes.

Watch-outs for implementation:

1. Normalize option values at the response boundary.
   - `response.Options` should use the existing typed values (`safety.Mode`, `safety.Toolset`), but `addCommonMeta` should still emit normalized strings via their `String()` methods or `safety.Parse*` helpers.
   - This preserves the old behavior for zero/invalid values: delete mode defaults to `safe`, toolset defaults to `core`.
   - Avoid emitting raw `string(opts.DeleteMode)` / `string(opts.Toolset)` without normalization.

2. Keep registry metadata consistent with registration capability.
   - Add explicit `DeleteMode`/`Toolset` fields to `tools.RegistryOptions` and `resources.ResourceOptions` as planned.
   - In `tools.NewRegistryWithOptions`, consider deriving the registry delete mode from `opts.DeleteMode`, falling back to `opts.Capability.Mode()` when `DeleteMode` is zero, then `safe`. This prevents non-app callers that already pass a full/delete capability from registering delete tools while responses still report `safe`.
   - `defaultStartServer` should pass the same resolved `deleteMode` and `toolset` values to the MCP server, tool registry, and resource registry.

3. Thread through the shared athlete profile path deliberately.
   - `athleteprofile.Shape` is shared by `get_athlete_profile` and `icuvisor://athlete-profile`; both call paths need the same safety metadata inputs.
   - Update the resource cache reader fields/options as well as the tool constructor path, not only the direct `response.Shape` call.

4. Make `list_advanced_capabilities` fully independent of `internal/response` globals.
   - Its `_meta.toolset` should come from the already-normalized captured `activeToolset` used for `current_toolset`.
   - After the change, `internal/tools/list_advanced_capabilities.go` should not need to import `internal/response` unless another use remains.

5. Use grep after the refactor, before tests.
   - Confirm there are no remaining references to `processDeleteMode`, `processToolset`, `SetDeleteMode`, `SetToolset`, `response.DeleteMode`, or `response.Toolset`.
   - Confirm `func init()` no longer exists in `internal/response`.

No additional plan changes are required before proceeding, but the normalization/fallback details above are important to preserve bit-identical `_meta` output and avoid replacing the current global-state bug with inconsistent per-registry metadata.

# Plan Review — Step 1: Inventory + capture regression-gate tests

Decision: **Changes requested**

The Step 1 checklist is directionally right, but the current `STATUS.md` does not yet contain an implementation plan, declaration inventory, or concrete regression gate. I would not start the split until those are recorded, because this refactor has a high chance of accidentally changing coach-mode visibility semantics.

## Findings

### 1. `STATUS.md` has no actual inventory to review

`STATUS.md` still says `_To be filled during Step 1._`. Step 1 requires every top-level declaration in `internal/mcp/server.go` to be assigned to a target file before the mechanical split. That mapping is the main safety artifact for Step 2.

At minimum, the inventory should include all current `server.go` declarations, including ones not explicitly named in the prompt:

- constants/errors: `genericToolErrorMessage`, `invalidInputToolErrorMessage`, `genericResourceErrorMessage`, `invalidTargetAthleteMessage`, `athleteIDArgumentDescription`, `StreamableHTTPPath`, HTTP timeout constants
- constructor/server surface: `Options`, `Server`, `NewServer`, `CatalogHash`, `newSDKServer`, `capabilityOrSafe`, `toolsetOrCore`
- transport: `Run`, `waitForSessionClose`, `RunStreamableHTTP`, `ServeStreamableHTTP`, `normalizeHTTPServerError`, `transportName`
- recovery: `withPanicRecovery` (especially if TP-060 changed/changes this area)
- tool registrar: `safeRegistrar`, `AddTool`, `toolsetAllows`, `capabilityAllows`, `validateTool`, `logToolHandlerError`, `publicToolErrorMessage`, `toolErrorResult`, conversion helpers
- coach-specific pieces currently inside `safeRegistrar`: `visibilityMiddleware`, `visibleToolNamesForAthlete`, `visibleForAthlete`, `prepareTool`, `coachAllows`, `withSelection`, `resolveToolTarget`, `resolveAthleteID`, `coachFilteredAdvancedCapabilitiesHandler`
- schema helpers: `schemaWithAthleteID`, `stripAthleteID`, `validateToolset`, `validateObjectSchema`, `convertResult`, `convertContent` if the chosen definition of `schema.go` includes SDK conversion
- resource registrar: `safeResourceRegistrar`, `AddResource`, `validateResource`, `isResourceNotFound`, `stringOrDefault`
- advanced-capability helper duplicates: `firstSentence`, `toolRequirement`

Also note that `safePromptRegistrar` is already in `internal/mcp/prompts.go`, not `server.go`, so the Step 1 inventory should reflect the repository's current state rather than blindly following the prompt's older description.

### 2. Regression-gate test discovery is incomplete/misdirected

The status says coach ACL tests are “probably in `mcp/server_test.go` and `internal/tools/list_advanced_capabilities_test.go`”. The important coach wire-behavior tests are actually in `internal/mcp/protocol_test.go`, including:

- `TestProtocolCoachACLFiltersCatalogAndResolvesAthleteID`
- `TestProtocolAthleteScopedSchemasExposeUniformAthleteID`
- `TestProtocolCoachToolsAbsentWhenCoachModeOff`
- `TestProtocolSelectAthleteUpdatesVisibleCatalogAndListAthletes`
- `TestProtocolGateCompositionTruthTable`
- `TestProtocolVisibleCatalogMetadataMatchesToolsListAndSessionsAreIsolated`
- `TestProtocolSelectAthleteMetadataUsesPostGateCatalog`
- `TestProtocolCoachModeEndToEndRoutesSelectedDefaultAndOverrideTargets`
- `TestProtocolAthleteIDRejectionMessageIsEnumerationSafe`
- `TestProtocolAdvancedCapabilitiesUsesCoachFilteredCatalog`

`internal/tools/list_advanced_capabilities_test.go` is still relevant for the non-coach base handler/output contract, but it is not the primary coach ACL gate. `internal/coach/evaluator_test.go`, `internal/coach/config_test.go`, and `internal/toolcatalog/catalog_test.go` should also be listed as supporting gates.

Please record exact test names and the exact command(s) to run before and after each refactor move. A good gate would include the named protocol tests plus the coach/tools packages, not just `server_test.go`.

### 3. The plan should capture the import-cycle risk before choosing the coach seam

The task suggests moving filter logic to `internal/coach` and exposing something like `coach.ToolFilter(... []Tool ...)`. Today `internal/tools` already imports `internal/coach` (`registry.go`, `list_athletes.go`, `select_athlete.go`). If `internal/coach` imports `internal/tools` to accept or return `tools.Tool`, that will create an import cycle.

Even if the detailed seam design is Step 3, Step 1's inventory should record this dependency constraint so the later plan can choose an acyclic API, e.g. a small coach-owned metadata type, callback/predicate over tool name, or keeping `tools.Tool`-aware rendering in `internal/tools` while coach owns only ACL/filter decisions.

### 4. Several declarations have ambiguous destinations and need an explicit call

The target-file list in the prompt does not say where to put some current declarations:

- `CatalogHash` likely belongs with catalog hashing (`catalog_hash.go`) or a small server metadata file if `server.go` must remain constructor-only.
- `withPanicRecovery` / `newSDKServer` may belong in `recover.go` / `server.go`; this depends on TP-060 state and should be recorded.
- `firstSentence` and `toolRequirement` duplicate concepts already present in `internal/tools/list_advanced_capabilities.go`; the inventory should decide whether the moved coach-filtered handler reuses that code or keeps a local equivalent without changing behavior.
- `athleteIDArgumentDescription` is referenced by protocol tests; if it moves to `schema.go`, tests and package visibility still need to work.

## Requested changes to the Step 1 plan

1. Fill `STATUS.md` with a declaration-by-declaration inventory and target file for every top-level declaration in `internal/mcp/server.go`.
2. Add a concrete “regression gate” section listing the exact tests above and the command(s) to run.
3. Record the coach/tools import-cycle constraint as a design note before Step 3.
4. Call out existing files (`internal/mcp/prompts.go`, `internal/mcp/catalog_hash.go`) so the split plan does not create duplicate or stale files unnecessarily.

Once those are in `STATUS.md`, Step 1 will be reviewable and the mechanical split can proceed with much lower risk.

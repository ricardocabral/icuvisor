# TP-064-split-mcp-server-and-coach-extract — Status

**Current Step:** Step 4: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 3
**Review Counter:** 9
**Iteration:** 1
**Size:** L

---

### Step 1: Inventory + capture regression-gate tests

**Status:** ✅ Complete

- [x] List every top-level decl in `mcp/server.go` and assign each to a target file. Record in `STATUS.md`.
- [x] Identify all coach-ACL tests across packages (probably in `mcp/server_test.go` and `internal/tools/list_advanced_capabilities_test.go`). They become the immovable gate — every one must pass after the refactor.
- [x] R001: Record a declaration-by-declaration split inventory for every current top-level declaration in `internal/mcp/server.go`.
- [x] R001: Add a concrete regression-gate section with exact coach ACL/supporting test names and commands.
- [x] R001: Record the `internal/coach` ↔ `internal/tools` import-cycle constraint for the Step 3 seam.
- [x] R001: Call out existing split-adjacent files (`internal/mcp/prompts.go`, `internal/mcp/catalog_hash.go`) to avoid duplicate/stale files.

#### Inventory Notes

Declaration inventory from `internal/mcp/server.go` (853 LOC at Step 1 start):

| Declaration                                                 | Target                                                                                                                                                                            |
| ----------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `genericToolErrorMessage`                                   | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `invalidInputToolErrorMessage`                              | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `genericResourceErrorMessage`                               | `internal/mcp/registrar_resources.go`                                                                                                                                             |
| `invalidTargetAthleteMessage`                               | `internal/mcp/registrar_tools.go` until Step 3; keep the public error string byte-identical if resolver helpers move behind the coach seam.                                       |
| `athleteIDArgumentDescription`                              | `internal/mcp/schema.go` (package-private; protocol tests depend on schema text/shape, not symbol visibility).                                                                    |
| `StreamableHTTPPath`                                        | `internal/mcp/transport.go`                                                                                                                                                       |
| `streamableHTTPSessionTimeout`                              | `internal/mcp/transport.go`                                                                                                                                                       |
| `streamableHTTPShutdownTimeout`                             | `internal/mcp/transport.go`                                                                                                                                                       |
| `snakeCaseToolName`                                         | `internal/mcp/schema.go` shared by tool and prompt validation.                                                                                                                    |
| `Options`                                                   | `internal/mcp/server.go` as constructor surface.                                                                                                                                  |
| `Server`                                                    | `internal/mcp/server.go` as constructor/runtime surface.                                                                                                                          |
| `NewServer`                                                 | `internal/mcp/server.go`; only orchestration/constructor logic should remain here.                                                                                                |
| `(*Server).CatalogHash`                                     | existing `internal/mcp/catalog_hash.go` or a tiny metadata file; not transport/registrar.                                                                                         |
| `(*Server).Run`                                             | `internal/mcp/transport.go`                                                                                                                                                       |
| `waitForSessionClose`                                       | `internal/mcp/transport.go`                                                                                                                                                       |
| `(*Server).RunStreamableHTTP`                               | `internal/mcp/transport.go`                                                                                                                                                       |
| `(*Server).ServeStreamableHTTP`                             | `internal/mcp/transport.go`                                                                                                                                                       |
| `normalizeHTTPServerError`                                  | `internal/mcp/transport.go`                                                                                                                                                       |
| `transportName`                                             | `internal/mcp/transport.go`                                                                                                                                                       |
| `capabilityOrSafe`                                          | `internal/mcp/registrar_tools.go` (also used by `catalog_hash.go`).                                                                                                               |
| `toolsetOrCore`                                             | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `withPanicRecovery`                                         | `internal/mcp/recover.go` (or keep consistent with TP-060 if already landed).                                                                                                     |
| `newSDKServer`                                              | `internal/mcp/server.go` if treated as constructor helper; otherwise `internal/mcp/recover.go` beside `withPanicRecovery`.                                                        |
| `safeRegistrar`                                             | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `(*safeRegistrar).AddTool`                                  | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `(*safeRegistrar).visibilityMiddleware`                     | Step 2: `internal/mcp/registrar_tools.go`; Step 3: reduce to MCP SDK adapter that calls a coach-owned predicate/filter.                                                           |
| `(*safeRegistrar).visibleToolNamesForAthlete`               | Step 2: `internal/mcp/registrar_tools.go`; Step 3: `internal/coach/filter.go` seam should own the visibility decision/list generation without importing `internal/tools`.         |
| `(*safeRegistrar).visibleForAthlete`                        | Step 2: `internal/mcp/registrar_tools.go`; Step 3: `internal/coach/filter.go` or coach predicate wrapper.                                                                         |
| `(*safeRegistrar).prepareTool`                              | Step 2: `internal/mcp/registrar_tools.go`; schema mutation remains `internal/mcp/schema.go`, coach decision should move behind Step 3 seam.                                       |
| `(*safeRegistrar).coachAllows`                              | Step 2: `internal/mcp/registrar_tools.go`; Step 3: coach-owned filter/predicate.                                                                                                  |
| `(*safeRegistrar).withSelection`                            | Step 2: `internal/mcp/registrar_tools.go`; Step 3: keep SDK-session adaptation in MCP, but selection semantics remain `internal/coach`.                                           |
| `(*safeRegistrar).resolveToolTarget`                        | Step 2: `internal/mcp/registrar_tools.go`; Step 3: split request target authorization into coach-owned logic plus MCP raw-JSON/schema handling.                                   |
| `(*safeRegistrar).resolveAthleteID`                         | Step 2: `internal/mcp/registrar_tools.go`; Step 3: coach-owned target resolution/predicate where acyclic.                                                                         |
| `schemaWithAthleteID`                                       | `internal/mcp/schema.go`                                                                                                                                                          |
| `stripAthleteID`                                            | `internal/mcp/schema.go`                                                                                                                                                          |
| `(*safeRegistrar).coachFilteredAdvancedCapabilitiesHandler` | Step 2: `internal/mcp/registrar_tools.go`; Step 3: move behavior to `internal/coach/filter.go` or expose a non-cyclic helper from `internal/tools/list_advanced_capabilities.go`. |
| `firstSentence`                                             | Step 2: `internal/mcp/registrar_tools.go`; Step 3: remove duplicate by reusing or mirroring `internal/tools` behavior without changing output.                                    |
| `toolRequirement`                                           | Step 2: `internal/mcp/registrar_tools.go`; Step 3: remove duplicate by reusing or mirroring `internal/tools` behavior without changing output.                                    |
| `(*safeRegistrar).toolsetAllows`                            | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `(*safeRegistrar).capabilityAllows`                         | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `(*safeRegistrar).validateTool`                             | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `safeResourceRegistrar`                                     | `internal/mcp/registrar_resources.go`                                                                                                                                             |
| `(*safeResourceRegistrar).AddResource`                      | `internal/mcp/registrar_resources.go`                                                                                                                                             |
| `(*safeResourceRegistrar).validateResource`                 | `internal/mcp/registrar_resources.go`                                                                                                                                             |
| `isResourceNotFound`                                        | `internal/mcp/registrar_resources.go`                                                                                                                                             |
| `stringOrDefault`                                           | `internal/mcp/registrar_resources.go`                                                                                                                                             |
| `validateToolset`                                           | `internal/mcp/schema.go`                                                                                                                                                          |
| `validateObjectSchema`                                      | `internal/mcp/schema.go`                                                                                                                                                          |
| `convertResult`                                             | `internal/mcp/schema.go` (SDK/tools conversion bucket from prompt).                                                                                                               |
| `convertContent`                                            | `internal/mcp/schema.go` (SDK/tools conversion bucket from prompt).                                                                                                               |
| `logToolHandlerError`                                       | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `publicToolErrorMessage`                                    | `internal/mcp/registrar_tools.go`                                                                                                                                                 |
| `toolErrorResult`                                           | `internal/mcp/registrar_tools.go`                                                                                                                                                 |

#### Regression Gate

Primary coach ACL/wire-behavior tests in `internal/mcp/protocol_test.go`:

- `TestProtocolCoachACLFiltersCatalogAndResolvesAthleteID`
- `TestProtocolAthleteScopedSchemasExposeUniformAthleteID`
- `TestProtocolCoachToolsAbsentWhenCoachModeOff`
- `TestProtocolSelectAthleteUpdatesVisibleCatalogAndListAthletes`
- `TestProtocolGateCompositionTruthTable`
- `TestProtocolVisibleCatalogMetadataMatchesToolsListAndSessionsAreIsolated`
- `TestProtocolSelectAthleteMetadataUsesPostGateCatalog`
- `TestProtocolCoachModeEndToEndRoutesSelectedDefaultAndOverrideTargets`
- `TestProtocolAthleteIDRejectionMessageIsEnumerationSafe`
- `TestProtocolListAdvancedCapabilitiesVisibilityWithRealRegistry`
- `TestProtocolAdvancedCapabilitiesUsesCoachFilteredCatalog`
- `TestProtocolHiddenFullToolIsAbsentAndUnknown`

Supporting split/regression tests:

- `internal/tools/list_advanced_capabilities_test.go`: `TestListAdvancedCapabilitiesOutputFromCatalog`, `TestListAdvancedCapabilitiesFullModeStatus`, `TestListAdvancedCapabilitiesRejectsArguments`.
- `internal/coach/evaluator_test.go`: `TestEvaluator`, `TestDisabledEvaluatorAllowsAthleteScopedTools`.
- `internal/coach/config_test.go`: `TestParseMode`, `TestValidateConfigNormalizesRosterAndPatterns`, `TestValidateConfigStateMachine`.
- `internal/toolcatalog/catalog_test.go`: `TestValidateACLPattern`, `TestToolSets`.
- `internal/mcp/server_test.go` remains the registrar/transport split sanity suite; no direct coach ACL tests were found there.

Targeted gate command before and after each refactor move:

```sh
go test ./internal/mcp ./internal/tools ./internal/coach ./internal/toolcatalog
```

Optional fast named protocol gate when isolating coach visibility:

```sh
go test ./internal/mcp -run 'TestProtocol(CoachACLFiltersCatalogAndResolvesAthleteID|AthleteScopedSchemasExposeUniformAthleteID|CoachToolsAbsentWhenCoachModeOff|SelectAthleteUpdatesVisibleCatalogAndListAthletes|GateCompositionTruthTable|VisibleCatalogMetadataMatchesToolsListAndSessionsAreIsolated|SelectAthleteMetadataUsesPostGateCatalog|CoachModeEndToEndRoutesSelectedDefaultAndOverrideTargets|AthleteIDRejectionMessageIsEnumerationSafe|ListAdvancedCapabilitiesVisibilityWithRealRegistry|AdvancedCapabilitiesUsesCoachFilteredCatalog|HiddenFullToolIsAbsentAndUnknown)$'
```

Baseline Step 1 gate result: `go test ./internal/mcp ./internal/tools ./internal/coach ./internal/toolcatalog` passed on 2026-05-17.

#### Step 3 Coach Seam Constraint

`internal/tools` already imports `internal/coach` from `registry.go`, `list_athletes.go`, and `select_athlete.go`. Therefore `internal/coach` must not import `internal/tools` to accept/return `tools.Tool`, or the Step 3 lift will create an import cycle. Candidate acyclic seams for Step 3:

- coach-owned filter over a small metadata struct (`Name`, `EffectiveToolset`, `Requirement`, `Description`) defined in `internal/coach`, with `mcp`/`tools` adapting from `tools.Tool`;
- coach-owned predicate/list generator over tool names plus callbacks supplied by `mcp`;
- keep `tools.Tool`-specific advanced-capabilities rendering in `internal/tools`, while `internal/coach` owns only ACL decisions and selected-athlete resolution.

The chosen seam must preserve the current ACL composition: capability gate, toolset gate, then coach ACL; any deny remains final.

#### Existing Split-Adjacent Files

- `internal/mcp/prompts.go` already contains `safePromptRegistrar`, prompt validation, prompt conversion, and prompt error shaping. Step 2 should either leave it in place or deliberately rename to `registrar_prompts.go`; do not create a second prompt registrar file with duplicate helpers.
- `internal/mcp/catalog_hash.go` already contains `CatalogHashOptions`, `ComputeToolCatalogHash`, and `hashToolCatalog`. `(*Server).CatalogHash` can move there if `server.go` is narrowed to constructor surface; do not duplicate catalog hashing helpers.
- `internal/tools/list_advanced_capabilities.go` already owns the base non-coach advanced-capabilities response shapes, status text, argument rejection, and summary/requirement helpers. Step 3 should preserve byte-compatible behavior for coach-filtered output and avoid divergent duplicate logic where possible.

### Step 2: Mechanical file split

**Status:** ✅ Complete

- [x] Move transport declarations into `internal/mcp/transport.go` without changing logic.
- [x] Move schema/validation/SDK conversion helpers into `internal/mcp/schema.go` without changing logic.
- [x] Move tool registrar declarations into `internal/mcp/registrar_tools.go` without changing logic.
- [x] Move resource registrar declarations into `internal/mcp/registrar_resources.go` without changing logic.
- [x] Move panic-recovery constructor helper(s) into `internal/mcp/recover.go` without changing logic.
- [x] Narrow `internal/mcp/server.go` to constructor/server surface and keep existing prompt/catalog-hash files non-duplicated.
- [x] Run the Step 1 regression gate after each concern move.

#### Step 2 Test Evidence

`go test ./internal/mcp ./internal/tools ./internal/coach ./internal/toolcatalog` passed after each Step 2 move: transport, schema, tool registrar, resource registrar, recover helper, and server/catalog narrowing.

### Step 3: Lift coach concerns to `internal/coach`

**Status:** ✅ Complete

- [x] Design the seam: probably `coach.ToolFilter` taking the registry catalog + selection ctx and returning the filtered catalog. Sketch in `STATUS.md` first.
- [x] R006: Extend the coach seam to cover request-time target authorization without introducing a `coach` -> `config` import cycle.
- [x] R006: Preserve exact non-athlete-scoped visibility semantics and advanced-capabilities catalog source/gate order in the Step 3 design.
- [x] R006: Add concrete tests for `coach.ToolFilter`, target authorization, and filtered advanced-capabilities compatibility.
- [x] Move `coachFilteredAdvancedCapabilitiesHandler` to its new home.
- [x] `mcp/registrar_tools.go` calls into `coach` instead of inlining the filter.
- [x] Add/update `internal/coach` coverage for coach-owned visibility filtering.
- [x] Re-run all tests; coach-ACL tests must pass unchanged.
- [x] R008: Restore coach-filtered advanced-capabilities wire metadata/text compatibility (`_meta.toolset`, no invented delete-mode, prior enable-instruction wording) and add assertions.

#### Step 3 Seam Sketch

Use an acyclic coach-owned visibility layer, not `tools.Tool` in `internal/coach`:

- Add `internal/coach/filter.go` with `ToolFilter` wrapping `coach.Evaluator`.
- `ToolFilter.VisibleForAthlete(athleteID, toolName string) bool` owns the always-visible coach/catalog tools (`list_athletes`, `select_athlete`, `icuvisor_list_advanced_capabilities`), preserves the current `Evaluator.Evaluate` behavior that non-athlete-scoped tools are coach-allowed, and applies per-athlete allow/deny only to athlete-scoped tools.
- `ToolFilter.VisibleToolNamesForAthlete(athleteID string, toolNames []string) []string` returns sorted visible names for `select_athlete` metadata without importing `internal/tools`.
- `ToolFilter.AllowedForAny(toolName string) bool` centralizes registration-time coach gating: non-athlete-scoped tools pass, athlete-scoped tools pass only if at least one roster athlete allows them.
- Request-time raw JSON stripping and `config.NormalizeAthleteID` stay in `mcp` to avoid a `coach` -> `config` cycle, but roster/default/selected fallback and ACL authorization move behind a coach-owned API such as `ToolFilter.ResolveTarget(suppliedAthleteID, defaultAthleteID, selectedAthleteID, toolName string, normalize func(string) (string, error)) (string, error)` or `AuthorizeTarget(normalizedAthleteID, toolName string) error`. `mcp` remains responsible for SDK session adaptation and `intervals.WithTargetAthleteID`.
- Preserve gate order and catalog source exactly: `AddTool` first applies capability; capability-allowed + coach-allowed-for-any tools populate the advanced-capabilities source catalog before the toolset gate; the toolset gate controls `tools/list`; per-selected-athlete coach visibility filters `tools/list` and advanced-capabilities rows at call time. `icuvisor_list_advanced_capabilities` is excluded from rows.
- Move the advanced-capabilities handler out of `mcp` by exporting a tools-package helper such as `tools.NewFilteredAdvancedCapabilitiesHandler(catalog, toolset, include func(context.Context, tools.Tool) bool)`. This keeps response rendering beside the base `list_advanced_capabilities` tool, must reuse the existing status text/argument rejection/row sorting/requirement formatting, and avoids an `internal/coach` -> `internal/tools` import cycle.
- `mcp/registrar_tools.go` keeps SDK session/raw-JSON/schema adaptation but calls `coach.ToolFilter` for visibility/target decisions and the tools helper for advanced-capabilities rendering.
- New Step 3 tests must cover `coach.ToolFilter` always-visible tools, non-athlete-scoped allowance, per-athlete allow/deny, unknown athlete denial for athlete-scoped tools, disabled evaluator behavior, `AllowedForAny`, enumeration-safe request-time denial through existing protocol tests, and filtered advanced-capabilities invalid arguments/status/sorting/meta count/selected-athlete filtering.

#### Step 3 Test Evidence

`go test ./internal/mcp ./internal/tools ./internal/coach ./internal/toolcatalog` passed after lifting coach filtering. `docs/coach-mode.md` now notes the exported internal `coach.ToolFilter` seam.

### Step 4: Verify

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` `[Unreleased]` with the internal MCP/coach refactor.
- [x] `make build` / `test` / `test-race` / `lint`.
- [x] `scripts/snapshot_tool_schemas.go` — diff empty.
- [x] Coach-mode integration tests (in TP-039's test set) — all pass.
- [x] No new public-API surface in `internal/mcp` beyond what existed before.

#### Step 4 Verification Evidence

- `make build && make test && make test-race && make lint` passed (`golangci-lint`: 0 issues).
- `go run scripts/snapshot_tool_schemas.go && git diff --exit-code -- internal/tools/schema_snapshot` produced an empty diff.
- Coach integration gate passed: `go test ./internal/mcp -run 'TestProtocol(CoachACLFiltersCatalogAndResolvesAthleteID|AthleteScopedSchemasExposeUniformAthleteID|CoachToolsAbsentWhenCoachModeOff|SelectAthleteUpdatesVisibleCatalogAndListAthletes|GateCompositionTruthTable|VisibleCatalogMetadataMatchesToolsListAndSessionsAreIsolated|SelectAthleteMetadataUsesPostGateCatalog|CoachModeEndToEndRoutesSelectedDefaultAndOverrideTargets|AthleteIDRejectionMessageIsEnumerationSafe|ListAdvancedCapabilitiesVisibilityWithRealRegistry|AdvancedCapabilitiesUsesCoachFilteredCatalog|HiddenFullToolIsAbsentAndUnknown)$' && go test ./internal/coach`.
- Exported `internal/mcp` non-test API remains `StreamableHTTPPath`, `Options`, `Server`, `NewServer`, `CatalogHashOptions`, `ComputeToolCatalogHash`, and existing `(*Server).CatalogHash`.

| 2026-05-17 03:24 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 03:24 | Step 1 started | Inventory + capture regression-gate tests |
| 2026-05-17 03:27 | Review R001 | plan Step 1: REVISE |
| 2026-05-17 03:34 | Review R002 | plan Step 1: APPROVE |
| 2026-05-17 03:37 | Review R003 | code Step 1: APPROVE |
| 2026-05-17 03:39 | Review R004 | plan Step 2: APPROVE |
| 2026-05-17 03:46 | Review R005 | code Step 2: APPROVE |
| 2026-05-17 03:50 | Review R006 | plan Step 3: REVISE |
| 2026-05-17 03:53 | Review R007 | plan Step 3: APPROVE |
| 2026-05-17 04:05 | Review R008 | code Step 3: UNKNOWN |
| 2026-05-17 04:12 | Review R009 | code Step 3: APPROVE |

| 2026-05-17 04:16 | Worker iter 1 | done in 3113s, tools: 188 |
| 2026-05-17 04:16 | Task complete | .DONE created |
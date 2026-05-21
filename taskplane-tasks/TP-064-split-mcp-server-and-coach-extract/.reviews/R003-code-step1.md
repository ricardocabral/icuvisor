# Code Review — Step 1: Inventory + capture regression-gate tests

Decision: **Approved**

No blocking findings.

## What I checked

- Reviewed the full diff from `03141b3bb8807639a38984687dbc6403243f3368..HEAD`; the only changed file is `taskplane-tasks/TP-064-split-mcp-server-and-coach-extract/STATUS.md`.
- Compared the `STATUS.md` declaration inventory against the current top-level declarations in `internal/mcp/server.go`; the inventory covers the current declarations.
- Verified the listed protocol and supporting regression tests exist via `go test -list`.
- Ran the recorded regression gate successfully:

```sh
go test ./internal/mcp ./internal/tools ./internal/coach ./internal/toolcatalog
```

- Ran the optional named protocol gate successfully:

```sh
go test ./internal/mcp -run 'TestProtocol(CoachACLFiltersCatalogAndResolvesAthleteID|AthleteScopedSchemasExposeUniformAthleteID|CoachToolsAbsentWhenCoachModeOff|SelectAthleteUpdatesVisibleCatalogAndListAthletes|GateCompositionTruthTable|VisibleCatalogMetadataMatchesToolsListAndSessionsAreIsolated|SelectAthleteMetadataUsesPostGateCatalog|CoachModeEndToEndRoutesSelectedDefaultAndOverrideTargets|AthleteIDRejectionMessageIsEnumerationSafe|ListAdvancedCapabilitiesVisibilityWithRealRegistry|AdvancedCapabilitiesUsesCoachFilteredCatalog|HiddenFullToolIsAbsentAndUnknown)$'
```

## Notes

- The Step 3 seam notes correctly call out the `internal/tools` -> `internal/coach` dependency and the resulting import-cycle risk.
- For Step 2, keep using the broader package gate after each mechanical move so non-coach registrar, transport, capability, and toolset behavior stays covered too.

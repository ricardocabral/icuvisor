# R008 Plan Review — Step 5: Testing & Verification

**Verdict:** APPROVE

The Step 5 plan matches the task completion criteria: rerun targeted coverage for the analyzer tier/catalog changes, run the full unit suite, build the binary, lint, and either fix failures or document clearly unrelated/pre-existing failures in `STATUS.md`.

## Execution notes

- Do not rely only on the Step 4 `make test` result. Step 5 should record fresh command results for all required gates, especially `make build` and `make lint`, which are not shown as completed in `STATUS.md` yet.
- A suitable targeted command is:
  - `go test ./cmd/gendocs ./internal/tools -run 'TestRunWritesToolsCatalogGolden|TestRegisteredToolTierMembership|TestNonCandidateAnalyzerFamilyRemainsFullToolset|TestAnalyzerCorePromotionCandidatesAreBenchmarkGated|TestCatalogIncludesAnalyzerFamilyPlacement|TestCatalogAnalyzerActivationHints|TestListAdvancedCapabilitiesOutputFromCatalog'`
- Then run the required full gates in order:
  - `make test`
  - `make build`
  - `make lint`
- If `make lint` fails because `golangci-lint` is unavailable locally, record that as an environment/tooling blocker or unrelated failure in `STATUS.md`; do not mark the lint checkbox as passing.
- If any command modifies generated files or golden data, review the diff, rerun the affected targeted test, and include the resulting files in the task changes or document why they are intentionally excluded.
- Update `STATUS.md` with exact commands and outcomes before moving to Step 6.

No plan changes are required before executing Step 5.

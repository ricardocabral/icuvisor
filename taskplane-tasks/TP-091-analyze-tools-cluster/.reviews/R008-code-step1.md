# R008 Code Review — Step 1: Design request/response contracts

**Verdict:** Changes requested

The Step 1 contract is now substantially complete: request schemas, window/baseline rules, grain semantics, lag boundaries, weighted aggregation, weekly metrics, terse/full behavior, and efforts unit fields are all specified well enough for implementation. I found one remaining metadata/source-contract inconsistency that should be resolved before Step 2 because it affects the public `_meta.source_tools` field and golden fixtures.

## Blocking finding

1. **`analyze_efforts_delta` source mapping and `_meta.source_tools` disagree.**

   The source-client mapping says `analyze_efforts_delta` uses the curve endpoints directly: `ListAthletePowerCurves`, `ListAthleteHRCurves`, and `ListAthletePaceCurves` (`STATUS.md:141`). However the R007 meta contract says efforts responses always emit `_meta.source_tools=["get_best_efforts"]` (`STATUS.md:151`). That is misleading if the analyzer fetches only the selected family from the specific curve endpoint, and it conflicts with the PRD rule that `_meta.source_tools` lists the `get_*` reads the analyzer actually ran.

   Please make the contract deterministic before implementation. Either:

   - report the family-specific read tool(s), e.g. `power -> ["get_power_curves"]`, `heart_rate -> ["get_hr_curves"]`, `pace -> ["get_pace_curves"]`; or
   - explicitly change the source mapping to say the analyzer delegates to the public `get_best_efforts` semantics and explain why it fetches/labels that aggregate surface for a single-family delta.

   The first option seems to match the current source-client mapping and avoids implying that all best-effort families were fetched when only one was used.

## Non-blocking notes

- The pace response sketch is close, but implementation should pin exact preferred-unit delta field names in tests, e.g. `absolute_delta_pace_seconds_per_km` / `absolute_delta_pace_seconds_per_mile`, so “matching baseline/delta pace fields” does not drift.
- Step 1 still shows `Status: In Progress`; that is fine while review is pending, but mark it complete once this source-tools ambiguity is resolved.

## Tests

Not run; reviewed the Step 1 contract/status diff only.

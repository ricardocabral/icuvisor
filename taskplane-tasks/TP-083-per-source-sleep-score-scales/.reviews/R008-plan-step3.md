# Plan Review — TP-083 Step 3: Fixture coverage

## Verdict: Approve

The revised Step 3 plan addresses the gaps called out in R007 and is specific enough to proceed. It names the unprotected providers/labels, uses the fixture path that `loadWellnessFixture` actually reads, keeps the unknown-source fallback intentional, and includes targeted test commands.

## Notes

- The Garmin assertion should close the existing gap by requiring readiness provenance to include both `source: garmin` and `native_scale: 0-100 Garmin Body Battery`.
- The proposed WHOOP fixture is the right divergent-source addition. As written, asserting both `sleepScore` and `readiness` provenance will protect the two WHOOP-specific labels introduced in Step 2.
- Reusing `custom_fields.json` for `source: unknown`, `native_scale: unknown` is acceptable and preferable to counting the manual-only fixture, since manual-only rows intentionally omit provenance.
- Placing the new fixture under `internal/intervals/testdata/wellness` is correct for the existing `internal/tools` fixture harness. Avoiding `internal/tools/testdata/wellness` prevents dead testdata.
- The targeted test commands are appropriate for this step. If the WHOOP fixture exposes a missing native extraction assertion in `internal/intervals`, adding a focused intervals test would also be reasonable, but it is not required by the current plan.

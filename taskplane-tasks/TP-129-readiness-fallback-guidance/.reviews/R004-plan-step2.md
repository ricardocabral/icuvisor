# Plan Review: Step 2 — Add fallback tests or prompt guidance

Verdict: Approved

The Step 2 plan addresses the required surfaces: regression coverage for `readiness: null` alongside usable supporting wellness data, prompt guidance for recovery/weekly workflows, and targeted verification with `go test ./internal/tools ./internal/prompts`.

Execution notes:
- Update the prompt source in `internal/prompts/catalog.go` and the golden files in `internal/prompts/testdata/`; editing golden testdata alone will not change rendered prompts.
- The wellness test should assert that terse output omits `readiness`, records it in `_meta.missing_fields`, still exposes available fallback/support fields such as HRV, resting HR, sleep, and `_native.garmin`, and does not synthesize a readiness score/provenance entry for absent canonical readiness.
- Keep the prompt wording cautious: state readiness is missing first, then use supporting signals as context only, not as a replacement score.

Verification:
- Ran `go test ./internal/tools ./internal/prompts` — pass on the current baseline.

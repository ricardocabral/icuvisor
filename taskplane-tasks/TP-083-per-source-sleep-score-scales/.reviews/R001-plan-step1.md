# Plan Review — TP-083 Step 1: Define source scale mapping

## Verdict: Request changes

The current Step 1 plan is too high-level to approve. It restates the task checkboxes but does not yet define the concrete source/field mapping that Step 1 is supposed to produce. Because the rest of the task depends on exact `native_scale` strings, this should be expanded before implementation proceeds.

## Findings

1. **No exact mapping is specified.**
   - Step 1 requires exact `native_scale` labels for Garmin, Whoop, Oura, and Polar, but `STATUS.md` contains only the original checklist.
   - Add a provider/field matrix with the exact strings that will be asserted in tests, and note whether each entry is based on public docs, PRD text, or an observed fixture.

2. **The plan does not address the current fallback conflict.**
   - Current code returns `"0-100 device nightly score"` for unknown `sleepScore` sources in `internal/tools/get_wellness_data.go`.
   - The task explicitly says unknown sources must be represented as `"unknown"`, not guessed. The plan should call out this behavior change and the existing test expectation that must be updated.

3. **Whoop and Garmin detection are not covered.**
   - Current native extraction only recognizes Polar, Garmin body-battery fields, and Oura sleep score in `internal/intervals/wellness.go`; Whoop is only recognized through generic source strings.
   - The Step 1 plan should identify which raw keys/source markers will be accepted for Whoop and for Garmin sleep/readiness before Step 2 changes shaping. Otherwise the mapping may exist but never be reached.

4. **Some mappings need to be field/native-field specific, not just source specific.**
   - PRD text distinguishes Polar `ans_charge` (-10..+10), `sleep_charge`, and `nightly_recharge_status` (1–6). Current code labels all Polar readiness as `nightly_recharge_status`, even when evidence might only be `ans_charge`.
   - The plan should define precedence and labels by `(field, source, native sidecar)` where the native sidecar determines the scale.

5. **Canonical response scales must remain separate.**
   - `_meta.scales.sleepScore` in `internal/response/scales.go` is the canonical response-level label. TP-083 is about `_meta.provenance.<field>.native_scale`.
   - The plan should explicitly state that Step 2 will not replace canonical scale labels with provider-native labels.

## Recommended Step 1 acceptance criteria

Before moving to Step 2, record in `STATUS.md` or the implementation notes:

- A table for `sleepScore` and `readiness` covering `polar`, `garmin`, `oura`, `whoop`, and `unknown`.
- Exact `native_scale` strings for each supported source/field/native-sidecar combination.
- Source detection precedence when multiple provider clues are present.
- Unknown fallback rule: `source: "unknown"`, `native_scale: "unknown"`.
- Which existing fixtures/tests will change, especially the unknown-source assertion and any Garmin/Whoop coverage gaps.


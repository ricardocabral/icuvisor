# Plan Review: Step 3 — Update cookbook docs

Verdict: Approved

The Step 3 plan matches the task scope: update only the readiness-check cookbook, add null-readiness/Garmin fallback guidance, keep scales explicit, and validate the docs build where available. This is the right user-facing follow-through for the Step 2 prompt changes without expanding into response-shape or generated reference changes.

Execution notes:
- Update both the recipe text and the sample answer so they explicitly say when Intervals readiness is missing/null before using HRV, resting HR, sleep, subjective fatigue/soreness/stress/feel/mood/motivation, and available `_native` provider fields as cautious support.
- Keep the Garmin wording conditional: Garmin may provide useful wellness/native fields, but do not claim it supplies an Intervals readiness score or map Body Battery/native metrics into a replacement readiness value.
- Preserve scale labels in the prompt/example, especially sleepQuality 1-4, sleepScore 0-100, feel 1-5, and any provider scale only when the field is present in icuvisor output/provenance.
- No generated `web/content/reference/tools.md` update should be needed unless the tool response docs/schema change.

Verification:
- Ran `make web-build` on the current baseline — pass, with existing Hugo deprecation warnings only.

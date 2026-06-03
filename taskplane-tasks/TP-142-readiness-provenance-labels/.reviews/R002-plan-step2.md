# Plan Review — Step 2: Add provenance and prompt regressions

**Verdict:** Approve with required clarifications.

The Step 2 plan is aligned with the task: add provider-specific readiness provenance regressions, harden recovery/weekly prompt wording and goldens, protect terse/null-stripping behavior, and run the targeted tool/prompt test packages.

## Required clarifications before/while executing Step 2

- Make the readiness provenance matrix explicit in tests: keep existing Garmin/Polar/WHOOP coverage, and add dedicated coverage for Oura readiness plus generic/unknown upstream `readiness`. The Step 1 discoveries specifically call out those gaps.
- For terse-mode/null-stripping coverage, assert that a non-null `readiness` still leaves `_meta.provenance.readiness.source`, `native_scale`, and `fetched_at` visible in the default `include_full=false` response. Do not add provenance for null/missing readiness; keep the existing missing-field behavior so prompts can say readiness is absent.
- Update both prompt surfaces in scope: `recovery_check` and `weekly_review` golden text/tests. The wording should instruct assistants to cite `_meta.provenance.readiness.source` and `native_scale` when readiness exists, and to avoid relabeling Body Battery/Oura readiness/Polar freshness/WHOOP recovery as a generic universal recovery score.

## Execution guidance

- Prefer focused fixture/inline rows and catalog/golden assertions over broad implementation changes; this task is regression hardening, not a semantic redesign.
- Be careful not to bloat prompt output enough to trip the existing terse prompt test; revise existing lines where possible rather than adding many new lines.
- Only update `internal/tools/schema_snapshot/get_wellness_data.json` if the input schema actually changes; provenance response behavior is not represented in that snapshot.
- Finish the step with `go test ./internal/tools ./internal/prompts` and record any discoveries in `STATUS.md`.

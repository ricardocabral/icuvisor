# R015 plan review — Step 4: Tests and verification

Verdict: REVISE

`STATUS.md` only marks Step 4 as in progress and keeps the original three checklist bullets. For this task, that is not enough of a test plan: the previous steps introduced new histogram math, stream orchestration, best-effort detail/profile lookups, public schema/catalog behavior, and analyzer metadata. Please amend Step 4 with the concrete test matrix and verification commands before implementing more tests.

## Blocking plan gaps

1. **Split pure engine tests from MCP/tool orchestration tests.**
   Add an explicit plan for `internal/analysis/histogram_test.go` covering bucket construction independent of MCP fixtures, and a separate plan for `internal/tools/get_activity_histogram_test.go` covering request decoding, fake clients, response shaping, and metadata. Without this split, the math contracts from Step 1 are easy to under-test or only test through brittle JSON payloads.

2. **List the zone and fixed-width fixtures/cases precisely.**
   The plan should name the cases to cover, not just “zone-based and fixed-width buckets.” At minimum include:
   - configured-zone lower-bound buckets, leading “Below …” bucket, final open-ended bucket, boundary inclusivity/exclusivity, sorted boundary/name pair preservation, and percentage/seconds rounding;
   - fixed-width fallback with exactly 10 raw-width buckets, max-value inclusion in the final bucket, deterministic one-decimal labels/meta, and the identical-min/max one-bucket case;
   - invalid/non-finite or non-positive-duration samples being skipped and reflected in `_meta.n`.

3. **Cover tool-level stream orchestration and fallback semantics.**
   Add fake-client tests proving each metric requests only the required streams with `IncludeDefaults:false` (`watts,time`; `heart_rate,time`; `distance,time`) and that no raw samples appear in default or `include_full` responses. Also cover best-effort details/profile behavior: zone selection when available, fixed-width fallback when details/profile are missing or fail non-contextually, context errors propagating, and nil optional clients not panicking.

4. **Make unit conversion coverage concrete.**
   “Test unit conversion” needs the exact matrix. Include pace emitted-unit behavior for metric vs imperial athlete preferences, at least one tool-level imperial pace response, and pure conversion tests for `MINS_KM`, `MINS_MILE`, `SECS_100M`, `SECS_500M`, plus unknown/empty pace units falling back to fixed-width rather than configured zones.

5. **Specify unavailable/missing-stream response assertions.**
   The plan should test missing required streams, length mismatches, and no valid positive-duration intervals. Assert `buckets: []`, `unavailable.reason/message`, `_meta.insufficient_sample:true`, `_meta.n`, `_meta.source_tools`, `_meta.missing_days:0`, `_meta.missing_action:"skip"`, emitted unit where applicable, and that `_meta.bucket_method` is omitted.

6. **Include schema/catalog/docs verification in the test plan.**
   Add assertions that the input schema enum exposes only `power_watts`, `heart_rate_bpm`, and `pace_seconds_per_km` while safe aliases parse server-side. Include catalog/registry tests affected by the new tool and document whether `make docs-tools` must be run. For docs, update `CHANGELOG.md`; do not hand-edit generated web tool artifacts except through the generator.

7. **Reconcile Step 4 vs Step 5 quality gates.**
   Step 4 currently says “Run full quality gate,” while Step 5 separately owns `make test`, `make build`, and `make lint`. Amend the plan with the exact commands for this step (for example targeted `go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./cmd/gendocs` plus `make docs-tools`/diff checks), and state whether the full gates are run now or deferred to Step 5. If they are run in Step 4, record results here and avoid leaving Step 5 ambiguous.

## Suggested minimum command set

- `go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./cmd/gendocs`
- `make docs-tools` followed by a focused diff check for generated catalog artifacts
- Either run `make test`, `make build`, and `make lint` in this step, or explicitly defer them to Step 5 and update the Step 4 wording accordingly

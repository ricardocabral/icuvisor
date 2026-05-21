# R016 plan review — Step 4: Tests and verification

Verdict: APPROVE

The amended Step 4 plan in `STATUS.md` now provides a concrete and appropriately scoped test/verification plan for `get_activity_histogram`. It resolves the R015 blocking gaps by separating pure histogram math tests from MCP/tool orchestration tests, naming the required zone/fixed-width/unit/unavailable/schema/catalog cases, and reconciling Step 4 targeted verification with Step 5 full-suite gates.

## What looks good

- The planned `internal/analysis/histogram_test.go` coverage directly exercises the Step 1 math contract: configured lower-bound buckets, leading and final open-ended buckets, boundary inclusivity, sorted boundary/name pairs, fixed-width edge generation, sample skipping, and rounding.
- The planned `internal/tools/get_activity_histogram_test.go` coverage addresses tool behavior that should not be hidden behind engine tests: strict decoding/schema behavior, fake stream/detail/profile clients, source-tool/method/n metadata, response shaping, no raw samples, unavailable payloads, and best-effort fallback behavior.
- The pace/unit matrix is concrete enough to prevent ambiguous implementation: emitted metric vs imperial units, all documented pace-zone unit conversions, and unknown/empty pace-unit fallback to fixed-width.
- Missing-stream and insufficient-sample assertions include the important public response contract: empty buckets, unavailable reason/message, analyzer meta, omitted `bucket_method`, and no raw samples.
- Schema/catalog/doc verification is included, including the histogram-only enum and generated catalog artifacts.
- The command plan is now explicit: targeted package tests plus generated-doc updates in Step 4, with `make test`, `make build`, and `make lint` deferred to Step 5.

## Non-blocking implementation notes

- When executing the plan, prefer `make docs-tools` as the canonical generated-doc path, then inspect the focused diff for `web/data/tools.json` and any gendocs golden files that the repo actually expects.
- Keep the Step 4 checkbox wording in mind: since full `make test`/`build`/`lint` are deferred, record targeted command results clearly in `STATUS.md` so Step 5 remains the single full-gate checkpoint.
- Ensure `CHANGELOG.md` is updated before closing Step 4, since this is user-visible tool behavior.

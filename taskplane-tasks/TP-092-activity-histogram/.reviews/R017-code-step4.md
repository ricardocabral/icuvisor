# R017 code review — Step 4: Tests and verification

Verdict: REVISE

## Findings

1. **[P2] Step 4 does not yet exercise the approved strict schema/decoding coverage.**  
   `internal/tools/get_activity_histogram_test.go` covers successful handler calls plus unavailable paths, but there is no handler/schema test for the public input contract from the Step 4 plan: unknown fields must be rejected by strict decoding, missing `activity_id`/`metric` should return the terse user error, invalid non-histogram metrics should be rejected, and the registered tool's `metric` schema should enumerate only the three canonical histogram values. The only schema assertion is the lower-level helper test in `internal/analysis/histogram_test.go`, so a regression in `activityHistogramInputSchema()` or handler wiring could still pass Step 4.

2. **[P2] Fixed-width metadata/raw-edge coverage is too weak to protect the Step 1 contract.**  
   In `internal/analysis/histogram_test.go:39-56`, the fixed-width case uses `min=0`, `max=10`, and `width=1`, then only checks `BucketCount` and `Width`. The task contract says `_meta.fixed_width` reports the raw `min`, `max`, and `width` used for buckets, with no nice-number rounding. This test would still pass if the implementation rounded those metadata fields or omitted incorrect `Min`/`Max` values. Please add a fractional min/max case that asserts `FixedWidth.Min`, `FixedWidth.Max`, and `FixedWidth.Width` against the raw bucket-edge values (and preferably through the tool response `_meta.fixed_width` as well).

3. **[P3] `STATUS.md` claims the full quality gate is complete while Step 5 remains not started.**  
   `taskplane-tasks/TP-092-activity-histogram/STATUS.md:70-72` marks “Run full quality gate and update docs/CHANGELOG” as done, but Step 5 still has `make test`, `make build`, and `make lint` unchecked, and the execution log does not record the targeted command outcome. The note at lines 177-179 says Step 4 should mean the targeted package tests only; please either reword that checkbox/status entry or log the targeted command explicitly so the status does not imply the full suite/build/lint gates have already run.

## Verification

- Ran: `go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./cmd/gendocs` — passed.

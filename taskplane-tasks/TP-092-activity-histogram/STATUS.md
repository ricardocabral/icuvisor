# TP-092: `get_activity_histogram` single-activity histogram tool — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 19
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Define histogram contract
**Status:** ✅ Complete

- [x] Limit metrics to power/HR/pace through the closed metric enum.
- [x] Define bucket output fields: label/range, seconds, percentage, unit.
- [x] Define `_meta.bucket_method` values for configured zones vs fixed-width fallback.
- [x] Resolve R001 metric enum/schema ambiguity with exact canonical values.
- [x] Resolve R001 row shape/range/percentage/ordering contract details.
- [x] Resolve R001 analyzer meta, zone-source, fixed-width, pace, and unavailable-response contract details.
- [x] Resolve R002 zone-boundary-to-bucket mapping and label fallback rules.
- [x] Resolve R002 fixed-width deterministic bucket edges and identical-value behavior.
- [x] Resolve R002 duration weighting, required stream keys, and pace response identity.
- [x] Resolve R004 power stream key contract mismatch (`watts` + `time`).
- [x] Resolve R005 configured-zone boundary/name pair preservation.
- [x] Resolve R006 pace-zone unit conversion table and unknown-unit fallback.
- [x] Resolve R007 pace stream base units/formula and sport-setting selection precedence.

---

### Step 2: Implement stream-backed histogram
**Status:** ✅ Complete

- [x] Add histogram analysis engine and metric subset helpers matching Step 1 contract.
- [x] Fetch only required streams for one activity.
- [x] Use athlete configured zones where available; fall back to fixed-width buckets with documented edges.
- [x] Return terse per-bucket summary only; no raw samples.
- [x] Resolve R009 metric subset/schema helper and package-boundary plan.
- [x] Resolve R009 stream request matrix and non-fatal profile/details lookup plan.
- [x] Resolve R009 unavailable/meta construction plan.

---

### Step 3: Tool registration and activation hint
**Status:** ✅ Complete

- [x] Register in `full` with description leading on single-activity distribution prompts.
- [x] Explicitly say not to pull `get_activity_streams` and bin manually.
- [x] Add source_tools/method/n meta.
- [x] Wire catalog grouping/tool constants so registration and generated docs can discover the tool.
- [x] Resolve R013 stale generated tool catalog artifacts.

---

### Step 4: Tests and verification
**Status:** ✅ Complete

- [x] Add fixtures for zone-based and fixed-width buckets.
- [x] Test unit conversion and missing stream handling.
- [x] Run full quality gate and update docs/CHANGELOG.
- [x] Resolve R015 pure engine vs tool orchestration test matrix.
- [x] Resolve R015 concrete zone/fixed-width/unit/unavailable/schema/catalog coverage plan.
- [x] Resolve R015 Step 4 targeted commands vs Step 5 full quality gates plan.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | Plan | 1 | APPROVE | inline |
| R004 | Code | 1 | REVISE | `.reviews/R004-code-step1.md` |
| R005 | Code | 1 | REVISE | `.reviews/R005-code-step1.md` |
| R006 | Code | 1 | REVISE | `.reviews/R006-code-step1.md` |
| R007 | Code | 1 | REVISE | `.reviews/R007-code-step1.md` |
| R008 | Code | 1 | APPROVE | inline |
| R009 | Plan | 2 | REVISE | `.reviews/R009-plan-step2.md` |
| R010 | Plan | 2 | APPROVE | inline |
| R011 | Code | 2 | APPROVE | inline |
| R012 | Plan | 3 | APPROVE | inline |
| R013 | Code | 3 | REVISE | `.reviews/R013-code-step3.md` |
| R014 | Code | 3 | APPROVE | inline |
| R015 | Plan | 4 | REVISE | `.reviews/R015-plan-step4.md` |
| R016 | Plan | 4 | APPROVE | inline |
| R017 | Code | 4 | APPROVE | inline |
| R018 | Plan | 5 | APPROVE | inline |
| R019 | Code | 5 | APPROVE | inline |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| No out-of-scope follow-up discoveries; generated catalog and adversarial static catalog tests needed expected updates for the new registered read tool. | Addressed in task scope. | `cmd/gendocs/testdata/tools.golden.json`, `web/data/tools.json`, `internal/safety/adversarial_test.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 17:20 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 17:20 | Step 0 started | Preflight |
| 2026-05-20 18:34 | Worker iter 1 | done in 4425s, tools: 267 |
| 2026-05-20 18:34 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

### Step 1 histogram contract

**Input schema / metric enum:** `get_activity_histogram` accepts `activity_id` (required string), `metric` (required closed enum), and `include_full` (optional bool; default false and never includes raw samples). The metric schema enumerates canonical `analysis_metric` values only. Step 1 will extend `internal/analysis` with missing stream-grain enum values `power_watts` and `heart_rate_bpm`, and will define a histogram-specific allowed subset that also includes the existing canonical value `pace_seconds_per_km`; safe aliases such as `power`, `hr`, `heart_rate`, and `pace` may parse server-side but schemas expose only canonical values. These are the only accepted histogram metrics; existing activity aggregate metrics such as `average_power_watts` and `average_heart_rate_bpm` are rejected for this tool.

**Bucket row shape:** response body is `{ "activity_id", "metric", "buckets", "_meta" }`. Each bucket is `{ "label": string, "lower": number|null, "upper": number|null, "lower_inclusive": bool, "upper_exclusive": bool, "seconds": number, "percentage": number, "unit": string }`. The first bucket may have `lower:null` for `(-inf, upper)`, the last may have `upper:null` for `[lower, +inf)`. Interior buckets are `[lower, upper)`. `seconds` is the summed duration in seconds represented by valid stream samples in the bucket, rounded to 0.1s. `percentage` is `seconds / bucketed_seconds * 100`, rounded to 0.1 percentage points; `bucketed_seconds` excludes missing/invalid/non-positive-duration samples. Bucket ordering is ascending metric value for power and heart rate. Pace ordering is fast-to-slow (ascending seconds-per-distance; faster/lower values first) so ranges still sort ascending by numeric value.

**Meta, zones, fallback, and unavailable semantics:** `_meta.bucket_method` is one of `configured_zones` or `fixed_width` when buckets can be constructed, and is omitted on unavailable/insufficient-sample payloads with no buckets. Analyzer metadata from TP-089 is always present: `_meta.method:"activity_stream_histogram"`, `_meta.source_tools` includes `get_activity_streams` plus `get_activity_details` and `get_athlete_profile` when used for sport/zone selection, `_meta.n` is the count of valid stream intervals contributing to buckets, `_meta.missing_days:0`, `_meta.missing_action:"skip"`, and `_meta.insufficient_sample` is true when `n < 1`. Zone selection fetches activity details to derive the activity sport/type, normalizes candidate strings with trim + case-fold, considers `Activity.Type` first and `Activity.SubType` second, and matches sport settings deterministically: exact `SportSettings.Type` matches for the first activity candidate win before any `SportSettings.Types` matches; then exact `Type` matches for the second candidate win before `Types` matches; ties keep profile order. If no deterministic match exists, use fixed-width fallback. The selected setting uses metric-specific `PowerZones`/`HRZones`/`PaceZones` plus zone names when at least one boundary exists. `_meta.zone_source` reports the selected sport setting (`sport`, `sport_setting_id`, `metric`, and `pace_units` for pace). If details/profile/zone settings are missing or have no boundaries for the metric, the tool falls back to fixed-width buckets; if profile fetch fails entirely, pace output defaults to `seconds_per_km`. `_meta.fixed_width` reports `min`, `max`, `bucket_count`, `width`, and `unit` for fallback edges.

Power stream samples use watts (`W`). Heart-rate samples use bpm (`bpm`). Pace derives seconds per distance from adjacent distance/time stream samples and emits `seconds_per_km` for metric/preferred units, or `seconds_per_mile` when the athlete's preferred units are imperial. Configured pace-zone values are interpreted using the following explicit `SportSettings.PaceUnits` table before bucketing: `MINS_KM` means seconds per km (`seconds_per_km=value`, `seconds_per_mile=value*1609.344/1000`); `MINS_MILE` means seconds per mile (`seconds_per_mile=value`, `seconds_per_km=value*1000/1609.344`); `SECS_100M` means seconds per 100m (`seconds_per_km=value*10`, `seconds_per_mile=value*16.09344`); `SECS_500M` means seconds per 500m (`seconds_per_km=value*2`, `seconds_per_mile=value*1609.344/500`). If pace zones are present but `PaceUnits` is empty or unknown, ignore the configured pace zones and use fixed-width fallback rather than assuming emitted units. Missing required streams or no valid intervals returns a terse payload with `buckets: []`, `unavailable: { reason, message }`, and `_meta.insufficient_sample:true` rather than raw samples.

**Configured-zone bucket construction:** zone boundary arrays are treated as ordered lower bounds for named zones. Boundaries and names are treated as pairs before filtering, unit conversion, and sorting: drop the paired name when dropping a non-finite boundary, sort the remaining pairs by emitted boundary value ascending, and ignore extra names without boundaries. Buckets are `[boundary[i], boundary[i+1])` and the last bucket is `[last_boundary, +inf)`. If the first boundary is greater than 0, emit a leading `(-inf, first_boundary)` bucket labeled `Below <first zone label>`; otherwise no leading bucket. Zone labels use the sorted boundary/name pair's name for the bucket whose lower bound is that boundary when present and non-empty; otherwise `Zone <i+1>`. If names are shorter than boundaries, fallback labels are used for the missing paired names; extra names are ignored. `_meta.zone_source` includes the emitted boundary values and `boundary_unit` after unit conversion, so pace-zone conversions remain auditable.

**Fixed-width fallback edges:** fallback uses raw stream value min/max after unit conversion, with no nice-number rounding. If `min == max`, emit one bucket `[min, +inf)` labeled `<min> <unit>` with `bucket_count:1`, `width:0`, and 100% of bucketed seconds. Otherwise emit exactly 10 buckets. `width = (max - min) / 10`; bucket `i` is `[min + i*width, min + (i+1)*width)` for `i < 9`, and `[min + 9*width, +inf)` for the final bucket so values equal to max are included. Fixed-width labels are deterministic: `<lower>-<upper> <unit>` for finite ranges and `>= <lower> <unit>` for the final open-ended bucket, with numeric label values formatted to one decimal place. `_meta.fixed_width` reports the raw `min`, `max`, `bucket_count`, `width`, and `unit` used to construct those ranges.

**Required streams and duration weighting:** power requires canonical stream keys `watts` and `time`; heart rate requires `heart_rate` and `time`; pace requires `distance` and `time`. Power/heart-rate intervals use adjacent time samples: value sample `i` is weighted by `time[i+1]-time[i]` when that delta is finite and positive. Pace interval `i` treats `time` samples as elapsed seconds and `distance` samples as meters: `dt_seconds := time[i+1]-time[i]`, `dd_meters := distance[i+1]-distance[i]`, `seconds_per_km = dt_seconds / (dd_meters / 1000)`, and `seconds_per_mile = dt_seconds / (dd_meters / 1609.344)`; intervals with non-positive time or distance deltas are skipped. If the value stream exists but required timing/distance streams are absent or length-mismatched, return the structured unavailable/insufficient-sample payload. For pace responses, the top-level `metric` remains the requested canonical `pace_seconds_per_km`; output `unit` and `_meta.emitted_unit` identify `seconds_per_km` or `seconds_per_mile` after preferred-unit conversion.

R004 suggestions: avoid duplicate `pace_seconds_per_km` metric catalog entries by using a histogram-specific subset that includes existing pace metric plus new stream-grain metrics; keep execution-log hygiene in future STATUS edits.

R005 suggestions: define unavailable `_meta.bucket_method` as omitted when no buckets can be constructed; keep execution-log hygiene in future STATUS edits.

R006 suggestions: add deterministic fixed-width label text; keep execution-log hygiene in future STATUS edits.

R007 suggestions: define profile-fetch failure pace-unit fallback and fixed-label numeric formatting; keep execution-log hygiene in future STATUS edits.

### Step 4 test plan

**Engine vs tool split:** add `internal/analysis/histogram_test.go` for pure math: configured-zone bucket construction, lower-bound ranges, leading `Below` bucket, final open-ended bucket, boundary inclusivity, sorted boundary/name pair preservation, fixed-width 10 raw buckets, max-value final inclusion, identical-value one-bucket fallback, sample skipping, seconds/percentage rounding, and pace-zone conversion. Add `internal/tools/get_activity_histogram_test.go` for MCP orchestration: strict request decoding/schema, fake stream/detail/profile clients, source-tool/method/n metadata, response shaping, no raw samples, unavailable payloads, and best-effort zone/profile fallback behavior.

**Concrete coverage:** zone tests cover configured lower bounds, shorter names fallback, extra names ignored, pair sorting after conversion/filtering, inclusive lower/exclusive upper assignment, and `_meta.bucket_method:"configured_zones"`. Fixed-width tests cover deterministic 10 buckets, `>=` final label with one-decimal formatting, raw min/max/width meta, max-value inclusion, and identical-min/max bucket. Unit tests cover emitted `seconds_per_km` vs `seconds_per_mile`, conversion factors for `MINS_KM`, `MINS_MILE`, `SECS_100M`, `SECS_500M`, and unknown/empty pace units selecting fixed-width fallback. Unavailable tests cover missing streams, length mismatches, and no valid intervals with `buckets:[]`, reason/message, omitted bucket_method, `insufficient_sample:true`, `n`, source tools, missing days/action, and emitted unit. Schema/catalog tests assert only the histogram metric subset is enumerated, aliases parse, the registered tool is full-tier activities group, and generated `web/data/tools.json` plus `cmd/gendocs/testdata/tools.golden.json` are updated by `make docs-tools`/`go run ./cmd/gendocs`.

**Step 4 commands:** run targeted `go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./cmd/gendocs` after adding tests and generated catalog updates. Defer full `make test`, `make build`, and `make lint` to Step 5; Step 4's existing "Run full quality gate" checkbox will mean the targeted quality gate plus docs/CHANGELOG updates, while Step 5 remains the full-suite gate.

### Step 2 implementation plan

**Metric subset and package boundary:** add histogram-specific helpers in `internal/analysis/histogram.go`: `HistogramMetric` constants, `ParseHistogramMetric`, `HistogramMetricValues`, and `HistogramMetricSchemaProperty`. The helper accepts aliases (`power`, `watts`, `hr`, `heart_rate`, `pace`) but returns only canonical `power_watts`, `heart_rate_bpm`, or `pace_seconds_per_km`; the tool schema uses this subset, never the full `MetricValues()`. Pure math stays in `internal/analysis`: stream interval construction, zone/fixed bucket construction, pace-zone unit conversion, bucket assignment, rounding, labels, percentages, and meta fragments. `internal/tools/get_activity_histogram.go` handles MCP decoding, stream/detail/profile fetching, unavailable payloads, and response shaping; `analysis` does not import `tools`.

**Stream requests and orchestration:** the handler calls `ActivityStreamsClient.GetActivityStreams` directly with `IncludeDefaults:false` and exact upstream/canonical types: power requests `[]string{"watts","time"}`, heart-rate requests `[]string{"heart_rate","time"}`, and pace requests `[]string{"distance","time"}`. Returned rows are read through `streams.CanonicalKey(firstNonEmpty(row.Type,row.Name))`; the public `get_activity_streams` handler is not called. Details/profile lookups are best-effort for zone selection: fetch activity details to get `Type`/`SubType`, fetch athlete profile for sport settings and preferred units, but any non-context error from either lookup is recorded in meta and falls back to fixed-width buckets. Profile fetch failure does not make pace unavailable; it emits `seconds_per_km`. Only missing required stream data or no valid weighted intervals returns unavailable/insufficient sample.

**Unavailable/meta construction:** tool responses have no raw samples in terse or full mode for this tool. Success returns `activity_id`, canonical `metric`, `buckets`, and `_meta` with analyzer fields plus `bucket_method`, `emitted_unit`, optional `zone_source`, optional `fixed_width`, and lookup warning strings when best-effort detail/profile fetches fail. Missing required streams, length mismatches, or zero valid intervals return `activity_id`, `metric`, `buckets: []`, `unavailable:{reason,message}`, and `_meta` with `method:"activity_stream_histogram"`, `source_tools:["get_activity_streams"]` plus attempted lookups, `n:0`, `missing_days:0`, `missing_action:"skip"`, `insufficient_sample:true`, and no `bucket_method`. The engine preserves Step 1 bucket semantics exactly: zone lower bounds, final open-ended bucket, fixed 10 raw-width fallback except identical values, one-decimal duration/percentage rounding, and percentage denominator as bucketed seconds.
| 2026-05-20 17:24 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 17:28 | Review R002 | plan Step 1: REVISE |
| 2026-05-20 17:32 | Review R003 | plan Step 1: APPROVE |
| 2026-05-20 17:34 | Review R004 | code Step 1: UNKNOWN |
| 2026-05-20 17:37 | Review R005 | code Step 1: UNKNOWN |
| 2026-05-20 17:40 | Review R006 | code Step 1: UNKNOWN |
| 2026-05-20 17:44 | Review R007 | code Step 1: UNKNOWN |
| 2026-05-20 17:48 | Review R008 | code Step 1: APPROVE |
| 2026-05-20 17:50 | Review R009 | plan Step 2: REVISE |
| 2026-05-20 17:52 | Review R010 | plan Step 2: APPROVE |
| 2026-05-20 18:02 | Review R011 | code Step 2: APPROVE |
| 2026-05-20 18:05 | Review R012 | plan Step 3: APPROVE |
| 2026-05-20 18:09 | Review R013 | code Step 3: REVISE |
| 2026-05-20 18:11 | Review R014 | code Step 3: APPROVE |
| 2026-05-20 18:14 | Review R015 | plan Step 4: REVISE |
| 2026-05-20 18:16 | Review R016 | plan Step 4: APPROVE |
| 2026-05-20 18:24 | Review R017 | code Step 4: APPROVE |
| 2026-05-20 18:26 | Review R018 | plan Step 5: APPROVE |
| 2026-05-20 18:31 | Review R019 | code Step 5: APPROVE |

# TP-017: Raise Go test coverage to 80% — Status

**Current Step:** Step 5: Documentation & Delivery
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-13
**Review Level:** 1
**Review Counter:** 1
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code
> changes. Workers expand steps when runtime discoveries warrant it — aim for
> 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm baseline with `go test ./... -coverprofile=coverage.out -covermode=atomic` and `go tool cover -func=coverage.out | tail -1`

---

### Step 1: Pick high-value coverage targets

**Status:** ✅ Complete

> ⚠️ Hydrate: Expand checkboxes when entering this step based on the current per-function coverage report.

- [x] Generate per-function coverage report and select high-value targets
- [x] Prioritize behavior-rich low-coverage functions over trivial command paths
- [x] Record selected target files/functions and rationale here

Selected targets:

| File | Functions | Rationale |
| ---- | --------- | --------- |
| `internal/intervals/activity_streams.go` | `ActivityStream.UnmarshalJSON`, `Client.GetActivityStreams` | 0% coverage; small API client surface with raw-field preservation, required input validation, query parameters, and HTTP error wrapping. |
| `internal/intervals/wellness.go` | `Wellness.UnmarshalJSON`, `ListWellness`, `extractWellnessNative`, `nativeSleepScoreSource`, `dedupeStrings` | 0% coverage; behavior-rich parsing/provenance logic with duplicate claimed-key handling and request construction. |
| `internal/tools/get_extended_metrics.go` | `getExtendedMetricsHandler`, `optionalIntervals`, `optionalPowerVsHR`, `stravaUnavailableExtendedMetricsResponse`, `shapeExtendedMetrics`, helper numeric/string branches | Low/0% branch coverage; tests can cover Strava-blocked responses, include_full raw payloads, optional not-found/unauthorized sources, and non-optional errors without production changes. |
| `internal/toolchecks/*.go` | catalog/snapshot guards | Fallback only if Step 2 + extended-metrics tests leave total coverage below 80%; deterministic and cheap but lower priority than product behavior. |

---

### Step 2: Add focused intervals-package tests

**Status:** ✅ Complete

> ⚠️ Hydrate: Expand with specific activity-stream and wellness scenarios after reading existing intervals test helpers.

- [x] Add activity-stream tests for JSON decoding, raw preservation, query parameters, required IDs, and errors
- [x] Create `internal/intervals/wellness_test.go` with native-provider and `ListWellness` tests
- [x] Use `httptest`/existing client helpers only; no network access in intervals tests
- [x] Targeted intervals tests pass

---

### Step 3: Add focused tool/toolcheck tests until the target is met

**Status:** ✅ Complete

> ⚠️ Hydrate: Expand based on remaining coverage gap after Step 2.

- [x] Add extended-metrics tests for uncovered behavior-rich branches
- [x] Re-run coverage after extended-metrics tests and skip toolcheck tests if total coverage is already >= 80.0%
- [x] Add or extend toolcheck tests only if additional coverage is needed
- [x] Targeted tests for touched packages pass

---

### Step 4: Testing & Verification

**Status:** ✅ Complete

- [x] Coverage gate passes: `go test -race -count=1 -coverprofile=coverage.txt -covermode=atomic ./...`
- [x] Total statement coverage is >= 80.0%
- [x] `make lint` passes
- [x] `make build` passes
- [x] Local coverage artifacts removed or intentionally left untracked/ignored

---

### Step 5: Documentation & Delivery

**Status:** 🟨 In Progress

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final coverage percentage recorded

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |
| 1 | plan | 1 | APPROVE | |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |
| Step 2 intervals tests alone raised full-suite coverage to 79.7%; only a small extended-metrics increment was needed after that. | Added focused extended-metrics branch tests; skipped toolcheck changes. | `internal/intervals/*_test.go`, `internal/tools/get_extended_metrics_test.go` |
| No production-code behavior changes were required, so `CHANGELOG.md` and `CONTRIBUTING.md` did not need edits. | Documented in delivery status only. | Step 5 |

---

## Execution Log

| Timestamp  | Action      | Outcome                         |
| ---------- | ----------- | ------------------------------- |
| 2026-05-13 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-13 12:21 | Task started | Runtime V2 lane-runner execution |
| 2026-05-13 12:21 | Step 0 started | Preflight |

---

## Blockers

_None_

---

## Notes

Current observed baseline before staging this task: 76.9% total statement coverage from `go test ./... -coverprofile=coverage.out -covermode=atomic`.

Step 1 per-function report (2026-05-13): overall baseline remains 76.9%. Highest-value behavior targets are the 0%-covered `internal/intervals/activity_streams.go` (`ActivityStream.UnmarshalJSON`, `Client.GetActivityStreams`) and `internal/intervals/wellness.go` (`Wellness.UnmarshalJSON`, `ListWellness`, native-provider helpers), followed by uncovered/low-covered `internal/tools/get_extended_metrics.go` optional-source and Strava-unavailable branches. `internal/toolchecks` remains a fallback if those tests do not reach 80%.

Step 3 coverage gap after intervals tests: full-suite coverage is 79.7%, so a small number of extended-metrics branch tests should reach the target. Remaining under-covered extended-metrics functions are `stravaUnavailableExtendedMetricsResponse` (0%), optional source helpers (37.5%), handler source-error branches (60.7%), and raw number/string helper edge cases.

Final verification coverage (2026-05-13): `go test -race -count=1 -coverprofile=coverage.txt -covermode=atomic ./...` plus `go tool cover -func=coverage.txt | tail -1` reported total statement coverage of 80.8%.
| 2026-05-13 12:27 | Review R001 | plan Step 1: APPROVE |

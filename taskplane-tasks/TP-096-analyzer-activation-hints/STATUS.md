# TP-096: Activation-hint pass on analyzer descriptions — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 1
**Review Counter:** 7
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Audit analyzer descriptions
**Status:** ✅ Complete

- [x] List every analyzer-family tool and current first sentence.
- [x] Identify missing activation hints and missing do-not-roll-your-own language.
- [x] Check for confusable wording across similar tools.

---

### Step 2: Update descriptions and checks
**Status:** ✅ Complete

- [x] Rewrite descriptions so each starts with a concrete prompt shape.
- [x] Add explicit raw-row/stream avoidance language where applicable.
- [x] Add or update catalog tests/checks to keep future analyzer descriptions compliant.

---

### Step 3: Regenerate docs/catalog artifacts
**Status:** ✅ Complete

- [x] Resolve or block on the current `compute_baseline.go` build error before generation; do not hand-edit generated catalog artifacts around a failing generator.
- [x] Run the catalog/docs generation command if descriptions feed generated docs, refreshing both `web/data/tools.json` and `cmd/gendocs/testdata/tools.golden.json` from generator output.
- [x] Review path-scoped generated diffs for analyzer-family summaries and rendered reference docs (or document Hugo unavailability and JSON/static-wrapper fallback).
- [x] Update CHANGELOG.md under `[Unreleased]` for user-visible analyzer activation-hint wording.

---

### Step 4: Verify
**Status:** ✅ Complete

- [x] Run explicit targeted catalog/toolcheck/gendocs tests: `go test ./internal/tools -run 'TestCatalog' -count=1`, `go test ./internal/toolchecks -run 'TestCheckConfusableCatalog|TestFirstDescriptionSentence|TestGenerateToolCatalogUsesCallerContext' -count=1`, and `go test ./cmd/gendocs -run TestRunWritesToolsCatalogGolden -count=1`.
- [x] Run targeted production analyzer tests for touched code: `go test ./internal/tools -run 'TestComputeBaseline|TestComputeActivitySegmentStats|TestGetActivityHistogram|TestGetFitnessProjection' -count=1` and fix any touched-code failures.
- [x] Run full quality gate concretely with `make check`; fix failures or document exact pre-existing unrelated failures before proceeding.

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

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| `get_fitness_projection`, `get_activity_histogram`, and `compute_activity_segment_stats` do not lead with the required concrete prompt-shape wording; `get_fitness_projection` also lacks explicit raw-row avoidance language, while `get_activity_histogram` puts stream-avoidance only in the second sentence and `compute_activity_segment_stats` is the raw-stream exception that should state why it is allowed. | Fix in Step 2 descriptions and guard with toolcheck coverage. | `internal/tools/get_fitness_projection.go`, `internal/tools/get_activity_histogram.go`, `internal/tools/compute_activity_segment_stats.go` |
| Existing `analyze_*`, `compute_zone_time`, `compute_load_balance`, `compute_baseline`, and `compute_compliance_rate` descriptions already contain do-not-roll-your-own language, but their first-sentence phrasing is inconsistent (`Use when...` vs `Use this when...`). | Normalized in Step 2 for future catalog consistency. | `internal/tools/analyze_*.go`, `internal/tools/compute_*.go` |
| Step 3 plan review R003 confirmed generator/doc refresh is blocked until `internal/tools` builds; generated catalog outputs must not be manually edited as a workaround. | Added Step 3 revision checklist and fixed the in-repo build blocker before running generation. | `.reviews/R003-plan-step3.md`, `internal/tools/compute_baseline.go` |
| Step 4 plan review R005 found current compute-baseline targeted tests fail after the Step 3 build-blocker repair, with wellness baseline rows producing zero baseline samples. | Fixed in Step 4 by preserving raw field fallbacks around shared analyzer source helpers; targeted production analyzer tests pass. | `.reviews/R005-plan-step4.md`, `internal/tools/compute_baseline.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 21:54 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 21:54 | Step 0 started | Preflight |
| 2026-05-20 22:23 | Worker iter 1 | done in 1743s, tools: 196 |
| 2026-05-20 22:23 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

Step 1 audit first sentences (2026-05-20):
- `analyze_trend`: "Use when the prompt asks whether an analysis metric is trending up/down or changing versus baseline; do not fetch rows and reduce them in chat."
- `analyze_distribution`: "Use when the prompt asks for an analysis metric's distribution, histogram, quantiles, or outliers; do not fetch rows and reduce them in chat."
- `analyze_correlation`: "Use when the prompt asks whether two analysis metrics are correlated or lagged together; do not fetch rows and reduce them in chat."
- `analyze_efforts_delta`: "Use when the prompt asks whether best-effort power, heart-rate, or pace buckets changed versus baseline; do not fetch rows and reduce them in chat."
- `get_fitness_projection`: "Project CTL, ATL, and TSB forward from a current fitness row using deterministic load assumptions: weekly ramp %, optional recovery-week cadence, horizon, and optional explicit planned daily loads."
- `get_activity_histogram`: "Summarize a single activity's power, heart-rate, or pace distribution into terse time-in-bucket histogram rows."
- `compute_activity_segment_stats`: "Compute deterministic stats over one activity segment from canonical raw streams as the analyzer-family raw-stream exception."
- `compute_zone_time`: "Use this when the user asks for time in power, heart-rate, or pace zones over a date window."
- `compute_load_balance`: "Use this when the user asks whether training distribution is polarized, pyramidal, threshold-heavy, or balanced across low/moderate/high intensity."
- `compute_baseline`: "Use this when the user asks whether a metric is high, low, suppressed, elevated, or unusual versus a baseline window."
- `compute_compliance_rate`: "Use this when the user asks how well completed activities matched scheduled workouts, targets, sport, or event type."

Step 1 verification note: `go test ./internal/tools -run TestCatalog -count=1` and `go test ./internal/toolchecks -count=1` initially failed before tests ran due to pre-existing TP-093 compute_baseline duplicate helper/signature build errors (`summaryMetricValue`, `wellnessMetricValue`, `activityMetricValue` redeclared and wrong call signatures). Step 3 resolved the local build blocker by reusing shared analyzer source helpers with `response.UnitSystem` instead of redeclaring duplicate helper functions.

Step 1 confusable wording check: static token-overlap audit across the 11 analyzer-family first sentences found no pair at or above 0.35 Jaccard after removing common activation words. The main ambiguity risk is not tool-to-tool confusion, but missing/late activation language in `get_fitness_projection`, `get_activity_histogram`, and `compute_activity_segment_stats`.

Step 3 generated-docs review: `make docs-tools` updated `web/data/tools.json`, then the same generator output was copied to `cmd/gendocs/testdata/tools.golden.json`. Path-scoped diff review showed only intended analyzer-family summary changes across all 11 tools. `make web-build` succeeded with existing Hugo deprecation warnings, and rendered `web/public/reference/tools/index.html` now includes analyzer, histogram, and fitness-projection activation-hint summaries; `web/layouts/partials/tool-catalog.html` was fixed to include the `analyzers` group so those generated rows render.

Step 4 verification: targeted catalog/toolcheck/gendocs tests passed; targeted production analyzer tests initially exposed compute-baseline raw-field fallback regression from the build-blocker repair, then passed after preserving raw fallbacks. `make check` passed (go vet, golangci-lint, and `go test -race -count=1 ./...`).

Step 5 verification: targeted analyzer/catalog/toolcheck/gendocs tests passed; `make test`, `make build`, and `make lint` all passed. No remaining failures to document.

Step 6 docs: Must Update docs were modified: `CHANGELOG.md` records the user-visible analyzer activation-hint wording under `[Unreleased]`, and `STATUS.md` contains audit/verification notes. Check If Affected docs reviewed: `README.md` only points to the website catalog and `make docs-tools` with no tool-description text to update; `web/content/reference/tools.md` remains the generated shortcode wrapper and already includes fitness-projection context; `web/data/tools.json` was regenerated; `web/layouts/partials/tool-catalog.html` was updated so analyzer group rows render; PRD analyzer section already contains the product-level do-not-roll-your-own contract and no product behavior change was made.
| 2026-05-20 21:57 | Review R001 | plan Step 1: APPROVE |
| 2026-05-20 22:01 | Review R002 | plan Step 2: APPROVE |
| 2026-05-20 22:06 | Review R003 | plan Step 3: REVISE |
| 2026-05-20 22:07 | Review R004 | plan Step 3: APPROVE |
| 2026-05-20 22:14 | Review R005 | plan Step 4: REVISE |
| 2026-05-20 22:15 | Review R006 | plan Step 4: APPROVE |
| 2026-05-20 22:20 | Review R007 | plan Step 5: APPROVE |

# TP-083: Per-source sleep-score scale labels in wellness provenance — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 15
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

### Step 1: Define source scale mapping
**Status:** ✅ Complete

- [x] Review existing provenance fields and native sidecars.
- [x] Define exact `native_scale` labels for at least Garmin, Whoop, Oura, and Polar based on public docs/observed fixtures.
- [x] Represent unknown sources as `unknown`, not a guessed scale.
- [x] R002: Update stale unknown-source wellness test expectation so `go test ./internal/tools` is not red after Step 1.
- [x] R002: Make Polar readiness `native_scale` sidecar-specific, distinguishing `ans_charge` from `nightly_recharge_status` with defined precedence.
- [x] R002: Record the exact field/source/sidecar native-scale matrix in STATUS.md for Step 1.

---

### Step 2: Apply provenance labels
**Status:** ✅ Complete

- [x] Update wellness shaping so each bridged field uses source-specific `native_scale`.
  - Apply the Step 1 matrix only inside wellness provenance generation in `internal/tools/get_wellness_data.go` (`wellnessProvenanceEntry`, `wellnessFieldSource`, `wellnessNativeScale`, `wellnessReadinessNativeScale`).
  - Use sidecar-aware source/scale selection for `sleepScore` and `readiness`; keep Polar readiness precedence as `nightly_recharge_status` before `ans_charge`; use Garmin/WHOOP/Oura labels only when source evidence resolves to those providers; pair unknown source with `native_scale: unknown`.
  - No native extraction changes in `internal/intervals/wellness.go` are expected unless shaping reveals missing source evidence.
- [x] Keep response-level canonical scale labels separate from native provenance labels.
  - Do not change `internal/response/scales.go` canonical labels; verify `_meta.scales.sleepScore` remains the canonical response label while `_meta.provenance.<field>.native_scale` carries provider-native labels.
- [x] Ensure stale/provenance behavior remains intact.
  - Preserve generic native units for non-score bridged fields and keep `fetched_at`, `stale`, and `stale_reason` using the resolved source without dropping unknown-source provenance.
  - Verify with `go test ./internal/tools` that existing fresh/stale Polar fixture behavior remains green.

---

### Step 3: Fixture coverage
**Status:** ✅ Complete

- [x] Add fixtures for at least two divergent sources and assert the exact `native_scale` strings.
  - Update the existing Garmin fixture assertion to require readiness provenance `source: garmin`, `native_scale: 0-100 Garmin Body Battery`.
  - Add a WHOOP fixture under `internal/intervals/testdata/wellness` (the directory loaded by `loadWellnessFixture`) and assert `sleepScore` provenance `source: whoop`, `native_scale: 0-100 WHOOP sleep performance percentage`, plus `readiness` provenance `source: whoop`, `native_scale: 0-100 WHOOP recovery score`.
  - Do not add unused fixtures under `internal/tools/testdata/wellness`.
- [x] Add unknown-source fallback test.
  - Keep the existing `custom_fields.json` unknown-provider case intentional and assert `source: unknown`, `native_scale: unknown`; do not count manual-only rows because they intentionally omit provenance.
- [x] Run targeted wellness tests.
  - Run `go test ./internal/tools -run 'TestGetWellnessData(Fixtures|NullStrippingAndIncludeFull)'`.
  - Run `go test ./internal/intervals -run Wellness` for native extraction coverage.

---

### Step 4: Docs and verification
**Status:** ✅ Complete

- [x] Update tool docs/reference wording.
  - Update `get_wellness_data` source/catalog wording so generated reference explains `_meta.provenance.<field>.native_scale` carries provider-native sleep/readiness labels for Garmin, WHOOP, Oura, and Polar, and `unknown` when unresolved.
  - Do not change canonical `_meta.scales.sleepScore` wording in `internal/response/scales.go`.
  - Run `make docs-tools` if registry/schema wording changes and include generated `web/data/tools.json` diff instead of manually editing generated reference output.
- [x] Update CHANGELOG.md.
  - Add a concrete `[Unreleased]` / `Changed` entry for `get_wellness_data` provenance reporting provider-native sleep/readiness scale labels and unknown fallback.
- [x] Run full quality gate.
  - Run and record `make test`, `make build`, and `make lint` in this step; Step 5 will repeat/record task-level verification.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
  - Run `go test ./internal/tools -run 'TestGetWellnessData(Fixtures|NullStrippingAndIncludeFull)'`.
  - Run `go test ./internal/intervals -run Wellness`.
  - Run `go test ./cmd/gendocs` because Step 4 updated generated docs/golden catalog data.
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures
  - Record the exact failing command, concise error summary, and why any remaining failure is demonstrably pre-existing/unrelated; fix task-related failures, stale generated docs/goldens, formatting/lint failures, and wellness regressions before completion.

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
| R001 | plan | 1 | APPROVE | `.reviews/R001-plan-step1.md` |
| R002 | code | 1 | REVISE | `.reviews/R002-code-step1.md` |
| R003 | code | 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | plan | 2 | REVISE | `.reviews/R004-plan-step2.md` |
| R005 | plan | 2 | APPROVE | `.reviews/R005-plan-step2.md` |
| R006 | code | 2 | APPROVE | `.reviews/R006-code-step2.md` |
| R007 | plan | 3 | REVISE | `.reviews/R007-plan-step3.md` |
| R008 | plan | 3 | APPROVE | `.reviews/R008-plan-step3.md` |
| R009 | code | 3 | APPROVE | `.reviews/R009-code-step3.md` |
| R010 | plan | 4 | REVISE | `.reviews/R010-plan-step4.md` |
| R011 | plan | 4 | APPROVE | `.reviews/R011-plan-step4.md` |
| R012 | code | 4 | APPROVE | `.reviews/R012-code-step4.md` |
| R013 | plan | 5 | REVISE | `.reviews/R013-plan-step5.md` |
| R014 | plan | 5 | APPROVE | `.reviews/R014-plan-step5.md` |
| R015 | code | 5 | APPROVE | `.reviews/R015-code-step5.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Generated docs tests compare against `cmd/gendocs/testdata/tools.golden.json` as well as `web/data/tools.json`; registry wording changes require regenerating both. | Fixed in scope; regenerated both docs outputs. | `cmd/gendocs/testdata/tools.golden.json`, `web/data/tools.json` |
| `internal/tools/testdata/wellness` was absent and unused; current wellness fixture helper loads from `internal/intervals/testdata/wellness`. | Created directory during preflight, but Step 3 fixtures were intentionally placed in the intervals fixture directory used by tests. | `internal/tools/get_wellness_data_test.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 13:37 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 13:37 | Step 0 started | Preflight |
| 2026-05-20 14:22 | Worker iter 1 | done in 2710s, tools: 194 |
| 2026-05-20 14:22 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

- Step 1 review: `intervals.Wellness.Native` is populated from nested/flat Polar (`sleep_score`, `nightly_recharge_status`, `ans_charge`), Garmin (`body_battery_min/max`), and Oura (`sleep_score`) sidecars; `get_wellness_data` emits `_meta.provenance` entries for bridged fields with `source`, `native_scale`, and `fetched_at` while preserving stale detection.
- Step 6 docs review: README needed no setup/catalog wording change; `web/content/reference/tools.md` is a generated wrapper and `web/data/tools.json` was regenerated; PRD wellness provenance wording was updated to match provider-native native-scale behavior.
- Step 5 full verification passed: `make test`, `make build`, and `make lint`; no remaining failures.
- Step 5 targeted verification passed: `go test ./internal/tools -run 'TestGetWellnessData(Fixtures|NullStrippingAndIncludeFull)'`, `go test ./internal/intervals -run Wellness`, and `go test ./cmd/gendocs`.
- Step 4 verification: first `make test` found stale `cmd/gendocs/testdata/tools.golden.json`; regenerated it with `go run ./cmd/gendocs --out cmd/gendocs/testdata/tools.golden.json`, then `make test && make build && make lint` passed.
- Step 3 verification: `TestGetWellnessDataFixtures/custom_null_fields_and_unknown_source_provenance` asserts `source: unknown` and `native_scale: unknown`; targeted fixture run passed after adding Garmin/WHOOP assertions.
- Step 2 verification: `go test ./internal/tools ./internal/intervals` and `go test ./internal/tools` passed after source-specific provenance application; existing Polar fresh/stale provenance expectations remain green.
- Step 1 scale labels defined in `internal/tools/get_wellness_data.go`:

  | Field | Source | Native sidecar/evidence | `native_scale` | Basis |
  |---|---|---|---|---|
  | `sleepScore` | Garmin | Garmin provider evidence / observed bridge rows | `0-100 Garmin sleep score` | public Garmin sleep score scale |
  | `sleepScore` | WHOOP | WHOOP provider evidence / observed bridge rows | `0-100 WHOOP sleep performance percentage` | public WHOOP sleep performance percent scale |
  | `sleepScore` | Oura | `oura.sleep_score` / `oura_sleep_score` | `0-100 Oura sleep score` | observed fixture + public Oura score scale |
  | `sleepScore` | Polar | `polar.sleep_score` / `polar_sleep_score` | `1-100 Polar sleep_score` | observed fixture + public Polar sleep score scale |
  | `sleepScore` | unknown | no provider/native evidence | `unknown` | task fallback requirement |
  | `readiness` | Garmin | `garmin.body_battery_min/max` | `0-100 Garmin Body Battery` | public Garmin Body Battery scale |
  | `readiness` | WHOOP | WHOOP provider evidence / observed bridge rows | `0-100 WHOOP recovery score` | public WHOOP recovery score scale |
  | `readiness` | Oura | Oura provider evidence / observed bridge rows | `0-100 Oura readiness score` | public Oura readiness score scale |
  | `readiness` | Polar | `polar.nightly_recharge_status` (takes precedence when present) | `1-6 Polar nightly_recharge_status` | observed fixture + public Polar status scale |
  | `readiness` | Polar | `polar.ans_charge` only | `-10 to +10 Polar ans_charge` | public Polar ANS charge scale |
  | `readiness` | unknown | no provider/native evidence | `unknown` | task fallback requirement |
| 2026-05-20 13:40 | Review R001 | plan Step 1: APPROVE |
| 2026-05-20 13:44 | Review R002 | code Step 1: UNKNOWN |
| 2026-05-20 13:49 | Review R003 | code Step 1: APPROVE |
| 2026-05-20 13:52 | Review R004 | plan Step 2: UNKNOWN |
| 2026-05-20 13:53 | Review R005 | plan Step 2: APPROVE |
| 2026-05-20 13:59 | Review R006 | code Step 2: APPROVE |
| 2026-05-20 14:01 | Review R007 | plan Step 3: REVISE |
| 2026-05-20 14:03 | Review R008 | plan Step 3: APPROVE |
| 2026-05-20 14:06 | Review R009 | code Step 3: APPROVE |
| 2026-05-20 14:08 | Review R010 | plan Step 4: UNKNOWN |
| 2026-05-20 14:09 | Review R011 | plan Step 4: APPROVE |
| 2026-05-20 14:12 | Review R012 | code Step 4: APPROVE |
| 2026-05-20 14:14 | Review R013 | plan Step 5: UNKNOWN |
| 2026-05-20 14:16 | Review R014 | plan Step 5: APPROVE |
| 2026-05-20 14:19 | Review R015 | code Step 5: APPROVE |

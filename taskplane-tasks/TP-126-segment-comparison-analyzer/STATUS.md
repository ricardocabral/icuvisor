# TP-126: Deterministic segment-comparison analyzer workflow — Status

**Current Step:** Step 2: Add segment-comparison eval/docs
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 4
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Audit current segment analyzer activation
**Status:** ✅ Complete

- [x] Inspect `compute_activity_segment_stats` description/schema/tests and existing eval scenarios.
- [x] Confirm it supports distance-bounded first/last segment stats for pace/power/HR and exposes audit metadata without raw streams in terse mode.
- [x] Record whether a higher-level helper is warranted or whether prompt/eval hardening is sufficient.
- [x] Run targeted tests: `go test ./internal/tools`

---

### Step 2: Add segment-comparison eval/docs
**Status:** 🟨 In Progress

- [x] Add an eval scenario for comparing first 10 km vs last 10 km that expects `compute_activity_segment_stats` rather than raw `get_activity_streams` reduction in chat.
- [x] Update activity retrospective cookbook guidance with a deterministic segment-comparison prompt.
- [x] If needed, tighten tool activation text without bloating core tool descriptions.
- [x] Run targeted tests: `make eval-validate` and `go test ./internal/tools`

---

### Step 3: Add missing tests for first/last distance segments
**Status:** ⬜ Not Started

- [ ] Add or extend unit tests for distance-bounded segment stats over first and last portions of a fixture stream.
- [ ] Assert insufficient/missing stream metadata remains explicit and terse output does not dump raw stream samples.
- [ ] Run targeted tests: `go test ./internal/tools`

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing: `make test`
- [ ] Lint passes or pre-existing linter limitations are documented: `make lint`
- [ ] Build passes: `make build`
- [ ] All failures fixed or clearly documented as pre-existing

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | plan | 1 | APPROVE | `.reviews/R001-plan-step1.md` |
| R002 | code | 1 | REVISE | `.reviews/R002-code-step1.md` |
| R003 | code | 1 | APPROVE | `.reviews/R003-code-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| `compute_activity_segment_stats` already supports deterministic distance-bounded segment stats over scalar metrics including `watts`, `heart_rate`, and `velocity_smooth`; terse responses include result + `_meta` without `series`, while `include_full` gates audit slices. A new higher-level helper is not warranted for TP-126; eval/docs/tests should harden activation for first-vs-last 10 km comparisons. | Use prompt/eval/test hardening; avoid broad API expansion. | `internal/tools/compute_activity_segment_stats.go`, `internal/analysis/segment_stats.go`, `internal/tools/compute_activity_segment_stats_test.go` |
| "Last 10 km" must be translated by the workflow: call `get_activities` or `get_activity_details` for total distance (`distance_km` or `distance_mi` in preferred units), convert to meters, then call `compute_activity_segment_stats` with `start_distance_m=max(total_distance_m-10000, 0)` and `end_distance_m=total_distance_m`; first 10 km is `0..10000`. | Document in cookbook/eval; no helper needed unless future activation shows repeated failures. | `internal/tools/get_activities.go`, `internal/tools/get_activities_row.go`, `internal/tools/compute_activity_segment_stats.go` |
| Description/schema mismatch found: tool description says "maximum" and "zone-time", but the schema enum supports `mean`, `median`, `p90`, `decoupling`, `drift`, `np`, and `if`. Pace wording also needs care: segment scalar supports `velocity_smooth` in m/s, not formatted pace, so final answers should convert velocity to pace when requested. | Tighten activation text in Step 2 without bloating the description. | `internal/tools/compute_activity_segment_stats.go` |
| Current eval coverage has `compute_activity_segment_stats` only as a bonus tool in `CB-ACT-01`; no scenario currently requires first 10 km vs last 10 km segment comparisons or forbids chat-side raw-stream reduction for that use case. Existing tests cover one terse distance scalar and full decoupling audit, but do not compare first-vs-last distance windows. | Add eval/docs in Step 2 and first/last distance unit coverage in Step 3. | `scripts/eval/scenarios/cookbook_scenarios.json`, `internal/tools/compute_activity_segment_stats_test.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 13:46 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 13:46 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
| 2026-05-29 13:55 | Review R003 | code Step 1: APPROVE |
| 2026-05-29 13:57 | Review R004 | plan Step 2: APPROVE |

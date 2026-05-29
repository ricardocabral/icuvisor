# Task: TP-126 - Deterministic segment-comparison analyzer workflow

**Created:** 2026-05-29
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Strengthens analyzer activation and eval coverage for raw-stream segment comparisons. It may add prompt/eval behavior but should avoid broad API changes unless clearly justified.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-126-segment-comparison-analyzer/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Turn “compare the first 10 km with the last 10 km” into a deterministic analyzer workflow instead of letting the LLM manually reduce raw streams and hallucinate averages. Public feedback showed an AI coach miscalculated a marathon segment from available data; icuvisor already has `compute_activity_segment_stats` and should route users there.

## Evidence from forum review

- Montis segment hallucination report: https://forum.intervals.icu/t/117856/471
- Maintainer noted stream-size/LLM limitations and future SIL/API idea: https://forum.intervals.icu/t/117856/472

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — analyzer and raw-stream token-efficiency requirements.
- `docs/kr5-benchmark.md` — analyzer-token context if tool descriptions are changed.
- `scripts/eval/README.md` — eval harness shape.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/compute_activity_segment_stats.go`
- `internal/tools/compute_activity_segment_stats_test.go`
- `internal/tools/catalog_test.go`
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `web/content/cookbook/activity-retrospective.md`
- `web/content/explain/fitness-projection.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit current segment analyzer activation
- [ ] Inspect `compute_activity_segment_stats` description/schema/tests and existing eval scenarios.
- [ ] Confirm it supports distance-bounded first/last segment stats for pace/power/HR and exposes audit metadata without raw streams in terse mode.
- [ ] Record whether a higher-level helper is warranted or whether prompt/eval hardening is sufficient.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/compute_activity_segment_stats.go`
- `internal/tools/compute_activity_segment_stats_test.go`
- `STATUS.md Discoveries`

### Step 2: Add segment-comparison eval/docs
- [ ] Add an eval scenario for comparing first 10 km vs last 10 km that expects `compute_activity_segment_stats` rather than raw `get_activity_streams` reduction in chat.
- [ ] Update activity retrospective cookbook guidance with a deterministic segment-comparison prompt.
- [ ] If needed, tighten tool activation text without bloating core tool descriptions.
- [ ] Run targeted tests: `make eval-validate` and `go test ./internal/tools`

**Artifacts:**
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `web/content/cookbook/activity-retrospective.md`
- `internal/tools/compute_activity_segment_stats.go`

### Step 3: Add missing tests for first/last distance segments
- [ ] Add or extend unit tests for distance-bounded segment stats over first and last portions of a fixture stream.
- [ ] Assert insufficient/missing stream metadata remains explicit and terse output does not dump raw stream samples.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/compute_activity_segment_stats_test.go`

### Step 4: Testing & Verification

> ZERO test failures allowed. This step runs the FULL test suite as a quality gate. Earlier steps should use targeted tests for fast feedback.

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures with exact command output
- [ ] Build passes: `make build`

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note new eval/docs/test hardening.

**Check If Affected:**
- `docs/kr5-benchmark.md` — update only if toolset placement or description-token benchmark changes.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-126): complete Step N — description`
- **Bug fixes:** `fix(TP-126): description`
- **Tests:** `test(TP-126): description`
- **Hydration:** `hydrate: TP-126 expand Step N checkboxes`

## Do NOT

- Expand task scope — add tech debt to CONTEXT.md instead
- Skip tests
- Modify framework/standards docs without explicit user approval
- Load docs not listed in "Context to Read First"
- Open, copy, paraphrase, or transliterate GPL/copyleft competitor source
- Add first-party Strava/TrainingPeaks ingestion or hosted SaaS behavior
- Commit without the task ID prefix in the commit message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->

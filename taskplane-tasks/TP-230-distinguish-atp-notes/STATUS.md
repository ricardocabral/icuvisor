# TP-230: Distinguish ATP-generated notes from personal calendar notes — Status

**Current Step:** Step 5: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 11
**Iteration:** 2
**Size:** M

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing provenance and keyword paths audited

---

### Step 1: Design provenance-aware note shaping

**Status:** ✅ Complete

- [x] ATP and personal note statuses defined
- [x] plan_applied provenance rule defined
- [x] Counts cannot misclassify personal notes
- [x] Ordering and terse/full behavior defined
- [x] Exact collection, status, count, and week-association migration documented
- [x] Classification boundary, recovery policy, and Step 2 regression matrix documented
- [x] Personal-context-only unavailable behavior and wording defined

---

### Step 2: Implement and cover classification

**Status:** ✅ Complete

- [x] Provenance-aware note shape implemented
- [x] English keyword dependence removed or constrained
- [x] Personal and localized ATP fixtures covered
- [x] Real TARGET week boundaries preserved
- [x] Projection bridge remains explicit-target-only

---

### Step 3: Update schema and generated surfaces

**Status:** ✅ Complete

- [x] Output schema snapshot updated
- [x] Generated website data updated
- [x] Tool description clarifies personal context
- [x] Gendocs golden fixtures updated and generator test passes

---

### Step 4: Testing & Verification

**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Race suite passing
- [x] Lint passing
- [x] Build passes
- [x] Generated docs clean

---

### Step 5: Documentation & Delivery

**Status:** ✅ Complete

- [x] Must Update docs modified
- [x] Check If Affected docs reviewed
- [x] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | Plan | 1 | APPROVE | `.reviews/R003-plan-step1.md` |
| R004 | Code | 1 | APPROVE | `.reviews/R004-code-step1.md` |
| R005 | Plan | 2 | APPROVE | `.reviews/R005-plan-step2.md` |
| R006 | Code | 2 | APPROVE | `.reviews/R006-code-step2.md` |
| R007 | Plan | 3 | APPROVE | `.reviews/R007-plan-step3.md` |
| R008 | Code | 3 | REVISE | `.reviews/R008-code-step3.md` |
| R009 | Code | 3 | APPROVE | `.reviews/R009-code-step3.md` |
| R010 | Plan | 4 | APPROVE | `.reviews/R010-plan-step4.md` |
| R011 | Code | 4 | APPROVE | `.reviews/R011-code-step4.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| `taskplane-tasks/CONTEXT.md` and the packet `.reviews/` directory were not staged; all implementation and documentation paths in task scope exist. | Continue from PROMPT and authoritative project docs; review tooling may create `.reviews/` when invoked. | Preflight |
| `intervals.Event.PlanApplied` was already decoded from the upstream JSON — no new struct field was needed; classification required only a `strings.TrimSpace` guard. | Implementation used the existing field directly. Locale-independent provenance confirmed. | Step 2 |
| `annualTrainingPlanRecoveryHint` was the sole English-keyword recovery path: it scanned name/type/description/tags for `recovery`, `rest`, `taper`, `deload`. Removing it eliminated all locale-dependence in the note-classification branch. | Removed entirely; `recovery_note_count` and the hint field were dropped from responses. | Step 2 |
| `for_week` is `false` on both ATP-generated and personal NOTE events; it is not a provenance signal. `plan_applied` (non-empty after trimming) is the only reliable ATP provenance marker available on NOTE rows. | Decision confirms Step 1 design. No additional field needed. | Step 2 |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 12:26 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 12:26 | Step 0 started | Preflight |
| 2026-07-10 13:19 | Worker iter 1 | done in 3161s, tools: 171 |
| 2026-07-10 13:21 | Worker iter 2 | done in 122s, tools: 13 |
| 2026-07-10 13:21 | Task complete | .DONE created |

## Blockers

*None*

## Notes

- Preflight audit: `intervals.Event.PlanApplied` already decodes `plan_applied`. The annual-plan path currently groups every `NOTE` as periodization context, increments week `note_count`, and derives `recovery_hint`/`recovery_note_count` solely via `annualTrainingPlanRecoveryHint`, which scans name/type/description/tags for English substrings `recovery`, `rest`, `taper`, and `deload`. No other annual-plan keyword recovery path was found.
- Step 1 response contract: ATP-generated NOTE rows remain in `notes` and have terse `status: "atp_generated"`; personal NOTE rows move to top-level `context_notes` and have terse `status: "personal_context"`. Both collections retain useful name/description/tags/date context. ATP rows also expose the non-empty `plan_applied` timestamp in terse output; personal rows omit it.
- Step 1 classification boundary: a NOTE is ATP-generated iff `strings.TrimSpace(stringValue(event.PlanApplied)) != ""`; null, empty, and whitespace-only values are personal context. `for_week`, names, descriptions, types, and tags are never provenance signals. This provenance split applies only to NOTE rows: category-matching PLAN/TARGET rows keep their existing inclusion so explicit target/projection behavior is unchanged; their classification is not being broadened or relabeled by this note-focused fix.
- Step 1 count/association migration: summary `note_count` becomes `atp_note_count` and adds `context_note_count`; week `note_count`/`note_ids` become `atp_note_count`/`atp_note_ids`, with separate `context_note_count`/`context_note_ids`. `_meta.note_event_count` becomes `_meta.atp_note_event_count` plus `_meta.context_note_event_count`; `_meta.periodization_event_count` counts PLAN, TARGET, and ATP NOTE rows only. Personal context never contributes to an ATP count or ATP ID list. Remove `recovery_note_count` because no locale-neutral upstream recovery semantic exists.
- Step 1 ordering/full contract: independently order `notes` and `context_notes` by athlete-local start date, then upstream `updated`, then source event ID, preserving the current stable comparator; sort every ATP/context week ID list lexically. Default terse rows always expose `status`, and ATP notes expose `plan_applied`; `include_full: true` additionally exposes the unchanged raw event under `full` for either class, while false omits all `full` payloads.
- Step 1 recovery policy and regression matrix: remove `recovery_hint`, `annualTrainingPlanRecoveryHint`, and all recovery-note counts; `plan_applied` proves ATP provenance only and never recovery meaning. Step 2 tests will cover null/empty/whitespace `plan_applied` personal `Travel — Rest`, localized ATP notes sharing a non-empty timestamp, multi-day note/week associations, independent stable ordering, terse versus `include_full`, real one-day TARGET Monday-through-Sunday boundaries, and unchanged TARGET-only projection bridge rows. Existing English-recovery assertions will migrate to provenance/count assertions.
- Step 1 context-only availability contract: when a scan returns personal context NOTE rows but no PLAN, TARGET, or ATP-generated NOTE rows, retain populated `context_notes` and return `unavailable.reason: "no_periodization_events"`. Its detail will say `no PLAN, TARGET, or ATP-generated NOTE calendar events were returned for the requested range; personal NOTE rows, when present, are retained separately in context_notes` so the response does not falsely claim there were no notes.
| 2026-07-10 12:31 | Review R001 | plan Step 1: REVISE |
| 2026-07-10 12:34 | Review R002 | plan Step 1: REVISE |
| 2026-07-10 12:36 | Review R003 | plan Step 1: APPROVE |
| 2026-07-10 12:38 | Review R004 | code Step 1: APPROVE |
| 2026-07-10 12:39 | Review R005 | plan Step 2: APPROVE |
| 2026-07-10 12:49 | Review R006 | code Step 2: APPROVE |
| 2026-07-10 12:54 | Review R007 | plan Step 3: APPROVE |
| 2026-07-10 13:04 | Review R008 | code Step 3: REVISE |
| 2026-07-10 13:09 | Review R009 | code Step 3: APPROVE |
| 2026-07-10 13:11 | Review R010 | plan Step 4: APPROVE |
| 2026-07-10 13:17 | Review R011 | code Step 4: APPROVE |

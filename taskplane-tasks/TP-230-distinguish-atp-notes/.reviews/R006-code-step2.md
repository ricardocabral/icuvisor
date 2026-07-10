# R006 — Code Review — Step 2: Implement and cover classification

**Verdict: APPROVE**

All Step 2 completion criteria are met. Tests pass cleanly including the race detector. The implementation is correct, well-scoped, and consistent with the approved plan.

---

## What was reviewed

- `internal/tools/get_annual_training_plan.go` — classification logic, struct changes, removed keyword heuristic
- `internal/tools/get_annual_training_plan_provenance_test.go` — new regression file
- `internal/tools/get_annual_training_plan_test.go` — updated assertions

---

## Classification logic

`annualTrainingPlanPeriodizationEvents` now routes NOTE events by `strings.TrimSpace(stringValue(event.PlanApplied)) != ""`. Null, empty, and whitespace-only `plan_applied` values all route to `contextNotes`. Non-empty values route to `notes`. This is the correct locale-independent provenance signal from the spec.

`annualTrainingPlanRecoveryHint` and all English keyword detection (`recovery`, `rest`, `taper`, `deload`) are fully removed. The only surviving reference to `recovery_hint` is in `get_annual_training_plan_test.go` as a negative assertion (`must not infer recovery semantics from English keywords`). Good.

## Struct changes

Every renamed field is consistent with the plan:

| Old | New |
|-----|-----|
| `NoteCount` (summary) | `ATPNoteCount` + `ContextNoteCount` |
| `NoteCount` (week) | `ATPNoteCount` + `ContextNoteCount` |
| `RecoveryNoteCount` (week) | removed |
| `NoteIDs` (week) | `ATPNoteIDs` + `ContextNoteIDs` |
| `RecoveryHint` (note) | removed |
| `NoteEventCount` (meta) | `ATPNoteEventCount` + `ContextNoteEventCount` |
| `note_count` (JSON) | `atp_note_count` + `context_note_count` |

`PlanApplied` is added to `annualTrainingPlanNote` with `omitempty`. ATP notes carry the non-empty timestamp in terse output; context notes have an empty string that is suppressed by `omitempty`. `Status` is always emitted (`"atp_generated"` or `"personal_context"`). Both are correct per the approved plan.

Schema version bumped to `annual_training_plan.v2`. ✓

## `classifiedCount` gating

The guard was changed from `periodizationCount > 0` to `classifiedCount > 0`. This allows phases, notes, weeks, and context notes to be computed when only personal context notes are present. The `unavailable` sentinel still fires when `periodizationCount == 0` (PLAN + TARGET + ATP NOTE count), so the personal-context-only response correctly sets both `context_notes` and `unavailable`. The updated `unavailable.detail` wording references `context_notes` and the test asserts `strings.Contains(response.Unavailable.Detail, "context_notes")`. ✓

## ID prefix separation

`annualTrainingPlanNotes` now takes `status string, idPrefix string`. ATP notes get prefix `"note"` (e.g. `note_atp-later`), context notes get `"context_note"` (e.g. `context_note_context-first`). This makes IDs unambiguous and sortable, and the test directly asserts the exact sorted `ATPNoteIDs` and `ContextNoteIDs` strings. ✓

## `annualTrainingPlanWeeks` week association

Two separate loops after targets populate `ATPNoteIDs`/`ATPNoteCount` from `notes` and `ContextNoteIDs`/`ContextNoteCount` from `contextNotes`. `sort.Strings` is applied to both ID lists before the week slice is finalized. This is clean and correct.

## Provenance test file (`get_annual_training_plan_provenance_test.go`)

Covers all required regression cases from the plan:
- Null, empty, and whitespace-only `plan_applied` → personal context
- Non-empty `plan_applied` (localized Portuguese/Spanish names) → ATP generated
- Terse default: `PlanApplied` present on ATP notes, absent on context notes, `Full == nil`
- `include_full` true/false for both note classes
- Multi-day ATP note spanning two ISO weeks → association in both weeks
- One-day TARGET → full Monday–Sunday week boundary with correct bridge row
- Personal context note in TARGET week stays outside `ATPNoteCount`
- Personal-context-only response: `context_notes` populated, `Unavailable` set, zero ATP counts

Helper functions `noteSourceIDs` and `annualTrainingPlanWeeksByStart` are in the provenance test file. They're unexported and only used in tests; appropriate.

## Existing tests updated correctly

- `n1` fixture in `TestGetAnnualTrainingPlanExtractsPhasesTargetsNotesAndBridge` gained `plan_applied` to keep it ATP-classified.
- Old assertions on `note_count` / `recovery_note_count` / `recovery_hint` replaced with `atp_note_count` / `context_note_count` / `status` / `plan_applied` assertions.
- `context_notes` empty array added to empty-response guard.

## Test run

```
go test ./internal/tools -run AnnualTrainingPlan   PASS
go test -race ./internal/tools -run AnnualTrainingPlan   PASS
go test ./...   all packages PASS
go vet ./internal/tools/...   clean
```

---

## Minor observations (no action required for Step 2)

1. **Tool description still says "recovery/context notes"** — The current `getAnnualTrainingPlanDescription` string still uses that phrase. Step 3 is responsible for updating the description to clarify the provenance-based split; flagging here so Step 3 doesn't overlook it.

2. **Schema snapshot and web data not updated** — Expected: these are Step 3 artifacts. `TestCheckSnapshotFreshness` and `TestCheckSchemaStability` both pass because the input schema hasn't changed.

3. **CHANGELOG.md and PRD not yet updated** — Step 5 artifacts. Expected at this stage.

---

Step 2 completion criteria from PROMPT.md:

- [x] Provenance fields added without silently dropping useful context
- [x] English keyword recovery classification removed
- [x] Fixtures for personal `Travel — Rest` note with null `plan_applied` and localized ATP notes
- [x] Real one-day TARGET spans produce Monday-through-Sunday week boundaries
- [x] Projection bridge remains explicit-TARGET-only
- [x] `go test ./internal/tools -run AnnualTrainingPlan` passes

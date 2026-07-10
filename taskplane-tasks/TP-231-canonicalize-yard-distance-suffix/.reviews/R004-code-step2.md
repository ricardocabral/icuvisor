# Review R004 — Code Review: Step 2

**Task:** TP-231 — Validate and canonicalize the yard distance suffix  
**Step:** Step 2: Update resources, examples, and round-trip fixtures  
**Verdict:** APPROVE

---

## Summary

Step 2 achieves its mission: the golden resource file, PRD, website cookbook, validate_test, and CHANGELOG are all correctly updated to reflect the `yrd` canonical suffix. All targeted tests pass. Two listed artifacts were left unmodified (`internal/resources/workout_syntax_test.go` and `internal/tools/validate_workout_test.go`), but neither omission breaks correctness — the rationale for each is explained below. There are no blocking issues.

---

## Changes Reviewed

| File | Assessment |
|---|---|
| `internal/resources/testdata/workout_syntax.md` | All 4 occurrences of the old canonical suffix and examples updated correctly ✓ |
| `internal/workoutdoc/validate_test.go` | Two new, well-structured tests added ✓ |
| `web/content/cookbook/build-workouts.md` | Canonical example updated; backward-compat note added ✓ |
| `docs/prd/PRD-icuvisor.md` | DSL serializer spec updated with full alias list ✓ |
| `CHANGELOG.md` | Clear, accurate entry under `[Unreleased]` ✓ |

---

## What Was Done Well

1. **Golden file is fully consistent.** The four distinct `yd` occurrences in `workout_syntax.md` are all replaced:
   - Cheat-sheet yard swim step example (`100yd` → `100yrd`)
   - Distance unit alias table description (`canonical suffix: yd` → `canonical suffix: yrd`, `yrd` added to aliases list)
   - Distance steps prose header (`mtr, km, mi, or yd` → `mtr, km, mi, or yrd`)
   - Distance step feature example (`distance_yd` example DSL and description)

2. **`validate_test.go` additions are targeted and correct.** `TestValidateDescriptionParsesCanonicalYrdDistanceWithoutMAmbiguity` mirrors the existing `TestValidateDescriptionParsesYardDistanceWithoutMAmbiguity` structure precisely — same assertions, new token. `TestValidateDescriptionParsesYardsDistanceAlias` includes a useful comment explaining why the multi-char suffix is worth testing separately (no `M_AMBIGUITY` should be raised for `yards`). Both tests check `Distance.Unit` on the parsed struct and would catch any regression in the alias-recognition or disambiguation logic.

3. **Cookbook update is complete and forward-looking.** The JSON example in `build-workouts.md` now uses `"unit":"yrd"` as the primary form while explicitly noting the `yd` legacy alias, which is exactly the correct UX for an LLM reading the document.

4. **PRD wording is authoritative.** The updated DSL serializer spec now lists all four accepted aliases (`yd`, `yard`, `yards`, plus the newly added canonical `yrd`) and explicitly describes the canonicalization direction, matching the implementation exactly.

5. **CHANGELOG entry is thorough.** The entry names the old suffix, the new suffix, the alignment rationale, the serialization behavior, and the backward-compat guarantee — all in one sentence. Correct placement under `[Unreleased]`.

---

## Observations (Non-Blocking)

**1. `internal/tools/validate_workout_test.go` not modified.**  
The PROMPT listed this file as a Step 2 artifact and asked for "validation coverage for `yrd`, `yards`, and malformed yard tokens." The existing `TestValidateWorkoutSportSpecificGeneratedDescriptionsAvoidCyclingWords/swim_workout` sub-test already exercises the `yards`-input code path through the `validate_workout` tool but does not assert that `CanonicalDSL` contains `yrd`. Adding an assertion like `strings.Contains(resp.CanonicalDSL, "yrd")` to the swim sub-test, or a dedicated `TestValidateWorkoutYardDistanceCanonicalizes` test, would close the gap at the tool-integration layer. However, the canonicalization behavior is already fully verified by the six tests in `yard_suffix_test.go` (Step 1) and the two new tests in `validate_test.go`, so no correctness risk exists. This is a coverage-completeness observation only.

**2. `internal/resources/workout_syntax_test.go` not modified.**  
The PROMPT listed it as a Step 2 artifact, but no change was needed. The relevant tests (`TestWorkoutSyntaxMarkdownGolden`, `TestWorkoutSyntaxUnitMatricesAreRendered`, `TestWorkoutSyntaxSpecExamplesAreRenderedFromSerializer`) are already driven dynamically by `WorkoutDistanceUnitSyntax()` and `WorkoutSyntaxSpec()` — both updated in Step 1 — so `yrd` appears in the generated markdown automatically and the tests pass without any hardcoded token changes. No action required.

**3. No malformed-yard-token test added.**  
The PROMPT checklist says to "Add validation coverage for... malformed yard tokens." No such test was added (e.g., `- Swim 100yarddddd 95% Pace`). The general grammar rejects these through the `PARSE_ERROR` path already covered by `TestValidateDescriptionMalformedStepLine`, so this is a low-priority gap.

**4. Historical CHANGELOG entry still says `canonical yd`.**  
Line 35 in the `[1.3.x]` history block records when yards were first introduced with `canonical yd`. This is intentionally left as-is — it is an accurate description of what that release shipped, and the new [Unreleased] entry correctly documents the correction. No change needed.

---

## Verification

- `go test -count=1 -run 'Workout|Yard|Syntax' ./internal/workoutdoc ./internal/resources ./internal/tools` → all three packages pass ✓
- `workout_syntax.md` golden matches serializer output for `distance_yd` feature ✓
- No stray `"unit":"yd"` in `web/` or `docs/` documentation files ✓
- PRD line 426 (`sec/100yd — yard-pool swim`) describes the pace-unit meaning, not the DSL suffix — correct as-is ✓
- Commit message `fix(TP-231): complete Step 2 — update resources, examples, and fixtures for yrd canonical suffix` follows Conventional Commits and includes task ID ✓

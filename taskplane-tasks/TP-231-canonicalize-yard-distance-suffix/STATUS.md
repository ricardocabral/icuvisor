# TP-231: Validate and canonicalize the yard distance suffix — Status

**Current Step:** Step 4: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 4
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Public upstream yard syntax confirmed

---

### Step 1: Update parser and canonical serializer

**Status:** ✅ Complete

- [x] Add yrd, yard, yards to distanceTokenRE (before yd in alternation)
- [x] Update workoutDistanceUnits: Canonical → yrd, add yrd to Aliases, update Description
- [x] Update CheatSheet.Examples yard DSL string → 100yrd
- [x] Update SyntaxExample distance_yd description → canonicalizes to yrd
- [x] Update distance_steps feature description to mention yrd
- [x] Update TestWorkoutDocYardDistanceSerializeParseValidate expectation → 100yrd
- [x] Update TestWorkoutDocDistanceAliasesRemainCanonical yards row → 100yrd
- [x] Add yard_suffix_test.go with alias round-trips and description-token checks
- [x] go test ./internal/workoutdoc passes cleanly

---

### Step 2: Update resources, examples, and round-trip fixtures

**Status:** ✅ Complete

- [x] Update workout_syntax.md golden file (yd → yrd)
- [x] Update validate_test.go yard coverage
- [x] Update web/content/cookbook/build-workouts.md
- [x] Update docs/prd/PRD-icuvisor.md
- [x] Update CHANGELOG.md
- [x] Run targeted tests: go test ./internal/workoutdoc ./internal/resources ./internal/tools -run 'Workout|Yard|Syntax'

---

### Step 3: Testing & Verification

**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Race suite passing
- [x] Lint passing
- [x] Build passes
- [x] Resource and golden diff clean

---

### Step 4: Documentation & Delivery

**Status:** ✅ Complete

- [x] Must Update docs modified
- [x] Check If Affected docs reviewed
- [x] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| `SECS_100Y` pace unit description uses "sec/100yd" (yards the unit, not the DSL token) — correctly left unchanged | In-scope; not a DSL token | `docs/prd/PRD-icuvisor.md:426` |
| `internal/response/units.go` has a separate `"yd"` for API response swim pool distance display — this is activity data shaping, not workout DSL | Out of scope for TP-231 | `internal/response/units.go` |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 13:21 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 13:21 | Step 0 started | Preflight |
| 2026-07-10 13:51 | Worker iter 1 | done in 1799s, tools: 130 |
| 2026-07-10 13:51 | Task complete | .DONE created |

## Blockers

*None*

## Notes

*Reserved for execution notes*
| 2026-07-10 13:29 | Review R001 | plan Step 1: REVISE |
| 2026-07-10 13:36 | Review R002 | code Step 1: REVISE |
| 2026-07-10 13:38 | Review R003 | code Step 1: APPROVE |
| 2026-07-10 13:48 | Review R004 | code Step 2: APPROVE |

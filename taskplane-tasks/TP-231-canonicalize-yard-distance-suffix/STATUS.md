# TP-231: Validate and canonicalize the yard distance suffix — Status

**Current Step:** Step 0: Preflight
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 1
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

**Status:** 🟨 In Progress

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

**Status:** ⬜ Not Started

- [ ] Canonical output expectations use yrd
- [ ] Legacy yd to canonical yrd round trip covered
- [ ] Yard validation scenarios covered
- [ ] Website and PRD updated

---

### Step 3: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Race suite passing
- [ ] Lint passing
- [ ] Build passes
- [ ] Resource and golden diff clean

---

### Step 4: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 13:21 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 13:21 | Step 0 started | Preflight |

## Blockers

*None*

## Notes

*Reserved for execution notes*
| 2026-07-10 13:29 | Review R001 | plan Step 1: REVISE |

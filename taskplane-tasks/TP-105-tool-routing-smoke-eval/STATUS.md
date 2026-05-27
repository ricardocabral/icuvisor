# TP-105: Tool routing smoke eval — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-05-26
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing benchmark/eval patterns identified

---

### Step 1: Design eval fixture and expected-result format
**Status:** ⬜ Not Started

- [ ] Fixture format defined
- [ ] Initial routing cases added
- [ ] Safe/full destructive-tool expectations represented where practical
- [ ] Fixture loading/result comparison tests passing

---

### Step 2: Implement opt-in first-tool-call runner
**Status:** ⬜ Not Started

> ⚠️ Hydrate: Expand based on chosen implementation language/provider path.

- [ ] Tool definitions loaded without executing handlers
- [ ] Provider call guarded by explicit environment configuration
- [ ] First-tool/no-tool result captured and reported
- [ ] Normal tests remain network-free

---

### Step 3: Wire command and documentation
**Status:** ⬜ Not Started

- [ ] Eval command or Make target added/documented
- [ ] Environment variables and guarantees documented
- [ ] Opt-in/non-default-CI behavior documented
- [ ] Changelog/docs updated if needed

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Targeted tests passing
- [ ] FULL test suite passing: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] Optional provider-backed eval run recorded if credentials are available
- [ ] All failures fixed or documented as pre-existing unrelated failures

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-26 | Task staged | PROMPT.md and STATUS.md created |

---

## Blockers

*None*

---

## Notes

- Tracking issue: https://github.com/ricardocabral/icuvisor/issues/32

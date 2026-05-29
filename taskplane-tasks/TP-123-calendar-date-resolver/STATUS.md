# TP-123: Calendar date resolver and future date anchors — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Design deterministic date surface
**Status:** ⬜ Not Started

- [ ] Inspect existing `_meta.as_of`, `get_today`, `get_activities`, `get_events`, and prompt guidance for date anchors.
- [ ] Decide whether to add a small read-only tool such as `resolve_calendar_dates` or to harden existing date metadata/prompts without a new tool.
- [ ] Document the chosen surface and non-goals in STATUS.md Discoveries, including why it avoids model date arithmetic.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`

---

### Step 2: Implement date anchors and tests
**Status:** ⬜ Not Started

- [ ] Implement the chosen deterministic date anchor behavior using athlete timezone, local date, weekday, and offsets.
- [ ] Add tests covering current day, future day offsets, timezone boundaries, and invalid input if a new tool is added.
- [ ] Update catalog/schema snapshots if the public tool surface changes.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`

---

### Step 3: Add activation guidance and eval coverage
**Status:** ⬜ Not Started

- [ ] Add or update eval/cookbook prompt text so prompts that mention future weeks or “tomorrow” use the deterministic date anchor.
- [ ] Add an eval scenario for a known-bad weekday/date pairing such as “Monday May 26” when the local date says otherwise.
- [ ] Ensure guidance does not ask the assistant to infer dates from UTC.
- [ ] Run targeted tests: `make eval-validate`

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

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*

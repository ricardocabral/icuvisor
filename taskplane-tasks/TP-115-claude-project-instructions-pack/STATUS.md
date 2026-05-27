# TP-115: Claude Project instructions pack — Status

**Current Step:** Step 3: Link and deduplicate with existing recipes
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 2
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing Claude, cookbook, and MCP prompt docs reviewed for overlap

---

### Step 1: Design the instruction pack structure
**Status:** ✅ Complete

- [x] Placement chosen for the instruction pack
- [x] Reusable instruction blocks defined for base discipline, weekly review, recovery check, race-week taper, and stale/missing data handling
- [x] Exact registered tool/prompt names identified before use

**Plan-review checkpoint**

---

### Step 2: Add copy-paste Claude Project instructions
**Status:** ✅ Complete

- [x] Complete base Project instruction block added
- [x] Optional blocks added for weekly review, recovery check, and race-week taper
- [x] Usage guidance included for where to paste instructions, when to start a new chat, and secret safety
- [x] Tool-grounding requirements included

---

### Step 3: Link and deduplicate with existing recipes
**Status:** 🟨 In Progress

- [ ] New guide linked from Claude connection docs and guides index
- [ ] Relevant cookbook/prompt-library pages link to the new guide where useful
- [ ] MCP Prompts versus Project instructions clarified if needed
- [ ] `CHANGELOG.md` updated under `[Unreleased]` if appropriate

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Docs/site build passing: `make web-build`
- [ ] Internal links and docs navigation checked
- [ ] FULL test suite run if non-doc/generated app files are touched: `make test`
- [ ] Build passes if app strings or generated assets are touched: `make build`
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
| 1 | plan | 1 | APPROVE | |
| 2 | plan | 2 | APPROVE | |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-27 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 17:46 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 17:46 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

- Step 1 placement plan: publish a new copy-paste guide at `web/content/guides/claude-project-instructions.md`, link it from the Guides index and Claude connection docs, and add short pointers from relevant cookbook recipes instead of duplicating their full prompts.
- Step 1 instruction-block plan: base block covers athlete-local timezone/date anchors, source-tool citations, terse-by-default tool use, missing/stale data caveats, no invented metrics, and no secrets/config in instructions; optional add-ons cover weekly review, recovery/readiness, and race-week taper behavior without replacing the server-side MCP prompt workflows.
- Step 1 registered-name plan: relevant prompt names documented in `web/content/reference/resources-prompts.md` are `weekly_review`, `recovery_check`, `race_week_taper`, `training_analysis`, and `weekly_planning` (the full documented set also includes `coach_roster_triage`); tool names verified in `web/data/tools.json` include `get_today`, `get_athlete_profile`, `get_wellness_data`, `get_fitness`, `get_training_summary`, `get_activities`, `get_events`, `get_training_plan`, `compute_zone_time`, `compute_load_balance`, `compute_compliance_rate`, `analyze_trend`, `get_fitness_projection`, and `icuvisor_list_advanced_capabilities`.
| 2026-05-27 17:50 | Review R001 | plan Step 1: APPROVE |
| 2026-05-27 17:52 | Review R002 | plan Step 2: APPROVE |

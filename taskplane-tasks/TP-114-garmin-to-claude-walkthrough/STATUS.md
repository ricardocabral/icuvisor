# TP-114: Garmin to Claude walkthrough — Status

**Current Step:** Step 3: Link the walkthrough from discovery surfaces
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 3
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing tutorial, connect, and cookbook structure reviewed

---

### Step 1: Plan the tutorial and visual treatment
**Status:** ✅ Complete

- [x] Redacted mock screenshots, lightweight diagrams, or existing UI-free documentation pattern chosen
- [x] Tutorial path defined from device sync to intervals.icu through first Claude prompts
- [x] Privacy guardrails for visuals and prompts identified

**Plan-review checkpoint**

---

### Step 2: Add the walkthrough content
**Status:** ✅ Complete

- [x] New tutorial page created for the Garmin → intervals.icu → icuvisor → Claude journey
- [x] Copy-paste prompts included for first call, weekly review, recovery check, and missing-data troubleshooting
- [x] Source limitations explained plainly
- [x] Redacted/mock visuals or diagrams added if useful and safe

---

### Step 3: Link the walkthrough from discovery surfaces
**Status:** ✅ Complete

- [x] Tutorial linked from tutorial index and relevant Claude connection docs
- [x] Homepage or cookbook pointer added only if useful
- [x] `CHANGELOG.md` updated under `[Unreleased]` if appropriate

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Docs/site build passing: `make web-build`
- [ ] Rendered links and image references checked
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
| R001 | plan | 1 | APPROVE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 2 | UNAVAILABLE | — |
| R003 | plan | 3 | APPROVE | `.reviews/R003-plan-step3.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| No existing Mermaid usage was found in `web/content`; the tutorial uses a plain Markdown text diagram instead of relying on Mermaid rendering. | Applied in Step 2 | `web/content/tutorials/garmin-to-claude.md` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-27 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 18:13 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 18:13 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

### Step 1 tutorial plan

- Visual treatment: use a lightweight Mermaid flow diagram plus a short text "grounded answer" example instead of screenshots. This matches the site's UI-free documentation patterns, avoids private training images, and still shows the Garmin/device-provider → intervals.icu → local icuvisor → Claude path.
- Tutorial path: start with device-provider sync already landing in intervals.icu, then verify intervals.icu has recent data, install/setup icuvisor using the existing macOS/Claude docs, connect Claude Desktop or Claude Code, ask a first grounded question, then graduate to weekly review, recovery check, and missing-data troubleshooting prompts.
- Privacy guardrails: no real screenshots, athlete IDs, API keys, access tokens, location-specific activity names, medical/wellness values, or unique training numbers. Prompts should ask Claude to use icuvisor/intervals.icu data without pasting secrets, and examples should use synthetic values with explicit "example" language.
- R001 plan review approved; follow-ups for Step 2: verify Mermaid support or fall back to plain Markdown/ASCII, keep intervals.icu-only source language explicit, reuse setup/prompt docs instead of duplicating long sections, and use only synthetic examples.

*Reserved for execution notes*
| 2026-05-27 18:17 | Review R001 | plan Step 1: APPROVE |
| 2026-05-27 18:22 | Review R003 | plan Step 3: APPROVE |

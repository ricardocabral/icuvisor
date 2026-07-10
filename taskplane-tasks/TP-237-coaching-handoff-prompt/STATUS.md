# TP-237: Add coaching conversation handoff prompt — Status

**Current Step:** Step 1: Define the handoff contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 1
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Prompt, pack, golden, and eval patterns reviewed

---

### Step 1: Define the handoff contract

**Status:** 🟨 In Progress

- [ ] Compact handoff sections defined
- [ ] Conversation statements separated from tool evidence
- [ ] Privacy and data-quality rules defined
- [ ] Read-only portable workflow defined
- [ ] R001 exact section order, acceptance semantics, and evidence-row shape locked
- [ ] R001 athlete-local date, record-window, and freshness semantics locked
- [ ] R001 stale, missing, unavailable, partial-day, tool-failure, and pagination handling locked
- [ ] R001 privacy, portability, manual-review, and sensitive-detail boundaries locked
- [ ] R001 minimum read route, terse payload policy, advanced fallback, and bounded arguments locked
- [ ] R001 Step 1 artifact ownership and non-vacuous verification clarified
- [ ] R001 portable pack index added to allowed documentation scope

---

### Step 2: Register the prompt and add golden coverage

**Status:** ⬜ Not Started

- [ ] Prompt implemented and registered
- [ ] Catalog expectations updated
- [ ] Focused test and golden fixture added
- [ ] Advanced-tool fallback behavior covered

---

### Step 3: Publish the portable workflow and eval

**Status:** ⬜ Not Started

- [ ] Cookbook page and client pack added
- [ ] Eval scenario added
- [ ] References, PRD, and changelog updated

---

### Step 4: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Prompt eval validation passing
- [ ] Lint passing
- [ ] All failures fixed
- [ ] Build passes
- [ ] Markdown and diff clean

---

### Step 5: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Tier 2 `taskplane-tasks/CONTEXT.md` is absent in this worktree; all scoped source/doc paths exist | Use PROMPT.md plus authoritative project docs and established repository patterns | Preflight |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 17:42 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 17:42 | Step 0 started | Preflight |

## Blockers

*None*

## Notes

### Step 1 reviewed implementation plan (R001)

- **Output contract, in order:** (1) `Handoff scope` with athlete-local generated-on date, timezone, and covered windows; (2) `Conversation-stated context` with separate Goals, Constraints, and Accepted decisions lists; (3) `Icuvisor evidence` as compact rows `Claim | Source tool | Athlete-local evidence date/window | Freshness/as-of`; (4) `Current plan state` containing only sourced calendar/training-plan state; (5) `Data gaps and unresolved questions`; (6) `Next actions`. A decision enters Accepted decisions only when the user explicitly accepted or stated it; assistant suggestions, model summaries, and calendar state do not become user decisions.
- **Dates/freshness:** call `get_athlete_profile` for timezone and `resolve_calendar_dates` for today/relative anchors, then use returned athlete-local dates. Record date/window describes when evidence applies; `as_of`/provider freshness describes how current it is. Preserve trustworthy returned freshness markers, label absent timestamps `not provided`, and never invent or require hidden `fetched_at` debug metadata.
- **Data quality:** surface `_meta.stale`, `_meta.missing_fields`, unavailable/Strava-blocked data, current-day partial rows, and unresolved tool failures. Missing is never zero and chat memory never fills a tool gap. When `next_page_token` exists, fetch pages needed for a completeness claim or label the result partial with covered window/count; omit opaque tokens from output.
- **Privacy/portability:** always exclude credentials, API/OAuth tokens, secrets, raw athlete IDs, local/config paths, raw payloads, raw streams, pagination tokens, and transport/debug metadata. Omit health details, precise locations, and private free-text notes by default; include only a user-approved minimum. The athlete manually reviews and copies Markdown into a fresh Claude, ChatGPT, Cursor, or other client chat; never claim automatic import, persistence, or memory.
- **Read-only route:** always use `get_athlete_profile` and `resolve_calendar_dates`; use terse `get_events`, `get_training_plan`, `get_fitness`, `get_training_summary`, `get_activities`, and `get_wellness_data` only as needed for the compact sections. Optional deterministic analyzers may preserve evidence already material to the conversation; if absent, call `icuvisor_list_advanced_capabilities`, name the gap, and do not calculate substitutes in chat. Never use `include_full`, raw streams, write tools, or delete tools.
- **Arguments:** `lookback_days` is optional, defaults to 28, and accepts 1-90; `race_context_days` is optional, defaults to 90, and accepts 1-365. Invalid values return short user-facing errors. Race context remains sourced from athlete-local events/plan data, not inferred from chat.
- **Artifacts/checkpoint:** Step 1 creates the prompt function, portable pack, and focused `TestCoachingHandoff...` contract tests, making `go test ./internal/prompts -run 'CoachingHandoff'` non-vacuous. Step 2 only registers it, updates shared catalog/golden expectations, and adds the golden fixture. Step 3 updates `docs/prompts/README.md` along with public docs.
| 2026-07-10 17:46 | Review R001 | plan Step 1: REVISE |

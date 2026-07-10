# TP-237: Add coaching conversation handoff prompt — Status

**Current Step:** Step 5: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 7
**Iteration:** 2
**Size:** M

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Prompt, pack, golden, and eval patterns reviewed

---

### Step 1: Define the handoff contract

**Status:** ✅ Complete

- [x] Compact handoff sections defined
- [x] Conversation statements separated from tool evidence
- [x] Privacy and data-quality rules defined
- [x] Read-only portable workflow defined
- [x] R001 exact section order, acceptance semantics, and evidence-row shape locked
- [x] R001 athlete-local date, record-window, and freshness semantics locked
- [x] R001 stale, missing, unavailable, partial-day, tool-failure, and pagination handling locked
- [x] R001 privacy, portability, manual-review, and sensitive-detail boundaries locked
- [x] R001 minimum read route, terse payload policy, advanced fallback, and bounded arguments locked
- [x] R001 Step 1 artifact ownership and non-vacuous verification clarified
- [x] R001 portable pack index added to allowed documentation scope

---

### Step 2: Register the prompt and add golden coverage

**Status:** ✅ Complete

- [x] Prompt implemented and registered
- [x] Catalog expectations updated
- [x] Focused test and golden fixture added
- [x] Advanced-tool fallback behavior covered

---

### Step 3: Publish the portable workflow and eval

**Status:** ✅ Complete

- [x] Cookbook page and client pack added
- [x] Eval scenario added
- [x] References, PRD, and changelog updated
- [x] R004 cookbook and canonical pack publish one six-section manual fresh-chat contract
- [x] R004 self-contained `CB-HANDOFF-*` eval enforces required reads, source/freshness separation, privacy, and forbidden routes
- [x] R004 registry-backed prompt and pack discovery surfaces reconciled without stale counts
- [x] R004 valid front matter, internal links, PRD catalog, and Unreleased entry verified

---

### Step 4: Testing & Verification

**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Race detector suite passing
- [x] Prompt eval validation passing
- [x] Lint passing
- [x] All failures fixed
- [x] Build passes
- [x] Hugo documentation build passes
- [x] Markdown and diff clean

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
| R002 | Plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | Plan | 2 | APPROVE | `.reviews/R003-plan-step2.md` |
| R004 | Plan | 3 | REVISE | `.reviews/R004-plan-step3.md` |
| R005 | Plan | 3 | APPROVE | `.reviews/R005-plan-step3.md` |
| R006 | Plan | 4 | REVISE | `.reviews/R006-plan-step4.md` |
| R007 | Plan | 4 | APPROVE | `.reviews/R007-plan-step4.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Tier 2 `taskplane-tasks/CONTEXT.md` is absent in this worktree; all scoped source/doc paths exist | Use PROMPT.md plus authoritative project docs and established repository patterns | Preflight |
| All Check If Affected surfaces enumerated handoff-adjacent content | Updated Claude Project fresh-chat guidance, cookbook card/catalog count, and README downloadable-pack list to match the canonical handoff workflow | Step 5 |
| MCP transport protocol coverage hard-coded the prior ten-prompt catalog | Updated the cross-transport prompt count and sorted name expectation to include `coaching_handoff`; full and race suites pass | Step 4 |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 17:42 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 17:42 | Step 0 started | Preflight |
| 2026-07-10 17:52 | Worker iter 1 | done in 599s, tools: 64 |
| 2026-07-10 18:24 | Worker iter 2 | done in 1902s, tools: 157 |
| 2026-07-10 18:24 | Task complete | .DONE created |

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
| 2026-07-10 17:49 | Review R002 | plan Step 1: APPROVE |

### Step 3 revised implementation plan (R004)

- Publish one canonical contract: the cookbook links the existing pack and both explain registry-prompt and pasted-pack entry points, six ordered sections, strict conversation/tool source separation, evidence row dates/freshness, `not provided`, privacy exclusions, and manual review/copy into a fresh chat.
- Add a self-contained `CB-HANDOFF-*` scenario whose prompt supplies explicit goals, constraints, and accepted decisions; requires profile/date resolution and bounded evidence/plan reads; checks source/freshness/gaps/manual-copy output; and forbids writes, deletes, streams, full payloads, invented calculations/physiology, and sensitive identifiers.
- Reconcile public discovery against the eleven-entry registry: index the canonical pack, add the recipe card, document arguments/resources/read-only route/output in the prompt reference, add Claude Project fresh-chat guidance, and correct or remove stale prompt counts/lists (including omitted `ride_analysis`). Review README's pack enumeration and record its disposition.
- Use established cookbook front matter and `relref` links; add the PRD catalog row and concise Unreleased entry, then run prompt tests and eval validation.

### Step 4 revised verification plan (R006)

- Retain `make test`, `make eval-validate`, `make lint`, `make build`, and `git diff --check`.
- Run `make test-race` after the full test suite to verify prompt-handler code under the race detector.
- Run `make web-build` after the code gates to validate cookbook front matter and internal `relref` links.
- Fix every failure before delivery.

| 2026-07-10 17:57 | Review R003 | plan Step 2: APPROVE |
| 2026-07-10 18:05 | Review R004 | plan Step 3: REVISE |
| 2026-07-10 18:07 | Review R005 | plan Step 3: APPROVE |
| 2026-07-10 18:17 | Review R006 | plan Step 4: REVISE |
| 2026-07-10 18:18 | Review R007 | plan Step 4: APPROVE |

# TP-238: Add grounded fueling review prompt pack — Status

**Current Step:** Step 2: Register the prompt and add regression coverage
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 8
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Fueling and nutrition field semantics reviewed

---

### Step 1: Define the fueling evidence contract

**Status:** ✅ Complete

- [x] Define executable date-only arguments, mode precedence, validated athlete-local dates, source-tool route, pagination, and optional race context
- [x] Require all-session date-range activity reads to retain unnamed rows and report their availability separately
- [x] Add a `FuelingReview` portable-pack contract test before Step 2's function, registry, and golden-fixture work
- [x] Define a nutrition-only wellness field projection, closed vocabulary, source-labelled return layout, and missing/freshness/availability reporting
- [x] Define grams-per-hour denominator, non-negative logged-intake eligibility, range aggregation, and coverage/exclusion rules
- [x] Define read-only health, product, target, and custom-field boundaries
- [x] Resolve Step 1/2 ownership for contract verification, handler validation, prompt function/registry, golden fixture, and portable-pack discoverability

---

### Step 2: Register the prompt and add regression coverage

**Status:** 🟨 In Progress

- [x] Prompt implemented and registered across catalog, registry, and MCP protocol surfaces
- [x] Date, unit, and missing-data discipline encoded
- [x] Focused tests and golden fixture added, including name-only race-context handler rejection
- [x] Read-only and no-product-library behavior covered

---

### Step 3: Publish cookbook, portable pack, and evals

**Status:** ⬜ Not Started

- [ ] Cookbook and client pack added
- [ ] Prompt library aligned
- [ ] Eval scenarios added
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
|---|------|------|---------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | Plan | 1 | REVISE | `.reviews/R003-plan-step1.md` |
| R004 | Plan | 1 | REVISE | `.reviews/R004-plan-step1.md` |
| R005 | Plan | 1 | REVISE | `.reviews/R005-plan-step1.md` |
| R006 | Plan | 1 | APPROVE | `.reviews/R006-plan-step1.md` |
| R007 | Plan | 2 | REVISE | `.reviews/R007-plan-step2.md` |
| R008 | Plan | 2 | APPROVE | `.reviews/R008-plan-step2.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Referenced area-context file is absent; all task implementation paths are present | Proceeded using the listed Tier 3 source/docs instead | `taskplane-tasks/CONTEXT.md` |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 18:24 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 18:24 | Step 0 started | Preflight |

## Blockers

*None*

## Notes

### Step 2 registration-and-validation plan

- Register `fueling_review` in `NewRegistry`, then update every consumer: registry count/order, golden table, client-pack linkage table, terse-resource prompt list, and `internal/mcp/protocol_test.go` `prompts/list` expectation (11 to 12, sorted `fueling_review`) plus MCP prompt retrieval coverage.
- Implement strict date-only pre-render validation in `FuelingReviewPrompt` for the approved modes. `race_name` without `race_date` returns `missing race_date; provide YYYY-MM-DD`; valid race context renders same-day `get_events` bounds and `limit:100`, never name-only scan. Handler table tests cover this alongside defaults, valid modes, conflicts, partial/malformed/date-time/reversed/overlong dates, and malformed race dates.

### Step 1 revised evidence-contract plan

- **Modes, arguments, and source route:** Arguments are optional `activity_id`, `start_date`, `end_date`, `race_date`, and `race_name`. `activity_id` is mutually exclusive with either range endpoint; `start_date`, `end_date`, and `race_date` are strict athlete-local `YYYY-MM-DD` dates (not date-times), and supplied ranges require both endpoints, must be in athlete-local order, inclusive, and 1–90 days. With neither mode argument, call `resolve_calendar_dates` for offsets `-14` and `-1` and use that resolved 14-completed-day range. Step 2's handler rejects conflicting, incomplete, malformed (including date-time), reversed, or overlong arguments before rendering; its table-driven tests cover the default, valid activity/range modes, activity/date conflict, one-sided range, malformed date/date-time, reversed/over-90-day ranges, and malformed `race_date`. Activity-ID mode uses `get_athlete_profile`, then `get_activity_details` for the selected activity. Date-range mode uses `get_athlete_profile`, then `get_activities` with `include_unnamed:true` and terse pages as the index/source for duration, load, and activity carbs; it uses `get_activity_details` only for a selected session needing more context, never for every row. Both modes call `get_wellness_data` only when daily nutrition evidence is requested or useful, with `fields:["kcalConsumed","carbohydrates","protein","fatTotal"]` plus only explicit user-requested custom codes; returned aliases remain `calories_intake`, `carbs_g`, `protein_g`, and `fat_g`, and unavailable nutrition freshness/provenance remains unavailable rather than widening into health fields. They call `get_training_summary` only for aggregate load context, and `get_events` only for a supplied `race_date`; `race_name` alone asks for `race_date` rather than scanning an unbounded calendar. Race reads use athlete-local `oldest`/`newest` equal to `race_date`, `limit:100`, and label any `_meta.truncated` response partial rather than treating a missing match as unconfirmed. The prompt fetches all activity pages needed before saying a range is complete; otherwise it reports count/window as partial and omits the opaque token.
- **Evidence and output:** `carbs_ingested_g` is a numeric athlete-logged during-activity intake; an absent key is a missing log and a numeric zero remains logged zero. `carbs_used_g` is an upstream used/burned estimate, never intake or an intake substitute. Wellness `carbs_g`, `calories_intake`, `protein_g`, and `fat_g` are daily dietary fields, separate from activity records and not summed/subtracted with them. `calories_burned` and training load are context only. Requested custom fields retain their exact code and unknown meaning. Return separate sections for sourced activity evidence, daily wellness evidence, race/calendar context, labelled calculations, coverage/data gaps, and separately labelled non-personalized general guidance.
- **Calculation:** Use `moving_time_seconds` only. Per-session `logged carbs/hour = carbs_ingested_g / (moving_time_seconds / 3600)` with units `g/h`; calculate only for a returned non-negative numeric ingested value and positive moving time, never unavailable/Strava-blocked rows. A logged zero is valid and yields `0 g/h`; an absent value is a missing log; a negative upstream value is invalid intake evidence and is labelled and counted as an exclusion. A range rate, if shown, is the sum of valid logged ingested grams divided by the sum of the same valid moving durations and states eligible/total sessions and every exclusion. It never uses carbs used, calories, load, wellness totals, or targets as either operand.
- **Dates and availability:** Anchor windows to athlete-local dates, preserve each activity timezone, label current-day `_meta.as_of` evidence partial, and surface stale wellness, provenance/field-semantics warnings, missing fields, unavailable/Strava-blocked rows, and missing/invalid durations. Missing is neither zero nor inadequate fueling. An absent requested race is reported as unconfirmed calendar context, never invented.
- **Boundaries and ownership:** The prompt is read-only: no write/delete tools, `include_full`, streams, or raw payloads. It does not diagnose, assess eating disorders, prescribe individual nutrition, calculate/recommend carbohydrate/calorie/sodium/fluid/sweat targets, infer deficits, claim product/performance effects, or invent food/product libraries. General material is visibly educational/conditional and refers individual or medical requests to a qualified sports dietitian/clinician. Step 1 owns the contract, portable pack draft, `internal/prompts/fueling_review_test.go`, and `TestFuelingReviewPortablePackContract`, which reads the pack and asserts the exact tools/arguments, `include_unnamed`, pagination/event truncation, nutrition-only wellness fields/aliases, field distinctions, absent/zero/negative intake and positive-moving-time formula/coverage exclusions, and read-only/health boundaries. Step 2 extends that file with the prompt function, registry, and deterministic golden-fixture tests; `docs/prompts/README.md` will list the new portable pack.
| 2026-07-10 18:29 | Review R001 | plan Step 1: REVISE |
| 2026-07-10 18:32 | Review R002 | plan Step 1: REVISE |
| 2026-07-10 18:34 | Review R003 | plan Step 1: REVISE |
| 2026-07-10 18:38 | Review R004 | plan Step 1: REVISE |
| 2026-07-10 18:41 | Review R005 | plan Step 1: REVISE |
| 2026-07-10 18:44 | Review R006 | plan Step 1: APPROVE |
| 2026-07-10 18:50 | Review R007 | plan Step 2: REVISE |
| 2026-07-10 18:52 | Review R008 | plan Step 2: APPROVE |

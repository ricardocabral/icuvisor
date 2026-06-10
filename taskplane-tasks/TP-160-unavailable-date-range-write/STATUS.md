# TP-160: Unavailable date-range write convenience — Status

**Current Step:** Step 1: Design the range-write contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-10
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied

---

### Step 1: Design the range-write contract
**Status:** 🟨 In Progress

- [x] Public surface selected
- [x] Allowed unavailability categories defined
- [x] Idempotency/range semantics defined
- [ ] Initial targeted event tests added/run
- [x] R001 concrete contract captured for API surface, response shape, idempotency, catalog, and tests

---

### Step 2: Implement the write convenience and catalog integration
**Status:** ⬜ Not Started

> ⚠️ Hydrate: Expand after inspecting current event write helpers, category normalization, and registry conventions.

- [ ] Range write implemented with validation
- [ ] Tool/schema/catalog integration complete
- [ ] Safety protections preserved
- [ ] Targeted tools/MCP/toolcheck tests pass

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Integration tests (if applicable)
- [ ] All failures fixed
- [ ] Build passes

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `README.md` updated
- [ ] PRD updated if catalog changed
- [ ] `CHANGELOG.md` updated
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | Step 1 | REVISE | `.reviews/R001-plan-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-09 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-10 11:54 | Task started | Runtime V2 lane-runner execution |
| 2026-06-10 11:54 | Step 0 started | Preflight |
| 2026-06-10 11:58 | Review R001 | plan Step 1: REVISE |

---

## Blockers

*None*

---

## Notes

Public signal: IntervalCoach forum #856 added Sick/Injured/Holiday range support.

R001 plan feedback requires a concrete contract before implementation.

### Step 1 Design Contract

**Public API surface:** add a dedicated core write tool named `add_unavailable_date_range`, rather than extending `add_or_update_event`. Rationale: the range-write contract is intentionally narrower than generic event writes, easier for LLMs to discover, and avoids adding range semantics to a single-date tool. Catalog integration in Step 2 must add the name to `internal/toolcatalog` known + athlete-scoped lists, register it near event writes in `registryBaseTools`, mark it `RequirementWrite`, use `coreTool`, include it in coach per-athlete ACL eligibility, schema snapshots/hash/doc catalog surfaces, README, PRD, and generated/toolcheck expectations.

**Category contract:** accept a closed unavailability set only: `HOLIDAY`, `SICK`, `INJURED`. Normalize by trimming whitespace, uppercasing, and accepting explicit aliases `HOLIDAYS`, `VACATION`, `TIME_OFF`, `TIME OFF`, `TRAVEL`, and `AWAY` as `HOLIDAY`; `ILL`, `ILLNESS`, and `SICKNESS` as `SICK`; `INJURY` as `INJURED`. Reject every other category, including `NOTE`, `WORKOUT`, race, fitness, and custom categories, with a short actionable user error. The generic `add_or_update_event` remains pass-through and unchanged.

**Request/response and idempotency contract:** request fields are required `start_date`, `end_date`, `category`; optional `name`, `description`, and `include_full`. Dates are athlete-local inclusive `YYYY-MM-DD`; same-day ranges are allowed; reversed ranges and ranges over 31 inclusive days are rejected. Implementation creates one upstream event per day because `intervals.WriteEventParams` writes a single `Date`. Each write uses normalized `Category`, `Type: "Unavailable"`, default `Name` equal to `Holiday`, `Sick`, or `Injured` when omitted, optional `Description`, and a generated stable `external_id` with prefix `icuvisor-unavailable-v1-` plus a short hash of normalized category/date/name. Repeated identical calls skip existing matching `external_id` rows; pre-existing same-day rows that exactly match writable fields also skip; mixed ranges create missing days and report skipped days. Nonmatching same-day events, including workouts, are not overwritten or deleted; they are reported as conflicts/warnings while the unavailable marker is still added unless it is a duplicate. Response shape: `{events:[terse event rows], status:"created", _meta:{operation:"create_range", date_range:{oldest,newest}, timezone, category, requested_days, created_count, skipped_count, conflict_count, range_cap_days:31, include_full, skipped:[{date,event_id,reason}], same_day_conflicts:[...]}}`; `include_full:true` only adds raw upstream payloads under event rows as existing event tools do.

**Initial test contract:** Step 1 failing tests should cover valid multi-day creation and write params, alias normalization, repeated-call idempotency, mixed existing/missing days, invalid/reversed/excessive ranges, unsupported categories, include_full shaping, and safety invariants that the tool is `RequirementWrite`/core and does not require delete mode or delete events. Step 2 targeted tests must include `go test ./internal/tools ./internal/mcp ./internal/toolchecks` after schema/catalog updates.

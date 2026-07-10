# TP-228: Align sport-settings update and apply requests with live OpenAPI — Status

**Current Step:** Step 2: Align the intervals.icu client
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 3
**Review Counter:** 6
**Iteration:** 1
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Live public OpenAPI contract reconfirmed

---

### Step 1: Define the corrected write boundary

**Status:** ✅ Complete

- [x] Exact update/apply request contract documented
- [x] Public schema migration for effective_date decided
- [x] Implicit historical apply removed from design
- [x] Truthful response metadata defined
- [x] R001: Presence-aware `recalc_hr_zones` defaults to true while forwarding explicit false
- [x] R001: Regression plan covers default/false forwarding, no implicit apply, no-query/no-body apply, and legacy argument rejection

---

### Step 2: Align the intervals.icu client

**Status:** ✅ Complete

- [x] Required recalcHrZones query implemented
- [x] Apply PUT sends no date or semantic body
- [x] Update no longer invokes apply
- [x] Client contract tests pass
- [x] R004: Use retry-safe body-plus-query and bodyless PUT transport helpers without changing existing callers
- [x] R004: Exact wire coverage verifies true/false query, no implicit apply, and a zero-byte bodyless apply request

---

### Step 3: Align the MCP schema and response

**Status:** ⬜ Not Started

- [ ] recalc_hr_zones schema and forwarding implemented
- [ ] Unsupported effective-date behavior removed
- [ ] Response metadata corrected
- [ ] Schema snapshots and generated data updated

---

### Step 4: Regression coverage

**Status:** ⬜ Not Started

- [ ] Exact update/apply wire tests added
- [ ] No-implicit-apply regression added
- [ ] Legacy invalid argument behavior covered

---

### Step 5: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Race suite passing
- [ ] Lint passing
- [ ] Build passes
- [ ] Generated docs clean

---

### Step 6: Documentation & Delivery

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
| 2026-07-10 11:36 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 11:36 | Step 0 started | Preflight |

## Blockers

*None*

## Notes

- R001 plan review: legacy `effective_date` must be rejected by strict decoding before an upstream request; response metadata may only report the requested HR-zone recalculation boolean.
- R004 plan review: client resolves no input defaults; it encodes a resolved `RecalcHRZones` bool and uses body-plus-query update and bodyless-PUT apply transports that preserve retries and 422 handling.
| 2026-07-10 11:40 | Review R001 | plan Step 1: REVISE |
| 2026-07-10 11:42 | Review R002 | plan Step 1: APPROVE |
| 2026-07-10 11:45 | Review R003 | code Step 1: APPROVE |
| 2026-07-10 11:47 | Review R004 | plan Step 2: REVISE |
| 2026-07-10 11:50 | Review R005 | plan Step 2: APPROVE |
| 2026-07-10 11:55 | Review R006 | code Step 2: APPROVE |

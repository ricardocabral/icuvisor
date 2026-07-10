# TP-228: Align sport-settings update and apply requests with live OpenAPI — Status

**Current Step:** Step 3: Align the MCP schema and response
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 3
**Review Counter:** 9
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

**Status:** 🟨 In Progress

- [ ] recalc_hr_zones schema and forwarding implemented
- [ ] Unsupported effective-date behavior removed
- [ ] Response metadata corrected
- [ ] Schema snapshots and generated data updated
- [ ] R007: Preserve explicit-false/default-true decoding and always emit truthful recalculation-requested metadata
- [ ] R007: Narrowly approve only update_sport_settings effective_date removal without weakening generic schema-removal protection
- [ ] R009: Add production schema-stability policy keyed by tool/property and test approved versus unrelated removals through CheckSchemaStability

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
- R007 plan review: schema stability needs a TP-228-only approved effective_date-removal exception; generic property-removal detection remains enforced.
- Step 3 implementation plan: (1) replace `EffectiveDate` with `*bool RecalcHRZones` in the tool request; decoding resolves nil to true and passes the resulting bool to `WriteSportSettingsParams.RecalcHRZones`; remove `EffectiveDate` from internal params. (2) Delete `effective_date` from the strict type, validation, public strings, schema requirements/properties/examples, metadata, and tests, relying on `DecodeStrict` to reject it before profile/writer calls. (3) Replace date/recompute metadata with always-emitted `hr_zone_recalculation_requested` equal to the resolved option. (4) Make sport the sole required schema field, add `recalc_hr_zones` as optional boolean default true with an LLM-readable description, then regenerate snapshot and website catalogs with `make docs-tools`. (5) Add table-driven omitted/default-true and explicit-false forwarding/metadata tests plus legacy-date/unknown rejection with zero client calls. (6) Add production `approvedSchemaPropertyRemovals` keyed by tool/property in schema_stability.go; `compareStableSchema` consults it solely before a property-removed failure. Its sole durable TP-228 rationale approves `update_sport_settings.effective_date`; CheckSchemaStability tests prove a different property and another tool's effective_date still fail. Run `go test ./internal/tools ./internal/toolchecks`.
| 2026-07-10 11:40 | Review R001 | plan Step 1: REVISE |
| 2026-07-10 11:42 | Review R002 | plan Step 1: APPROVE |
| 2026-07-10 11:45 | Review R003 | code Step 1: APPROVE |
| 2026-07-10 11:47 | Review R004 | plan Step 2: REVISE |
| 2026-07-10 11:50 | Review R005 | plan Step 2: APPROVE |
| 2026-07-10 11:55 | Review R006 | code Step 2: APPROVE |
| 2026-07-10 11:57 | Review R007 | plan Step 3: REVISE |
| 2026-07-10 11:58 | Review R008 | plan Step 3: REVISE |
| 2026-07-10 12:00 | Review R009 | plan Step 3: REVISE |

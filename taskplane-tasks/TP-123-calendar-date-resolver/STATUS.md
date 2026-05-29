# TP-123: Calendar date resolver and future date anchors — Status

**Current Step:** Step 2: Implement date anchors and tests
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 5
**Iteration:** 2
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Design deterministic date surface
**Status:** ✅ Complete

- [x] Inspect existing `_meta.as_of`, `get_today`, `get_activities`, `get_events`, and prompt guidance for date anchors.
- [x] Decide whether to add a small read-only tool such as `resolve_calendar_dates` or to harden existing date metadata/prompts without a new tool.
- [x] Document the chosen surface and non-goals in STATUS.md Discoveries, including why it avoids model date arithmetic.
- [x] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`

---

### Step 2: Implement date anchors and tests
**Status:** 🟨 In Progress

- [x] Implement `resolve_calendar_dates` as a strict read-only core tool with optional `base_date` (YYYY-MM-DD, default from injected clock converted to athlete timezone), optional `offsets` (default `[0]`, unique integers, max 32 items, each between -366 and 366), athlete-local `AddDate(0,0,offset)` arithmetic, strict `additionalProperties:false`, response rows containing `offset_days`, `date`, `weekday`, and `_meta` containing `timezone`, `base_date`, `base_weekday`, `server_version`, and `count`.
- [x] Register the public tool across `registryBaseTools`, `toolCatalogGroup`, `internal/toolcatalog`, and schema stability surfaces so it is core and athlete-scoped.
- [x] Add tests covering current day, future day offsets, base_date parsing, DST/timezone boundaries, invalid input, registration metadata, and catalog membership.
- [x] Update catalog/schema snapshots if the public tool surface changes.
- [x] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`
- [ ] Update stale public catalog guard surfaces for `resolve_calendar_dates` so full-suite catalog/safety tests pass.
- [ ] Return an athlete-timezone-specific error for timezone load failures instead of the invalid-arguments message.

---

### Step 3: Add activation guidance and eval coverage
**Status:** 🟨 In Progress

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
| Existing `get_today`, `get_activities`, and `get_events` expose current-day `_meta.as_of*` only when fetching data, but planning prompts need deterministic future anchors without requiring unrelated activity/event reads or model date arithmetic. | Add a small read-only `resolve_calendar_dates` tool that uses athlete timezone, an optional `base_date`, and integer offsets to return exact local dates/weekdays; keep non-goals limited to no calendar writes, no event inference, and no UTC/client-time inference. | Step 1 design |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 13:19 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 13:19 | Step 0 started | Preflight |
| 2026-05-29 13:43 | Worker iter 1 | done in 1422s, tools: 99 |
| 2026-05-29 13:43 | Step 3 started | Add activation guidance and eval coverage |

---

## Blockers

*None*

---

## Notes

Plan review R002 requested explicit registration/catalog surfaces, an athlete-local `AddDate` arithmetic contract (not UTC/24h duration math), and a pinned strict input/output schema before coding.
Plan review R003 further pinned the default-base injected clock conversion, exact `offsets` defaults/bounds/uniqueness, strict `additionalProperties:false`, row fields (`offset_days`, `date`, `weekday`), and `_meta` fields (`timezone`, `base_date`, `base_weekday`, `server_version`, `count`).

| 2026-05-29 13:24 | Review R001 | code Step 1: APPROVE |
| 2026-05-29 13:27 | Review R002 | plan Step 2: REVISE |
| 2026-05-29 13:29 | Review R003 | plan Step 2: REVISE |
| 2026-05-29 13:31 | Review R004 | plan Step 2: APPROVE |
| 2026-05-29 13:43 | Review R005 | code Step 2: REVISE |

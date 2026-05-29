# TP-123: Calendar date resolver and future date anchors — Status

**Current Step:** Step 5: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 11
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
**Status:** ✅ Complete

- [x] Implement `resolve_calendar_dates` as a strict read-only core tool with optional `base_date` (YYYY-MM-DD, default from injected clock converted to athlete timezone), optional `offsets` (default `[0]`, unique integers, max 32 items, each between -366 and 366), athlete-local `AddDate(0,0,offset)` arithmetic, strict `additionalProperties:false`, response rows containing `offset_days`, `date`, `weekday`, and `_meta` containing `timezone`, `base_date`, `base_weekday`, `server_version`, and `count`.
- [x] Register the public tool across `registryBaseTools`, `toolCatalogGroup`, `internal/toolcatalog`, and schema stability surfaces so it is core and athlete-scoped.
- [x] Add tests covering current day, future day offsets, base_date parsing, DST/timezone boundaries, invalid input, registration metadata, and catalog membership.
- [x] Update catalog/schema snapshots if the public tool surface changes.
- [x] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`
- [x] Update stale public catalog guard surfaces for `resolve_calendar_dates` so full-suite catalog/safety tests pass.
- [x] Return an athlete-timezone-specific error for timezone load failures instead of the invalid-arguments message.

---

### Step 3: Add activation guidance and eval coverage
**Status:** ✅ Complete

- [x] Update existing relative-future cookbook scenarios `CB-PLAN-02` and `CB-TAPER-01` so `resolve_calendar_dates` is an expected tool and must use returned athlete-local `date` + `weekday` values.
- [x] Add an eval scenario for a known-bad weekday/date pairing such as “Monday May 26” with `resolve_calendar_dates` expected and model-arithmetic/UTC/client-time anti-patterns.
- [x] Update Claude Project guidance to name `resolve_calendar_dates`, athlete-local offsets, and the requirement not to infer dates from UTC/client time or model arithmetic.
- [x] Regenerate generated public tool catalog/reference docs before eval validation: `make docs-tools`.
- [x] Run targeted tests: `make eval-validate`

---

### Step 4: Testing & Verification
**Status:** ✅ Complete

- [x] FULL test suite passing: `make test`
- [x] Lint passes or pre-existing linter limitations are documented: `make lint`
- [x] Build passes: `make build`
- [x] All failures fixed or clearly documented as pre-existing

---

### Step 5: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Existing `get_today`, `get_activities`, and `get_events` expose current-day `_meta.as_of*` only when fetching data, but planning prompts need deterministic future anchors without requiring unrelated activity/event reads or model date arithmetic. | Add a small read-only `resolve_calendar_dates` tool that uses athlete timezone, an optional `base_date`, and integer offsets to return exact local dates/weekdays; keep non-goals limited to no calendar writes, no event inference, and no UTC/client-time inference. | Step 1 design |
| Eval validation reads `web/data/tools.json`, not the Go registry directly, so scenarios expecting a newly added public tool fail until `make docs-tools` refreshes the generated website catalog. | Regenerated `web/data/tools.json` during Step 3 and kept `web/content/reference/tools.md` on its generated shortcode surface. | Step 3 eval/doc sync |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 13:19 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 13:19 | Step 0 started | Preflight |
| 2026-05-29 13:43 | Worker iter 1 | done in 1422s, tools: 99 |
| 2026-05-29 13:43 | Step 3 started | Add activation guidance and eval coverage |
| 2026-05-29 13:45 | Exit intercept reprompt | Supervisor provided instructions (670 chars) — reprompting worker |
| 2026-05-29 14:06 | Worker iter 2 | done in 1429s, tools: 105 |
| 2026-05-29 14:06 | Task complete | .DONE created |

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
| 2026-05-29 13:48 | Review R006 | code Step 2: APPROVE |
| 2026-05-29 13:51 | Review R007 | plan Step 3: REVISE |
| 2026-05-29 13:53 | Review R008 | plan Step 3: APPROVE |
| 2026-05-29 13:59 | Review R009 | code Step 3: APPROVE |
| 2026-05-29 14:01 | Review R010 | plan Step 4: APPROVE |
| 2026-05-29 14:04 | Review R011 | code Step 4: APPROVE |

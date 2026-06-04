# TP-154: Generated per-tool schema docs — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-06-04
**Review Level:** 2
**Review Counter:** 12
**Iteration:** 5
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Current docs-tools behavior identified

---

### Step 1: Design generated docs data shape
**Status:** ✅ Complete

- [x] R001 plan feedback addressed in concrete design notes
- [x] Data-file strategy decided
- [x] Concise schema projection defined
- [x] Nested schema/example handling decided
- [x] No-secret/no-path boundary confirmed

---

### Step 2: Implement generator and tests
**Status:** ✅ Complete

- [x] Generator emits per-tool schema/reference data
- [x] Deterministic checks/tests added
- [x] No-network stable generation preserved
- [x] Targeted tests passing

---

### Step 3: Render docs and refine UX
**Status:** ✅ Complete

- [x] Reference rendering updated
- [x] Write-tool examples visible or linked/expandable
- [x] Page readability checked
- [x] Docs generation/build checked or caveat recorded

---

### Step 4: Update contributor/user guidance
**Status:** ✅ Complete

- [x] Generated-file workflow/help docs mention both `web/data/tools.json` and `web/data/tool_schemas.json`, or rationale recorded for no contributor-doc change
- [x] Stale-generation guard covers `web/data/tool_schemas.json`, or caveat/follow-up recorded
- [x] CHANGELOG updated
- [x] Internal-only exposure reviewed by comparing schema keys with catalog names
- [x] TP-153 caveat/dependency recorded: docs generation projects from live registry, not schema snapshots

---

### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Lint passes
- [x] All failures fixed
- [x] Build passes

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Generated-file summary included

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Website schema docs generation projects from the live registered catalog, not from stable schema snapshot fixtures. | Recorded as a TP-153 dependency/caveat; no generator dependency on snapshots added. | `cmd/gendocs`, STATUS Step 4 notes |
| Contributor schema snapshot instructions are separate from website docs data generation. | Left `CONTRIBUTING.md` unchanged; updated `README.md`/`web/README.md` for `make docs-tools`. | `CONTRIBUTING.md`, `README.md`, `web/README.md` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-03 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-03 23:08 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 23:08 | Step 0 started | Preflight |
| 2026-06-03 | Step 0 completed | `make docs-tools` regenerates summary-only `web/data/tools.json` via `cmd/gendocs --out`; no extra services/dependencies required |
| 2026-06-03 | Step 1 started | Design generated docs data shape |
| 2026-06-03 | Step 1 plan reviewed | R001 requested concrete design notes; R002 approved updated plan |
| 2026-06-03 | Step 1 code reviewed | Approved |
| 2026-06-03 | Step 2 started | Implement generator and tests |
| 2026-06-03 | Step 3 started | Render docs and refine UX |
| 2026-06-03 | Step 4 started | Update contributor/user guidance |
| 2026-06-03 | Step 5 started | Testing & Verification |
| 2026-06-03 23:24 | Worker iter 1 | done in 987s, tools: 57 |
| 2026-06-03 23:52 | Worker iter 2 | done in 1655s, tools: 127 |
| 2026-06-04 00:08 | Worker iter 3 | done in 964s, tools: 10 |
| 2026-06-04 00:30 | Worker iter 4 | done in 1319s, tools: 13 |
| 2026-06-04 00:30 | Step 6 started | Documentation & Delivery |
| 2026-06-04 | Step 6 completed | Delivery notes, docs review, discoveries, and generated-file summary recorded |
| 2026-06-04 00:33 | Worker iter 5 | done in 158s, tools: 26 |
| 2026-06-04 00:33 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

### Step 6 delivery notes

- Must Update docs verified: `web/content/reference/tools.md` now describes generated arguments/examples, `web/layouts/partials/tool-catalog.html` renders `web/data/tool_schemas.json` argument details and input examples, and `CHANGELOG.md` includes the `[Unreleased]` docs/reference improvement.
- Check If Affected docs reviewed: `README.md` and `web/README.md` now mention `make docs-tools`/generated `tool_schemas.json`; `CONTRIBUTING.md` remains correct because its schema snapshot instructions are a separate live-registry compatibility workflow.
- Generated-file summary: `web/data/tools.json` has 60 catalog rows and `web/data/tool_schemas.json` has 60 matching schema entries, 260 total projected arguments, and curated examples for 9 write-oriented tools; no missing or extra schema keys were found versus catalog tool names.

### Step 4 guidance notes

- `README.md`, `web/README.md`, and `Makefile` now describe `make docs-tools` as regenerating both website catalog and schema data. `CONTRIBUTING.md` remains unchanged because its schema-snapshot workflow (`scripts/snapshot_tool_schemas.go`) is separate from generated website docs data.
- Internal exposure check: `web/data/tools.json` names and `web/data/tool_schemas.json` keys both contain 60 tools with no missing or extra schema entries.
- TP-153 caveat: generated website docs project directly from the live registered catalog via `tools.SchemaCatalog()`, not from `internal/tools/schema_snapshot` files. Future snapshot-policy work should not make docs generation depend on stale snapshots without a reviewed design change.

### Step 1 design notes

- **Generator target:** extend `cmd/gendocs` rather than adding a separate scripts generator. `make docs-tools` should continue to be the single workflow and will call `go run ./cmd/gendocs --out web/data/tools.json`, with a new optional/default schema-data output path written by the same command. Update `cmd/gendocs/main_test.go` golden coverage rather than moving generation to `scripts/*`.
- **Data-file strategy:** add separate generated `web/data/tool_schemas.json` keyed by tool name. Keep `web/data/tools.json` summary-only for the existing table and use the new data file only for per-tool argument/example details. This avoids bloating the summary catalog and keeps Hugo lookups straightforward.
- **Projection contract:** each tool entry should include `name`, `description`, `arguments`, and `examples`. Each argument includes `name`, `required`, `type`, `description`, optional `enum`, optional `default`, optional `format`, optional `items`, optional `properties`/`children` summary for object fields, and optional `additional_properties`. Types are normalized from JSON Schema primitives; arrays summarize item type; `anyOf`/`oneOf`/nullable are represented as joined type labels (for example `string | null`) without copying raw branch schemas.
- **Nested/large schemas:** include only one nested level of child summaries for object-heavy fields. `workout_doc` should show its top-level object fields and required flags, but not fully recurse through every nested interval step. Custom item `content` and other free-form/additionalProperties-heavy objects should show the field description plus an `additional_properties`/summary marker rather than dumping arbitrary JSON schemas.
- **Examples policy:** derive examples from schema `input_examples` (falling back to `examples` only if the mirror exists without `input_examples`). Preserve examples in generated JSON for write tools, capped to a small deterministic count (2 examples per tool) and intended for Hugo `<details>` rendering so the table remains readable. Do not synthesize examples or include runtime values.
- **Safety/determinism:** generate from the full registered catalog with full delete capability, coach mode enabled, no HTTP calls, sorted tools/arguments, stable JSON indentation, and no local paths or secrets. Placeholder IDs already present in catalog examples such as `12345`, `event-123`, or `folder-abc` are acceptable only as synthetic examples; no configured athlete IDs/API keys should be read or emitted.


| 2026-06-03 23:10 | Review R001 | plan Step 1: REVISE |
| 2026-06-03 23:12 | Review R002 | plan Step 1: APPROVE |
| 2026-06-03 23:13 | Review R003 | code Step 1: APPROVE |
| 2026-06-03 23:14 | Review R004 | plan Step 2: APPROVE |
| 2026-06-03 23:33 | Review R005 | code Step 2: APPROVE |
| 2026-06-03 23:36 | Review R006 | plan Step 3: APPROVE |
| 2026-06-03 23:41 | Review R007 | code Step 3: APPROVE |
| 2026-06-03 23:43 | Review R008 | plan Step 4: REVISE |
| 2026-06-03 23:44 | Review R009 | plan Step 4: APPROVE |
| 2026-06-03 23:48 | Review R010 | code Step 4: APPROVE |
| 2026-06-03 23:49 | Review R011 | plan Step 5: APPROVE |
| 2026-06-04 00:10 | Review R012 | code Step 5: APPROVE |

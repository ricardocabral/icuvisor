# TP-118: Activity tombstone delete endpoint — Status

**Current Step:** Step 3: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 6
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Clean-room guardrail confirmed

---

### Step 1: Determine the correct activity deletion contract
**Status:** ✅ Complete

- [x] Existing delete implementation and tests inspected
- [x] Public upstream evidence checked without competitor source
- [x] Endpoint decision recorded in Discoveries
- [x] `/api/v1` base-path handling captured for the selected endpoint
- [x] Targeted tests run with regex covering intervals path and target-athlete safety tests

---

### Step 2: Implement and lock the endpoint behavior
**Status:** ✅ Complete

- [x] Client path/method updated if needed
- [x] httptest coverage asserts exact method/path and target-athlete safety
- [x] Tool metadata/source-endpoint response updated and asserted if affected
- [x] Tool schema snapshot added or updated if affected
- [x] Targeted tests run with selector covering `DeleteMethods|ActivityIDEndpointsRequireResolvedTargetOwnership|DeleteTools|delete_activity|Schema`

---

### Step 3: Testing & Verification
**Status:** 🟨 In Progress

- [ ] FULL test suite passing
- [ ] Lint passing if source changed
- [ ] All failures fixed
- [ ] Build passes

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `CHANGELOG.md` updated if needed
- [ ] Roadmap/PRD checked if affected
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | Step 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | Step 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | Code | Step 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | Plan | Step 2 | REVISE | `.reviews/R004-plan-step2.md` |
| R005 | Plan | Step 2 | APPROVE | `.reviews/R005-plan-step2.md` |
| R006 | Code | Step 2 | APPROVE | `.reviews/R006-code-step2.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Repository search found no local OpenAPI file and no existing tombstone fixture; only icuvisor docs/tests and the TP-118 prompt's newly observed public OpenAPI path mention activity deletion/tombstone. | Use clean-room public API signal from task prompt plus existing tests; no competitor source opened. | `PROMPT.md`, repo grep for `tombstone`/delete activity |
| `DeleteActivity` should issue `DELETE /activity/{id}/tombstone` because the newly observed public Intervals.icu OpenAPI path is more specific for activity deletion than the existing direct `/activity/{id}` path. No fallback is planned because destructive retries against multiple endpoints would broaden deletion semantics without documented need. | Implement in Step 2 and lock with exact-path httptest coverage. | `internal/intervals/delete.go`, `internal/intervals/delete_test.go` |
| The observed upstream path includes `/api/v1`, but `config.DefaultAPIBaseURL` and test base URLs already represent the API root; client calls must pass relative path parts `activity`, `{id}`, `tombstone`, yielding `/activity/{id}/tombstone` in httptest and `/api/v1/activity/{id}/tombstone` against the default base URL. | Assert the relative request path in Step 2; do not duplicate `/api/v1`. | `internal/intervals/client.go`, `internal/intervals/delete_test.go` |
| `delete_activity` input schema was not previously included in schema snapshots; adding it required adding the tool to the schema-stability allowlist. Running the snapshot generator also showed unrelated pre-existing drift in `add_or_update_event.json`, which was not included in this task's Step 2 changes. | Keep the focused delete_activity snapshot; revisit unrelated snapshot drift separately if the schema-stability command is part of a later gate. | `internal/toolchecks/schema_stability.go`, `internal/tools/schema_snapshot/delete_activity.json` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 15:21 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 15:21 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

Plan review R001 requires Step 1 tests to include `DeleteMethods|ActivityIDEndpointsRequireResolvedTargetOwnership` coverage and discoveries to explicitly note `/api/v1` base URL handling.
Plan review R004 requires Step 2 tests to include intervals endpoint/safety selectors and requires locking `delete_activity` `_meta.source_endpoint` if the endpoint changes.
| 2026-05-29 15:25 | Review R001 | plan Step 1: REVISE |
| 2026-05-29 15:26 | Review R002 | plan Step 1: APPROVE |
| 2026-05-29 15:29 | Review R003 | code Step 1: APPROVE |
| 2026-05-29 15:32 | Review R004 | plan Step 2: REVISE |
| 2026-05-29 15:33 | Review R005 | plan Step 2: APPROVE |
| 2026-05-29 15:40 | Review R006 | code Step 2: APPROVE |

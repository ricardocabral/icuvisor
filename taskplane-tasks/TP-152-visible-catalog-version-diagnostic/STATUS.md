# TP-152: Visible catalog/version diagnostic tool — Status

**Current Step:** Step 4: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 11
**Iteration:** 2
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Current catalog metadata tests/docs confirmed

---

### Step 1: Design diagnostic contract
**Status:** ✅ Complete

- [x] Stable tool name chosen
- [x] Response shape defined
- [x] Description baseline strategy decided
- [x] No-secret/no-athlete boundary confirmed
- [x] Description catalog fingerprint contract defined
- [x] Mismatch/status semantics clarified
- [x] Same-version fingerprint drift test plan captured

---

### Step 2: Implement tool and tests
**Status:** ✅ Complete

- [x] Core read-only diagnostic registered
- [x] Output/no-leak tests added
- [x] Catalog/hash tests updated
- [x] No-network behavior confirmed
- [x] Targeted tests passing
- [x] Live catalog hash source and no-arg handler plan documented
- [x] Fingerprint helper/package boundary and self-reference normalization documented
- [x] Effective catalog/registration order and shared catalog membership documented
- [x] Step 2 test coverage plan documented

---

### Step 3: Update generated docs and stale-catalog guidance
**Status:** ✅ Complete

- [x] Tool docs/catalog regenerated
- [x] Generated catalog data verified for the meta-tool entry
- [x] Upgrade/troubleshooting docs updated
- [x] Exact diagnostic comparison fields documented without comparing fingerprint to live catalog hash
- [x] Reconnect versus new-conversation actions documented
- [x] CHANGELOG updated
- [x] Privacy wording verified

---

### Step 4: Testing & Verification
**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Lint passes
- [x] All failures fixed
- [x] Build passes

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Client-specific caveats noted

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | Step 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | Step 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | Code | Step 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | Plan | Step 2 | REVISE | `.reviews/R004-plan-step2.md` |
| R005 | Plan | Step 2 | APPROVE | `.reviews/R005-plan-step2.md` |
| R006 | Code | Step 2 | UNAVAILABLE | n/a |
| R007 | Plan | Step 3 | REVISE | `.reviews/R007-plan-step3.md` |
| R008 | Plan | Step 3 | APPROVE | `.reviews/R008-plan-step3.md` |
| R009 | Code | Step 3 | APPROVE | `.reviews/R009-code-step3.md` |
| R010 | Plan | Step 4 | APPROVE | `.reviews/R010-plan-step4.md` |
| R011 | Code | Step 4 | APPROVE | `.reviews/R011-code-step4.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-03 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-03 22:57 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 22:57 | Step 0 started | Preflight |
| 2026-06-03 23:24 | Worker iter 1 | done in 1623s, tools: 90 |

---

## Blockers

*None*

---

## Notes

### Step 1 design plan
- Tool name: `icuvisor_check_server_version`; no conflict with existing `toolcatalog` names and the `icuvisor_` prefix matches the existing meta-tool namespace.
- Response shape: no-argument read-only response with top-level `server_version`, `catalog_hash`, `description_server_version`, `description_catalog_fingerprint`, `toolset`, `delete_mode`, `status`, and `action`; `_meta` repeats non-secret diagnostic source fields only.
- Description baseline: generate the tool description at registration with visible fields `description_server_version=<version>`, `description_catalog_fingerprint=<fingerprint>`, `description_toolset=<toolset>`, and `description_delete_mode=<delete_mode>`. The fingerprint is a deterministic SHA-256 over the active catalog records that pass known delete-mode/toolset gates with the diagnostic tool's fingerprint token normalized to a stable sentinel before injecting the final token. The response returns both the live runtime `catalog_hash` and the comparable `description_catalog_fingerprint`; assistants compare visible description fields to response fields when clients hide `_meta`.
- Mismatch semantics: the tool does not claim the server can observe stale client state. It always returns `status: "compare_visible_description"` plus an `action` telling the assistant to reconnect/start a new conversation if the visible description fields differ from the response fields, or if `_meta.schema_changed` is visible.
- Test plan: add same-version drift coverage showing `description_catalog_fingerprint` changes when a catalog description/schema changes even when `server_version` is unchanged.
- Privacy boundary: the tool has no intervals client dependency, no arguments, and returns no API key, athlete ID, filesystem path, username, raw env value, or network-derived data.
- R002 implementation notes: runtime `catalog_hash` must come from metadata after `NewServer` computes it; keep fingerprint helper out of an `internal/tools` <-> `internal/mcp` import cycle; mirror visible description fields in response with unambiguous names; document/test any coach dynamic visibility limitation.

### Step 2 implementation plan
- Add `response.RuntimeCatalogMetadata()` as a locked getter returning normalized current version/hash; the diagnostic handler calls this at request time for live `catalog_hash`, never a registration placeholder.
- Implement `internal/tools/check_server_version.go` with no-argument validation, `newCheckServerVersionTool(version, descriptionFingerprint, deleteMode, toolset, shaping...)`, visible description fields, and a response containing `server_version`, `catalog_hash`, `description_server_version`, `description_catalog_fingerprint`, `toolset`, `delete_mode`, `status: compare_visible_description`, `action`, and `_meta` source details.
- Keep the fingerprint helper in `internal/tools` to avoid a tools<->mcp cycle. It hashes the effective `Tool` records (name, description, input schema, output schema) with the diagnostic description's injected fingerprint value normalized to a sentinel token before hashing.
- Register base tools first, then `icuvisor_list_advanced_capabilities`, then compute/register `icuvisor_check_server_version`; fingerprint input includes tools that pass known capability/delete-mode and toolset gates plus the diagnostic tool. Coach per-athlete dynamic ACL filtering is not included in this static description fingerprint; tests/docs will state it is a catalog-mode fingerprint, while live `catalog_hash` remains authoritative for the server's exposed catalog.
- Update `internal/toolcatalog/catalog.go` with `ICUvisorCheckServerVersion`, include it in `allToolNames`, keep it out of `athleteScopedToolNames`, add it to the `meta` group in `internal/tools/catalog.go`, and update tier/catalog expectations.
- Tests: output shape and no-leak/no-network handler tests; same-version fingerprint drift on description/schema changes; catalog hash sensitivity to diagnostic description/schema; shared catalog membership/descriptor group tests; targeted `go test ./internal/tools ./internal/toolcatalog ./internal/mcp ./internal/response -run 'Check|Catalog|Schema|Advanced'`.

### Step 3 docs plan
- Run `make docs-tools`, then verify `web/data/tools.json` contains `icuvisor_check_server_version` in the `meta` group; treat `web/content/reference/tools.md` as generated-shortcode surrounding prose only and do not hand-edit unless needed.
- Update `web/content/guides/after-upgrade.md` with both paths: visible `_meta.schema_changed` means start a new conversation, and clients that hide `_meta` can call `icuvisor_check_server_version` after reconnect/reload to compare visible description baselines.
- Update `web/content/guides/troubleshooting.md` stale-catalog steps with exact field comparisons: visible `description_server_version`, `description_catalog_fingerprint`, `description_toolset`, and `description_delete_mode` versus response `server_version`, `description_catalog_fingerprint`, `toolset`, and `delete_mode`; do not compare `description_catalog_fingerprint` with live `catalog_hash`.
- Document actions clearly: reconnect/reload MCP tools when visible description fields differ from the diagnostic response, and start a new conversation when `_meta.schema_changed` is visible or stale schemas/context persist after reconnecting.
- Add a `[Unreleased]` changelog entry under `Added` for the local, read-only diagnostic and verify wording avoids telemetry, cloud checks, credential upload, filesystem inspection, intervals.icu API calls, athlete-data exposure, or claims that the server automatically detects stale client state.
| 2026-06-03 23:02 | Review R001 | plan Step 1: REVISE |
| 2026-06-03 23:06 | Review R002 | plan Step 1: APPROVE |
| 2026-06-03 23:07 | Review R003 | code Step 1: APPROVE |
| 2026-06-03 23:08 | Review R004 | plan Step 2: REVISE |
| 2026-06-03 23:11 | Review R005 | plan Step 2: APPROVE |
| 2026-06-03 23:32 | Review R006 | code Step 2: UNAVAILABLE; proceeded with targeted tests passing |
| 2026-06-03 23:33 | Review R007 | plan Step 3: REVISE |
| 2026-06-03 23:35 | Review R008 | plan Step 3: APPROVE |
| 2026-06-03 23:42 | Review R009 | code Step 3: APPROVE |
| 2026-06-03 23:43 | Review R010 | plan Step 4: APPROVE |
| 2026-06-03 23:51 | Review R011 | code Step 4: APPROVE |
| 2026-06-03 23:36 | Review R007 | plan Step 3: REVISE |
| 2026-06-03 23:37 | Review R008 | plan Step 3: APPROVE |
| 2026-06-03 23:41 | Review R009 | code Step 3: APPROVE |
| 2026-06-03 23:42 | Review R010 | plan Step 4: APPROVE |
| 2026-06-03 23:45 | Review R011 | code Step 4: APPROVE |

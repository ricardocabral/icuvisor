# TP-152: Visible catalog/version diagnostic tool — Status

**Current Step:** Step 1: Design diagnostic contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 2
**Iteration:** 1
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
**Status:** 🟨 In Progress

- [x] Stable tool name chosen
- [x] Response shape defined
- [x] Description baseline strategy decided
- [x] No-secret/no-athlete boundary confirmed
- [x] Description catalog fingerprint contract defined
- [x] Mismatch/status semantics clarified
- [x] Same-version fingerprint drift test plan captured

---

### Step 2: Implement tool and tests
**Status:** ⬜ Not Started

- [ ] Core read-only diagnostic registered
- [ ] Output/no-leak tests added
- [ ] Catalog/hash tests updated
- [ ] No-network behavior confirmed
- [ ] Targeted tests passing

---

### Step 3: Update generated docs and stale-catalog guidance
**Status:** ⬜ Not Started

- [ ] Tool docs/catalog regenerated
- [ ] Upgrade/troubleshooting docs updated
- [ ] CHANGELOG updated
- [ ] Privacy wording verified

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Lint passes
- [ ] All failures fixed
- [ ] Build passes

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
| 2026-06-03 23:02 | Review R001 | plan Step 1: REVISE |
| 2026-06-03 23:06 | Review R002 | plan Step 1: APPROVE |

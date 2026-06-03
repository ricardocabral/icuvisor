# TP-152: Visible catalog/version diagnostic tool — Status

**Current Step:** Step 1: Design diagnostic contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 1
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

- [ ] Stable tool name chosen
- [ ] Response shape defined
- [ ] Description baseline strategy decided
- [ ] No-secret/no-athlete boundary confirmed
- [ ] Description catalog fingerprint contract defined
- [ ] Mismatch/status semantics clarified
- [ ] Same-version fingerprint drift test plan captured

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
- Response shape: no-argument read-only response with top-level `server_version`, `catalog_hash`, `toolset`, `delete_mode`, `description_server_version`, `status`, and `action`; `_meta` repeats non-secret diagnostic source fields only.
- Description baseline: generate the tool description at registration with the available server version plus active toolset/delete-mode and a comparable `description_catalog_fingerprint`. The fingerprint is a deterministic SHA-256 over the active catalog records with the diagnostic tool description normalized to a stable sentinel before injecting the fingerprint token. The response returns both the live runtime `catalog_hash` and the same comparable `description_catalog_fingerprint`; assistants compare visible description fields to response fields when clients hide `_meta`.
- Privacy boundary: the tool has no intervals client dependency, no arguments, and returns no API key, athlete ID, filesystem path, username, raw env value, or network-derived data.
| 2026-06-03 23:02 | Review R001 | plan Step 1: REVISE |

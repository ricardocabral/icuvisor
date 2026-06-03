# TP-149: OpenAPI endpoint-diff triage automation — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Clean-room constraint confirmed

---

### Step 1: Design endpoint-diff triage workflow
**Status:** ⬜ Not Started

- [ ] Inspect existing scripts/workflows and decide whether to add a standalone script, scheduled workflow, or documented manual command
- [ ] Ensure normal tests do not hit the network; any live fetch must be opt-in or confined to CI schedule/manual workflow
- [ ] Define output that highlights added/removed OpenAPI paths and creates a human-triage artifact without auto-implementing endpoints
- [ ] Plan-review checkpoint completed before implementation

---

### Step 2: Implement OpenAPI diff tooling
**Status:** ⬜ Not Started

- [ ] Add script or workflow that compares a pinned/baseline intervals.icu OpenAPI spec against latest fetched spec
- [ ] Add fixture-based tests for added path detection, removed path detection, and no-change output
- [ ] Document how maintainers triage new endpoints into Taskplane/backlog tasks
- [ ] Run targeted tests for the script/tooling

---

### Step 99: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Targeted tests passing
- [ ] FULL test suite passing
- [ ] Build passes if code changed
- [ ] All failures fixed

---

### Step 100: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] Must-update docs modified
- [ ] Check-if-affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

<!-- Workers log durable discoveries here. -->

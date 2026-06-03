# TP-149: OpenAPI endpoint-diff triage automation — Status

**Current Step:** Step 1: Design endpoint-diff triage workflow
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 1
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Clean-room constraint confirmed

---

### Step 1: Design endpoint-diff triage workflow
**Status:** 🟨 In Progress

- [x] Inspect existing scripts/workflows and decide whether to add a standalone script, scheduled workflow, or documented manual command
- [x] Ensure normal tests do not hit the network; any live fetch must be opt-in or confined to CI schedule/manual workflow
- [x] Define output that highlights added/removed OpenAPI paths and creates a human-triage artifact without auto-implementing endpoints
- [x] Plan-review checkpoint completed before implementation
- [x] Address R001 plan feedback by using a testable OpenAPI diff package/command layout rather than logic only in a build-ignored root script

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

| 2026-06-03 | Step 1 design | Add a testable `scripts/openapidiff/` Go package/command with offline fixture-friendly `-baseline`/`-latest` inputs and opt-in `-latest-url` fetching, plus a manual/scheduled GitHub workflow that writes a Markdown triage summary. Normal `make test` remains offline because tests use local fixtures only. Output reports added/removed path keys and next triage steps; it must not generate tools or auto-implement endpoints. |

| 2026-06-03 16:11 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:11 | Step 0 started | Preflight |
| 2026-06-03 | R001 plan review | Requested testable layout instead of logic only in a root build-ignored script; revised plan to `scripts/openapidiff/` normal package/command. |
| 2026-06-03 16:14 | Review R001 | plan Step 1: REVISE |

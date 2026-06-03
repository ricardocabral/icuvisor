# TP-139: Coach-mode athlete routing and authorization errors — Status
**Current Step:** Step 1: Audit coach/local athlete routing
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 1
**Size:** M
> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.
---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Audit coach/local athlete routing
**Status:** 🟨 In Progress

- [ ] Inspect `internal/coach`, athlete ID normalization, `list_athletes`, `select_athlete`, and MCP registration/routing behavior.
- [ ] Identify where unauthorized coached-athlete access, local-mode athlete targeting, or tool ACL failures currently become generic upstream errors or ambiguous state.
- [ ] Define expected public error message classes for invalid athlete ID format, unauthorized/not-configured athletes, and selected-athlete tool denial without leaking credentials or raw sensitive identifiers.
- [ ] Document an audit matrix in STATUS.md mapping entrypoint, current behavior, ambiguous failure, expected message, and Step 2 target.
- [ ] Run targeted tests: `go test ./internal/coach ./internal/config ./internal/tools ./internal/mcp`.

---

### Step 2: Add routing/error tests and hardening
**Status:** ⬜ Not Started

- [ ] Add tests for normalized `i123`/numeric athlete IDs, unauthorized coached-athlete selection, and local-athlete fallback when coach mode is not active.
- [ ] Implement explicit authorization/routing errors where tests reveal ambiguity.
- [ ] Ensure tool catalog/ACL behavior still hides disallowed tools and does not accept API keys in chat/tool parameters.
- [ ] Run targeted tests: `go test ./internal/coach ./internal/config ./internal/tools ./internal/mcp`.

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures with exact command output
- [ ] Build passes: `make build`

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

---

## Discoveries

| Date | Step | Finding | Impact |
|------|------|---------|--------|

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|

| 2026-06-03 15:43 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 15:43 | Step 0 started | Preflight |
| 2026-06-03 | Plan Review | REVISE | Step 1 plan expanded to include MCP registration/routing layer, local-mode behavior, concrete audit matrix, and distinct public message classes. |
| 2026-06-03 15:45 | Review R001 | plan Step 1: UNKNOWN |

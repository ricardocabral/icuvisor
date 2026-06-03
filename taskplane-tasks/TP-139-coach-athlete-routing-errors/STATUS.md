# TP-139: Coach-mode athlete routing and authorization errors — Status
**Current Step:** Step 1: Audit coach/local athlete routing
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 3
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
**Status:** ✅ Complete

- [x] Inspect `internal/coach`, athlete ID normalization, `list_athletes`, `select_athlete`, and MCP registration/routing behavior.
- [x] Identify where unauthorized coached-athlete access, local-mode athlete targeting, or tool ACL failures currently become generic upstream errors or ambiguous state.
- [x] Define expected public error message classes for invalid athlete ID format, unauthorized/not-configured athletes, and selected-athlete tool denial without leaking credentials or raw sensitive identifiers.
- [x] Document an audit matrix in STATUS.md mapping entrypoint, current behavior, ambiguous failure, expected message, and Step 2 target.
- [x] Run targeted tests: `go test ./internal/coach ./internal/config ./internal/tools ./internal/mcp`.

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
| 2026-06-03 | Step 1 | Inspected coach evaluator/filter/selection, config athlete ID normalization, list/select athlete tools, and MCP schema/registrar target resolution. | Audit covers coach roster ACLs, selected-athlete fallback, athlete_id schema injection/stripping, and public tool-error conversion before Step 2 hardening. |
| 2026-06-03 | Step 1 | `coach.ToolFilter.ResolveTarget` collapses invalid format, not-in-roster, missing default/selection, and ACL denial into `invalid target athlete`/`coach ACL denied...`; MCP then maps all target failures to `invalid athlete_id; use a configured target athlete`. | Client-visible errors are enumeration-safe but ambiguous; Step 2 should keep safe public text while distinguishing invalid format, unauthorized target, and per-athlete tool denial. |
| 2026-06-03 | Step 1 | Local mode does not inject `athlete_id` into schemas and `resolveToolTarget` returns before local `resolveAthleteID`; extra `athlete_id` reaches strict tool decoders as an unknown field rather than a dedicated routing error. | Step 2 should verify local-mode fallback rejects model-supplied athlete targeting before upstream calls with an actionable message and no credential/API-key parameter path. |
| 2026-06-03 | Step 1 | Expected public routing messages: invalid format -> `invalid athlete_id; use format i12345 or 12345`; unauthorized/not-configured -> `athlete_id is not authorized for this icuvisor coach roster`; ACL denial -> `tool is not allowed for the selected athlete`; local mode supplied target -> `athlete_id is only supported when coach mode is enabled`. | Messages are short, actionable, and avoid echoing credentials or raw target IDs; Step 2 tests can assert exact text or stable substrings. |
| 2026-06-03 | Step 1 | Targeted audit tests passed: `go test ./internal/coach ./internal/config ./internal/tools ./internal/mcp`. | Baseline is green before Step 2 behavior changes. |

## Audit Matrix

| Entrypoint | Current behavior | Ambiguous/generic failure | Expected public message | Step 2 target |
|------------|------------------|---------------------------|-------------------------|---------------|
| `config.NormalizeAthleteID` / coach config load | Trims, lowercases leading `I`, preserves numeric IDs; config load validates roster/default. | Public tool calls can still collapse bad format with unauthorized roster. | `invalid athlete_id; use format i12345 or 12345` | Add normalized `i123`/numeric test coverage where missing. |
| `select_athlete` | Normalizes target and requires roster membership; returns a single `invalid athlete_id; use a configured target athlete` for bad JSON, bad format, and out-of-roster. | Not possible to tell malformed input from unauthorized coached athlete; no raw ID leak today. | bad JSON/format: `invalid athlete_id; use format i12345 or 12345`; unauthorized: `athlete_id is not authorized for this icuvisor coach roster` | Add tests and split errors. |
| MCP athlete-scoped calls in coach mode | Injects schema `athlete_id`, strips it before tool handlers, chooses supplied -> selected -> default, authorizes via coach ACL, then sets `intervals.WithTargetAthleteID`. | Bad format, not in roster, and selected-athlete ACL denial all surface as `invalid athlete_id; use a configured target athlete`. | format/unauthorized/ACL-denied message classes above; ACL denial: `tool is not allowed for the selected athlete` | Add protocol tests asserting no upstream call and distinct messages. |
| MCP tools/list and advanced capabilities | Middleware filters by selected athlete; control tools remain visible; catalog registration hides tools unavailable for all roster athletes/capability/toolset. | Need regression protection that ACL-hidden tools remain hidden after routing changes. | Hidden tools should remain absent rather than callable; denied registered tools return ACL denial if called after stale catalog. | Extend/keep MCP catalog/ACL tests. |
| Local mode athlete-scoped calls | No `athlete_id` schema injection; supplied extra `athlete_id` is handled by individual strict decoders if present, not central routing. | Local fallback to configured athlete is implicit; model-supplied athlete targeting has inconsistent invalid-argument text across tools. | `athlete_id is only supported when coach mode is enabled` when a caller supplies it. | Add central rejection before upstream and verify fallback routes to configured athlete. |
| API keys in tool/chat parameters | Config/keychain owns API key; no tool schemas include API key fields. | Unknown-field errors are generic; must avoid adding any API-key override while hardening routing. | No API-key parameter accepted; unknown/invalid arguments remain public invalid-input errors without echoing secrets. | Add schema/strict rejection regression if affected. |

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|

| 2026-06-03 15:43 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 15:43 | Step 0 started | Preflight |
| 2026-06-03 | Plan Review | REVISE | Step 1 plan expanded to include MCP registration/routing layer, local-mode behavior, concrete audit matrix, and distinct public message classes. |
| 2026-06-03 | Code Review | APPROVE | Step 1 audit deliverable approved. |
| 2026-06-03 15:45 | Review R001 | plan Step 1: UNKNOWN |
| 2026-06-03 15:46 | Review R002 | plan Step 1: APPROVE |
| 2026-06-03 15:50 | Review R003 | code Step 1: APPROVE |

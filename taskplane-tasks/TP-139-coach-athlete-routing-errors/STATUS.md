# TP-139: Coach-mode athlete routing and authorization errors — Status
**Current Step:** Step 3: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 8
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
**Status:** ✅ Complete

- [x] Add unit/protocol tests for normalized `i123`/numeric athlete IDs, unauthorized coached-athlete selection, coach target resolution, and local-athlete fallback/rejection when coach mode is not active.
- [x] Implement explicit routing error types/messages for invalid athlete ID format, unauthorized/not-configured roster athletes, per-athlete tool denial, and local-mode athlete targeting.
- [x] Ensure tool catalog/ACL behavior still hides disallowed tools and schemas do not expose or accept API keys in chat/tool parameters.
- [x] Run targeted tests: `go test ./internal/coach ./internal/config ./internal/tools ./internal/mcp`.
- [x] Revise `select_athlete` runtime decoding to reject credential-like extra parameters without changing selected athlete.

---

### Step 3: Testing & Verification
**Status:** ✅ Complete

- [x] Run FULL test suite: `make test`
- [x] Run lint: `make lint`
- [x] Fix all failures or document pre-existing unrelated failures with exact command output
- [x] Build passes: `make build`

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
| 2026-06-03 | Step 2 | Added/updated coach filter, select_athlete, and MCP protocol tests for numeric/prefixed normalization, unauthorized targets, ACL denial, and local-mode fallback/rejection. | Regression coverage now pins explicit routing behavior before delivery. |
| 2026-06-03 | Step 2 | Implemented coach routing sentinel errors plus public messages for invalid athlete_id format, unauthorized roster target, selected-athlete tool denial, and local-mode athlete_id override. | Client-facing failures are actionable without echoing athlete IDs or credentials. |
| 2026-06-03 | Step 2 | Added registry coverage that tool input schemas do not expose credential/API-key parameters; existing MCP ACL visibility tests continue to assert hidden tools stay absent. | Protects against model-controlled credentials while preserving coach ACL catalog filtering. |
| 2026-06-03 | Step 2 | Targeted hardening tests passed: `go test ./internal/coach ./internal/config ./internal/tools ./internal/mcp`. | Coach routing changes are green across touched packages. |
| 2026-06-03 | Step 2 | Revised `select_athlete` to use strict argument decoding and added a regression that `api_key` is rejected without changing selected athlete state. | Addresses R005 code review; targeted tests still pass. |
| 2026-06-03 | Step 3 | Full test suite passed: `make test`. | Quality gate test phase is green. |
| 2026-06-03 | Step 3 | Lint passed after changing an error wrap in `coach.ToolFilter.ResolveTarget` to use `%w`: `make lint`. | No lint issues remain. |
| 2026-06-03 | Step 3 | Fixed the only observed quality-gate failure (`errorlint` on non-wrapping `%v`); no pre-existing unrelated failures were observed. | All verification failures addressed in-scope. |
| 2026-06-03 | Step 3 | Build passed: `make build`. | Binary builds after routing hardening. |

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
| 2026-06-03 | Code Review | REVISE | Step 2 reviewer found `select_athlete` schema hides credential fields but runtime JSON decoding silently accepts extra `api_key`-like fields. |
| 2026-06-03 | Code Review | APPROVE | Step 2 hardening approved after strict `select_athlete` decoding revision. |
| 2026-06-03 | Code Review | APPROVE | Step 3 verification approved. |
| 2026-06-03 15:45 | Review R001 | plan Step 1: UNKNOWN |
| 2026-06-03 15:46 | Review R002 | plan Step 1: APPROVE |
| 2026-06-03 15:50 | Review R003 | code Step 1: APPROVE |
| 2026-06-03 15:52 | Review R004 | plan Step 2: APPROVE |
| 2026-06-03 15:59 | Review R005 | code Step 2: REVISE |
| 2026-06-03 16:02 | Review R006 | code Step 2: APPROVE |
| 2026-06-03 16:03 | Review R007 | plan Step 3: APPROVE |
| 2026-06-03 16:05 | Review R008 | code Step 3: APPROVE |

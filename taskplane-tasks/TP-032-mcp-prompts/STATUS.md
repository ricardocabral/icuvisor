# TP-032-mcp-prompts: TP-032-mcp-prompts — Status

**Current Step:** Step 5: Verify
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 1
**Review Counter:** 5
**Iteration:** 1
**Size:** S/M

---

### Step 1: Prompt registration plumbing

**Status:** ✅ Complete

- [x] Wire `prompts/list` and `prompts/get` into the MCP server via the Go SDK
- [x] One greppable registration per prompt, mirroring the tool/resource registry pattern
- [x] Decide which prompts take arguments (e.g. date range, athlete_id for coach triage) and define minimal typed argument schemas with LLM-readable descriptions

### Step 2: Author the five prompts

**Status:** ✅ Complete

- [x] `training analysis` — guides the LLM through fitness/effort reads for a training-load + trend readout
- [x] `recovery check` — wellness-led; surfaces sleep dual-scale, HRV/staleness, readiness
- [x] `weekly planning` — events + training-plan reads; structures a week against planned vs completed
- [x] `race-week taper` — events + fitness reads; taper-week framing
- [x] `coach roster triage` — per-athlete scan; takes `athlete_id` selection, respects coach-mode conventions (the `athlete_id` argument is a selector, not a credential)
- [x] Each prompt cites resources (`icuvisor://...`) and names the tools it expects, but does not inline long-form schema content

### Step 3: Token discipline

**Status:** ✅ Complete

- [x] Prompt bodies stay terse — they orchestrate, they do not lecture; long-form content lives in Resources (TP-031)
- [x] Golden-file lock each rendered prompt so wording changes are deliberate

### Step 4: Docs

**Status:** ✅ Complete

- [x] README: document the five prompts and which clients surface them
- [x] CHANGELOG `[Unreleased]` entry

### Step 5: Verify

**Status:** 🟨 In Progress

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual: `prompts/list` shows five; `prompts/get` renders each; confirm at least one MCP client surfaces them (note any client that ignores `prompts/list` in `STATUS.md` per §7.4 #13)

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |
| 1 | code | Step 1 | REVISE | `.reviews/R001-code-step1.md` |
| 3 | code | Step 3 | REVISE | `.reviews/R003-code-step3.md` |
| 4 | code | Step 3 | REVISE | `.reviews/R004-code-step3.md` |
| 5 | code | Step 3 | APPROVE | inline |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |
| Go SDK prompts are registered with `Server.AddPrompt`, listed via `ClientSession.ListPrompts` / `Prompts`, and rendered via `ClientSession.GetPrompt`; `PromptArgument` values are string arguments with descriptions, not JSON Schema objects. Canonical docs: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#Server.AddPrompt and https://modelcontextprotocol.io/specification/2025-06-18/server/prompts | Use prompt argument descriptions to state expected string types/formats. | `/Users/jusbrasil/go/pkg/mod/github.com/modelcontextprotocol/go-sdk@v1.4.1/internal/docs/server.src.md` |

---

## Execution Log

| Timestamp  | Action      | Outcome                          |
| ---------- | ----------- | -------------------------------- |
| 2026-05-14 | Task staged | Scaffolded from ROADMAP.md v0.4   |
| 2026-05-14 19:12 | Task started | Runtime V2 lane-runner execution |
| 2026-05-14 19:12 | Step 1 started | Prompt registration plumbing |
| 2026-05-14 19:18 | SDK prompt support reviewed | Using `Server.AddPrompt` plus SDK `prompts/list` / `prompts/get` handlers. |
| 2026-05-14 19:24 | Prompt registry scaffolded | Five greppable prompt registrations call `TrainingAnalysisPrompt`, `RecoveryCheckPrompt`, `WeeklyPlanningPrompt`, `RaceWeekTaperPrompt`, and `CoachRosterTriagePrompt`. |
| 2026-05-14 19:31 | Step 1 implementation complete | Added prompt registry plumbing to `internal/mcp` and wired the default app server to `prompts.NewRegistry()`. |
| 2026-05-14 19:35 | Review R001 addressed | Prompt handler errors now expose only `prompts.UserError` messages; internal errors return a generic prompt failure. |
| 2026-05-14 19:38 | Step 2 completed | Five prompts authored with resource URIs and expected tool names. |
| 2026-05-14 19:43 | Review R003 addressed | Weekly planning now includes the advanced-capabilities discovery fallback; race-week taper validates missing `race_date`. |
| 2026-05-14 19:46 | Review R004 addressed | Removed unrelated TP-016 status diff from the worktree so TP-032 commits stay scoped. |
| 2026-05-14 19:49 | README updated | Documented the five MCP prompts and client support expectations. |
| 2026-05-14 19:50 | CHANGELOG updated | Added an Unreleased entry for MCP Prompts. |

---

## Blockers

_None_

---

## Notes

- Targeted `go test ./internal/prompts ./internal/mcp ./internal/app` hit an existing Streamable HTTP shutdown flake in `internal/mcp`; focused prompt/parity rerun passed.

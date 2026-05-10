# icuvisor v0.1 — Taskplane Context

**Last Updated:** 2026-05-10
**Status:** Active
**Next Task ID:** TP-007

---

## Current State

This taskplane is scoped to **v0.1 — Walking skeleton** from `ROADMAP.md`.
The v0.1 goal is an end-to-end pipe from the `icuvisor` binary, through MCP stdio, to intervals.icu.

The repository already has initial scaffolding: `go.mod`, `cmd/icuvisor/main.go`, empty `internal/{config,intervals,mcp,tools}` directories, Makefile, GoReleaser config, CI, README, CONTRIBUTING, SECURITY, CHANGELOG, PRD, and ROADMAP.
The current binary only supports `icuvisor version` and otherwise exits with `icuvisor: not yet implemented`.

Use `/orch all` for the full v0.1 batch, or `/orch taskplane-tasks/<task>/PROMPT.md` for one task.

---

## Authoritative Product Sources

Read these before changing scope:

- `docs/prd/PRD-icuvisor.md` — what/why and product constraints.
- `ROADMAP.md` — v0.1 scope and gates.
- `CLAUDE.md` — repository rules for AI agents.
- `CONTRIBUTING.md` — coding/process rules.
- `SECURITY.md` — credential and vulnerability rules.

If documents conflict: PRD owns product behavior, ROADMAP owns phasing, CONTRIBUTING owns process.

---

## v0.1 Acceptance Gate

- Go module and internal project layout are usable and documented.
- intervals.icu client supports Basic Auth from server config, retry/backoff for 429/5xx, structured errors, shared `*http.Client` timeout, `User-Agent: icuvisor/<version>`, and always closes response bodies.
- MCP stdio transport is wired via `github.com/modelcontextprotocol/go-sdk`.
- `get_athlete_profile` works end-to-end via stdio against intervals.icu, using a manual MCP JSON config on macOS/Claude Desktop.
- The created MCP server is additionally validated against a real local Codex CLI session at `/Users/jusbrasil/Library/pnpm/codex`, exercising every registered MCP tool with prompts.
- API keys are never accepted as tool parameters, never logged, and not stored in plaintext beyond explicit manual client config/env instructions for v0.1.
- Local end-to-end validation may read the repository's untracked `.env` file for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`; never print, commit, copy into docs, or persist those values elsewhere.
- `make build`, `make test`, and `make lint` pass.

---

## Reference Implementation and License Guardrails

- `hhopke/intervals-icu-mcp` is Python-based and MIT-licensed per the PRD. Workers may consult it as a reference for endpoint coverage, argument/response semantics, and API behavior when public intervals.icu docs are incomplete. Do not add it as a dependency, copy Python code mechanically, or preserve Python-specific architecture. Re-derive the Go implementation from first principles and record any referenced behavior in `STATUS.md`.
- `mvilanova/intervals-mcp-server` is Python-based and GPLv3. Do **not** read, copy, paraphrase, transliterate, or port its source code. It may only be used indirectly via the PRD's already-summarized product insights or non-code public materials such as README/issue behavior descriptions.
- Prefer intervals.icu public API docs and black-box testing for reverse-engineering. Use permissively licensed references only as supporting evidence.
- Check licenses before adding any dependency. No GPL dependencies.
- Prefer stdlib first. Keep packages under `internal/` unless external consumers truly need them.

---

## Key Files

| Category | Path |
| --- | --- |
| Entrypoint | `cmd/icuvisor/main.go` |
| Config | `internal/config/` |
| intervals.icu client | `internal/intervals/` |
| MCP transport/server | `internal/mcp/` |
| Tools | `internal/tools/` |
| Tests | matching `*_test.go` files and `testdata/` fixtures |
| Docs | `README.md`, `docs/`, `CHANGELOG.md` |
| Taskplane | `taskplane-tasks/` |
| Config | `.pi/taskplane-config.json` |

---

## Technical Debt / Future Work

Items discovered while executing v0.1 tasks should be logged here by workers.

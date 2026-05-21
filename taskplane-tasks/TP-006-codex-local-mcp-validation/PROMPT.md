# TP-006 — Validate MCP server against a real local Codex CLI session

## Mission

Validate the created icuvisor MCP server against a real local Codex session using the Codex CLI at `/Users/jusbrasil/Library/pnpm/codex`. Configure Codex to launch the freshly built icuvisor MCP stdio server, then run prompts through Codex that exercise every MCP tool currently registered by icuvisor.

Roadmap support: strengthens the v0.1 walking-skeleton gate by proving binary → MCP stdio → Codex local session → intervals.icu works outside unit tests.

Complexity: Blast radius 1, Pattern novelty 2, Security 2, Reversibility 1 = 6 → Review Level 2. Size: M.

## Dependencies

- **TP-004** — `get_athlete_profile` tool end-to-end wiring

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` — especially §7.2.B MCP transports, §7.2.C tool catalog, §7.2.D response shaping, §7.4 #7 MCP schema caching
- `ROADMAP.md` — v0.1 only
- `SECURITY.md`
- `README.md`
- `taskplane-tasks/CONTEXT.md`
- `taskplane-tasks/TP-004-get-athlete-profile-tool/STATUS.md`
- Codex CLI help from `/Users/jusbrasil/Library/pnpm/codex --help`
- Local `.env` file, if present, for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` values used only for real end-to-end validation; never print or commit secrets

## File Scope

Expected files:

- `taskplane-tasks/TP-006-codex-local-mcp-validation/STATUS.md`
- `docs/clients/codex-local.md` if a reusable Codex setup recipe is discovered
- `README.md` only if adding a short pointer to the Codex guide
- `CHANGELOG.md` only if user-facing docs or code are changed
- Temporary files under `.tmp/`, `/tmp`, or another ignored location for Codex config/log capture

Avoid changing application code unless validation reveals a concrete bug. If code changes are needed, keep them minimal and covered by tests.

## Steps

### Step 1: Discover current server and Codex CLI behavior

- [ ] Build the current binary with `make build`
- [ ] Confirm the binary path to use for MCP launch, preferably an absolute path to `bin/icuvisor`
- [ ] Inspect `/Users/jusbrasil/Library/pnpm/codex --help` and any relevant MCP/config help without modifying persistent user settings
- [ ] Identify Codex's MCP server configuration mechanism and whether a temporary config/profile can be used
- [ ] Write the validation plan in `STATUS.md` before changing any config

### Step 2: Prepare safe credentials and isolated Codex config

- [ ] Read `.env` only to check whether `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` are available; record only availability, never values
- [ ] Prefer a temporary Codex config/profile that points to the icuvisor binary and passes required env vars without printing them
- [ ] If persistent Codex config must be touched, back it up first, document the exact file in `STATUS.md`, and restore it before finishing
- [ ] Ensure API keys are not written to tracked files, terminal transcripts, docs, fixtures, or `STATUS.md`
- [ ] Confirm `.env` remains untracked and unchanged

### Step 3: Launch Codex with icuvisor as an MCP server

- [ ] Configure Codex to launch the built icuvisor binary over stdio as an MCP server
- [ ] Start a fresh Codex session so MCP tool schema is loaded after the latest binary build
- [ ] Confirm Codex can see the icuvisor MCP server and list available tools
- [ ] Record the non-sensitive tool list in `STATUS.md`

### Step 4: Exercise every registered MCP tool through Codex prompts

- [ ] Determine the complete set of registered icuvisor tools for this build from Codex and/or direct MCP `tools/list`
- [ ] For each registered tool, write and run a Codex prompt that should cause Codex to invoke that tool
- [ ] For v0.1, explicitly test `get_athlete_profile` with a prompt like: "Use icuvisor to fetch my intervals.icu athlete profile and summarize only non-sensitive fields."
- [ ] Verify each tool call reaches the server and returns a valid, terse response shape
- [ ] For responses backed by intervals.icu, validate that real data was read without copying raw personal data into `STATUS.md`
- [ ] Record only pass/fail, tool name, high-level response shape, and redacted observations in `STATUS.md`

### Step 5: Cleanup, document, and verify

- [ ] Stop any Codex/icuvisor processes started for validation
- [ ] Restore any modified persistent Codex config from backup
- [ ] Remove temporary files containing secrets; retain only redacted logs if useful and ignored
- [ ] Add `docs/clients/codex-local.md` if a repeatable Codex setup recipe was discovered
- [ ] Run `make test` and `make build` after any code/doc changes
- [ ] Update `CHANGELOG.md` if docs or behavior changed
- [ ] Mark task done only when every registered MCP tool has a Codex validation result or a documented blocker

## Reference Implementation Policy

- You may use `hhopke/intervals-icu-mcp` as a permissively licensed, Python-only reference implementation for MCP tool semantics if validation reveals ambiguity. It must not be added as a dependency. Extract behavior/API contracts and implement/debug idiomatic Go from first principles.
- You may not use GPL/copyleft source code as an implementation reference. Do not read, copy, paraphrase, transliterate, or port its code.

## Acceptance Criteria

- Codex CLI at `/Users/jusbrasil/Library/pnpm/codex` launches or connects to the built icuvisor MCP stdio server.
- A fresh Codex local session lists the icuvisor tools.
- Every registered icuvisor MCP tool has been prompted through Codex and has a pass/fail validation record in `STATUS.md`.
- `get_athlete_profile` is validated against real intervals.icu data when `.env` provides credentials.
- No API keys, raw secrets, or unnecessary personal training data are printed, committed, or documented.
- Any persistent Codex config touched during validation is restored or explicitly documented with user approval.

## Do NOT

- Do not commit Codex config files containing local absolute paths, API keys, or athlete IDs unless they use placeholders only.
- Do not print, log, commit, or copy local `.env` values into docs, fixtures, prompts, or `STATUS.md`.
- Do not make unit tests depend on live intervals.icu or Codex.
- Do not leave background Codex/icuvisor processes running.
- Do not add Codex or Python MCP servers as dependencies.
- Do not read, copy, paraphrase, transliterate, or port GPL/copyleft implementation code.

## Documentation

Must update:

- `STATUS.md`

Check if affected:

- `docs/clients/codex-local.md`
- `README.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-006`, for example: `TP-006 validate codex local mcp session`.

---

## Amendments

_Add amendments below this line only._

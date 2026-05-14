# Plan Review: Step 1 — Transport selection plumbing

**Verdict:** Changes requested

I read `PROMPT.md` and `STATUS.md`. There is not enough of an implementation plan to approve: `STATUS.md` currently restates the step checklist, but does not specify the concrete config/CLI surface, precedence, validation rules, logging location, or test coverage needed for this security-sensitive transport-selection change.

## Findings

### 1. No concrete plan artifact for the Step 1 API surface

Step 1 needs an explicit plan for the public configuration surface before implementation. Please spell out the exact names and precedence for all inputs, for example:

- config JSON fields to add, if any (`transport`, `http_bind_addr`, etc.)
- environment variables to add and recognize in `.env`
- CLI flags to add (`--transport`, `--http-bind`, or alternatives)
- precedence between JSON, `.env`, process env, and CLI flags
- fields added to `config.Config`, `config.Options`, and `app.ServerInfo`

This matters because `internal/app.Run` currently only parses `version` and `--config`, and `config.Load` only receives a path. If CLI flags are required by the task, the plan must describe how CLI overrides reach config validation without making `cmd/icuvisor/main.go` fat.

### 2. Bind-address validation and “explicit opt-in” semantics are underspecified

The task’s main security requirement is that non-loopback binding is never the default and is only possible through an explicit value. The plan should define validation behavior for at least these cases:

- omitted bind address: default must be loopback, not `:port` or `0.0.0.0:<port>`
- valid IPv4 loopback, `localhost`, and IPv6 loopback forms, if supported
- non-loopback values such as `0.0.0.0:<port>` or LAN IPs: accepted only when explicitly configured and must log a clear `WARN`
- malformed values: empty host, empty port, non-numeric port, out-of-range port, URL strings, missing host/port
- whether the bind address is validated even when `transport=stdio`, or only when `transport=http`

Please also choose and document the default HTTP port in the plan. The prompt says `127.0.0.1:<port>`, but the actual port is not specified in `STATUS.md`.

### 3. Invalid transport values must not silently fall back

Existing config patterns like safety mode parsing tolerate unknown values by falling back. That would be wrong for this step: invalid transport values must fail loudly at startup. The plan should explicitly call for a strict parser/typed enum such as `TransportStdio` / `TransportHTTP`, with actionable and redacted errors for unknown values.

### 4. Warning location and redaction expectations need to be planned

The plan should state where the non-loopback warning is emitted. Prefer keeping `config` validation side-effect free and logging from startup/app transport selection, using structured `slog` at warning level. The warning must not include API keys or raw athlete identifiers. If the bind address itself is logged, that is probably acceptable, but the plan should say so explicitly.

### 5. Test coverage for Step 1 should be listed before coding

Please add a focused test plan, including:

- default config remains `stdio`
- HTTP bind default is loopback-only
- config/env/CLI selection of `http`
- invalid transport fails
- invalid bind address fails
- explicit non-loopback bind is accepted and causes a warning
- CLI parsing remains backward-compatible for `version`, `--config path`, and `--config=path`
- `.env` recognizes the new env keys if env keys are part of the design

## Recommendation

Revise `STATUS.md` or add a short plan section for Step 1 covering the items above. Once the exact config/flag contract, validation rules, and tests are defined, the step is small and should be safe to implement without touching tool handlers or starting the Streamable HTTP server yet.

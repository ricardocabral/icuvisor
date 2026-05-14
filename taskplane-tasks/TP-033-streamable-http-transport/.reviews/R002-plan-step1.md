# Plan Review: Step 1 — Transport selection plumbing

**Verdict:** APPROVE

I read `PROMPT.md`, `STATUS.md`, and the previous `R001` plan review. The revised Step 1 plan in `STATUS.md` now specifies the configuration surface, precedence, validation rules, logging placement, and focused tests needed for this security-sensitive plumbing step.

## What is now adequate

- The public config contract is concrete: JSON `transport` / `http_bind`, env keys `ICUVISOR_TRANSPORT` / `ICUVISOR_HTTP_BIND`, and CLI overrides `--transport` / `--http-bind`.
- Precedence is explicit and consistent with the existing loader model: JSON, then `.env` only for absent values, then process env, then CLI overrides.
- The transport model is strict: default `stdio`, accepted values `stdio` and `http`, invalid values fail startup instead of silently falling back.
- The bind default is concrete and loopback-only: `127.0.0.1:8765`.
- Bind validation is appropriately defensive: explicit IP host plus numeric port, rejects wildcard-by-omission, URLs, missing/invalid/out-of-range ports, and malformed addresses.
- The non-loopback posture is clear: accepted only when explicitly configured, with the warning emitted at startup rather than from config validation.
- The logging plan is redaction-aware: structured `slog.Warn`, transport and bind address only, no API key or athlete ID.
- The test plan covers the important Step 1 risks: defaults, JSON/env/CLI selection, `.env` recognition, invalid transport/bind, non-loopback warning, and existing `version` / `--config` compatibility.

## Non-blocking implementation notes

- When adding CLI parsing, prefer allowing `--config`, `--transport`, and `--http-bind` to be combined in a predictable way. If both `--flag value` and `--flag=value` forms are supported for new flags, cover both; if not, document/test the chosen syntax.
- The IP-only bind rule intentionally rejects `localhost`; that is acceptable for this task, but make sure user-facing errors and README wording do not imply hostnames are supported.
- Keep the Step 1 implementation limited to selection/config plumbing. Actual Streamable HTTP listener wiring belongs in Step 2.


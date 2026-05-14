# Plan Review: Step 1 — Transport selection plumbing

## Verdict

Not approved yet. `STATUS.md` only marks Step 1 as in progress and restates the task checklist; it does not contain a concrete implementation plan to review.

## Required plan details before coding

1. Specify the configuration surface and precedence for transport selection and bind address:
   - CLI flags, env vars, `.env`, and JSON config field names.
   - `stdio` remains the default transport.
   - HTTP bind defaults to `127.0.0.1:<port>` and is only active when transport is `http`.
2. Keep `cmd/icuvisor` thin. If flags are parsed in `internal/app`, call that out explicitly rather than expanding `main`.
3. Define validation rules:
   - accepted transports are exactly `stdio` and `http`;
   - bind values must be explicit host/IP plus port and fail loudly on malformed input;
   - non-loopback/wildcard behavior must be explicit. Prefer rejecting unspecified binds such as `0.0.0.0`/`::`; if allowed, document why and ensure they require explicit config.
4. Define the warning behavior: log a clear WARNING only when HTTP mode is active with a non-loopback bind, and do not include secrets or raw athlete IDs.
5. List Step 1 tests to add/update, including default loopback bind, invalid transport, invalid bind, precedence/CLI override, and non-loopback detection.
6. Record the canonical Go SDK Streamable HTTP documentation link in `STATUS.md` as requested by the prompt, even if the actual transport wiring is deferred to Step 2.

## Notes

The scoped changes are appropriate for Step 1 if limited to config parsing, validation, flag plumbing, logging policy, and tests. Do not wire the HTTP listener or fork MCP handler logic in this step; that belongs to Step 2.

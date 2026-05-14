# Plan Review: Step 1 — Transport selection plumbing

## Verdict

Approved, with minor clarifications to carry into implementation.

The updated `STATUS.md` now has a concrete Step 1 plan covering the config surface, defaults, precedence, validation, warning behavior, tests, and the Go SDK Streamable HTTP references requested by the prompt. The scope stays appropriately limited to selection/plumbing and does not propose wiring the HTTP listener or forking MCP handler logic.

## Notes before coding

1. Clarify precedence wording in implementation/tests: `.env` should fill absent values only, so JSON config values remain authoritative over `.env`; process env overrides JSON/`.env`; CLI overrides all.
2. Be explicit in tests about wildcard IPs (`0.0.0.0`, `::`). If they are allowed, they must be treated as explicit non-loopback opt-ins and trigger the WARN path; if not, add them to invalid bind cases.
3. Keep invalid bind validation startup-fatal even when transport is `stdio`, as the plan states “invalid transport/bind values fail loudly at startup.”
4. Ensure the non-loopback warning includes only transport/bind metadata and no config dump or athlete/API-key data.

No blocking changes to the plan are required.

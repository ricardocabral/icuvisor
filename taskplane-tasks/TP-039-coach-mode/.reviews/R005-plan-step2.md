# Plan Review — TP-039 Step 2: Config + feature flag

## Verdict

Approved. I reviewed `PROMPT.md`, `STATUS.md`, and the prior R003/R004 plan reviews. The current `STATUS.md` notes now pin down the decisions that were previously ambiguous enough to cause import cycles, drift, or contradictory ACL behavior.

## What is now sufficiently specified

- `internal/toolcatalog` is the shared catalog contract, with exported canonical tool-name constants consumed by both registry/tool construction and config validation.
- Dependency direction is explicit: `config -> coach -> toolcatalog`, with athlete-ID normalization owned by `internal/config` and no `coach -> config` import.
- Feature-flag behavior is defined: default `off`, invalid values fail, `auto` depends on a non-empty roster, `on` requires a roster, and `.env` support must include the new key.
- Any present coach stanza is validated even when effective coach mode is `off`, while runtime remains single-athlete.
- Roster validation rules are covered: normalized IDs, duplicate/default handling, default filling for a single athlete, and redacted string/log output.
- ACL semantics are no longer contradictory: `allowed_tools` is a positive allow list, empty means deny-all, and `denied_tools` overrides allow patterns. Documentation/examples must not use `denied_tools: ["*"]` as a read-only example.

## Implementation notes

- Keep Step 2 limited to config parsing/normalization/validation plus the feature flag. Do not implement the Step 3 evaluator or registry filtering early except for the shared catalog constants needed to validate config.
- Add table-driven config tests for the exact state machine and validation matrix captured in `STATUS.md`, especially `off + invalid coach stanza`, `.env` parsing, wildcard ACL typo defense, duplicate normalized IDs, and `Config.String()` redaction.
- If the implementation introduces `internal/toolcatalog`, include a drift-prevention test or ensure all existing tool constructors use the exported constants immediately; otherwise the source-of-truth benefit can be lost.

I did not run the test suite because this is a plan review only.

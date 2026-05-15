# Plan Review — TP-039 Step 2: Config + feature flag

## Verdict

Request changes. I could not find a separate Step 2 implementation plan in the task directory, so I reviewed the Step 2 checklist in `PROMPT.md` / `STATUS.md` as the plan. The checklist captures the headline requirements, but it is not yet specific enough for a safe implementation in this repo.

## Findings

### [P1] Define a cycle-free source of truth for tool-name / pattern validation before coding

Step 2 requires config-load validation to fail loudly on unknown entries in `coach.athletes[].allowed_tools` and `denied_tools`. The plan needs to say how this will be done without creating an import cycle.

Today `internal/config` is low-level and is imported by `internal/intervals`, `internal/mcp`, and several `internal/tools` files. It cannot import `internal/tools` to discover the catalog. At the same time, hard-coding a separate list in `internal/config` or `internal/coach` will drift from the actual registry.

Please revise the plan to introduce a cycle-free validation boundary, for example:

- a small catalog metadata package that contains canonical tool names/pattern validation and is imported by both the registry and coach config validation; or
- a two-phase validation where config parses/normalizes coach ACLs and a higher layer validates them against the actual registered catalog before server startup.

Whichever option is chosen, the plan must cover wildcard patterns from the prompt (`"*"`, `"get_*"`) explicitly. “Unknown tool names fail loudly” cannot mean exact-name-only validation, or the documented `get_*` example will be rejected.

### [P1] Specify the feature-flag state machine and validation failures

The plan should define the exact behavior for `ICUVISOR_COACH_MODE=off|on|auto`, not just add the env var.

Please include decisions for these cases:

- default / empty value: `off`;
- invalid value: fail startup with an actionable config error (recommended), or document a deliberate safe fallback;
- `auto`: enabled iff the parsed coach roster is present and non-empty;
- `on` with no coach stanza or an empty roster: fail startup rather than registering an unusable coach mode;
- `off` with a coach stanza present: preserve single-athlete behavior, but still decide whether to validate the stanza for typo defense;
- `.env` support: add `ICUVISOR_COACH_MODE` to `recognizedEnvKey`, because this loader whitelists env-file keys.

This is important because Step 3/4 registration-time gating depends on a single authoritative “effective coach mode” value, and the default-off acceptance criterion depends on not changing today’s behavior.

### [P1] Add full coach config validation rules, not only field parsing

The plan needs to list the validations Step 2 will enforce at load time. At minimum:

- normalize every `coach.athletes[].id` via `config.NormalizeAthleteID` and emit canonical `i12345` form;
- reject duplicate normalized athlete IDs;
- reject missing/empty athlete IDs;
- trim labels and allow them to be optional, but do not log raw athlete IDs as labels;
- normalize `coach.default_athlete_id` and require it to be present in the roster when coach mode can be enabled (`on` or effective `auto`);
- define whether `default_athlete_id` may be omitted when exactly one roster athlete exists;
- define whether empty `allowed_tools` means deny-all or inherit a default;
- define precedence between `allowed_tools` and `denied_tools` before later ACL work.

The prompt’s example includes `{ "allowed_tools": ["get_*"], "denied_tools": ["*"] }` for a read-only athlete. If deny patterns override allow patterns, that example denies every tool. The plan should resolve this now so Step 3 does not encode contradictory ACL semantics.

### [P2] Keep config parsing decoupled from ACL evaluation, but expose normalized data for later steps

Step 2 should avoid implementing the full `coach.Evaluator` early, but it should create types that Step 3 can consume without reparsing raw strings. A good plan would name the public structs/methods (for example `config.CoachMode`, `config.CoachConfig`, `config.CoachAthlete`, `Config.EffectiveCoachMode()`), and state where normalized IDs and validated ACL pattern strings live.

This also keeps `internal/coach` pure logic as requested by the prompt, while letting `internal/app` / `internal/mcp` pass the effective mode and roster into the registry in later steps.

### [P2] Include the test matrix in the plan

Please add table-driven tests for the Step 2 edge cases, not only happy-path parsing:

- default mode is off and existing config files still load unchanged;
- env / `.env` parsing of `ICUVISOR_COACH_MODE` including `on`, `off`, `auto`, invalid, and whitespace/case;
- JSON config accepts the new `coach` stanza while `DisallowUnknownFields` still rejects misspelled fields;
- roster IDs normalize and duplicates are rejected;
- default athlete must be in roster;
- wildcard ACL patterns are accepted only when they match at least one known tool, while typos fail loudly;
- `Config.String()` / logs do not print raw API keys or raw athlete roster IDs.

## Notes

- I did not run tests; this is a plan review only.
- Step 1 remains documented as operator-deferred for the authenticated roster probe, so Step 2 should assume config-backed roster data and avoid introducing any upstream roster dependency.

# Plan Review — TP-039 Step 2: Config + feature flag

## Verdict

Request changes. The R003 notes in `STATUS.md` move the plan in the right direction, but the current written plan is still too implicit for a safe Step 2 implementation. In particular, it names the right concepts (`internal/toolcatalog`, effective coach-mode state, normalized coach config) without pinning down the ownership/validation rules that will prevent drift, import cycles, and contradictory ACL behavior.

## Findings

### [P1] Make `internal/toolcatalog` the actual shared source of truth, not just a validation helper

`STATUS.md` now says ACL validation will use an `internal/toolcatalog` boundary, which addresses the import-cycle directionally, but the plan still does not say how this package stays synchronized with the real MCP registry.

If Step 2 only adds a second list of tool names for config validation while the registry keeps its existing private `const` names in `internal/tools`, the typo-defense requirement can drift silently: a renamed/new tool could be registered but rejected by ACL validation, or accepted by validation but never registered.

Please revise the plan to specify one concrete contract, for example:

- `internal/toolcatalog` exports canonical tool-name constants and pattern validation, and `internal/tools` uses those constants when constructing tools; or
- `internal/toolcatalog` exports the expected catalog and a registry test asserts every registered tool is present and every catalog entry registers; or
- a startup two-phase validation validates ACLs against the actual registered catalog, while config loading only parses normalized strings.

The plan should also say whether coach-only tools (`list_athletes`, `select_athlete`, `icuvisor_list_advanced_capabilities`) are valid ACL entries or are outside per-athlete ACL filtering. This affects both wildcard matching and “unknown tool names fail loudly.”

### [P1] Resolve package ownership around normalized coach config without reintroducing a config/coach cycle

The revised status says Step 2 will use `internal/coach` normalized config types. That is plausible, but the plan must spell out the dependency direction because athlete-ID normalization is currently centralized in `internal/config` (`NormalizeAthleteID`), and `internal/config.Load` is the code that parses JSON.

A safe plan would state something like: `internal/config` parses raw JSON, normalizes IDs with `config.NormalizeAthleteID`, validates env/effective mode, and constructs `coach.Config`/`coach.Athlete` values from already-normalized strings; `internal/coach` must not import `internal/config`. Alternatively, move normalization to a lower package and update callers. Without this decision, implementation can easily create an `internal/config` ↔ `internal/coach` cycle or duplicate normalization logic.

### [P1] Finish the feature-flag state machine for disabled mode with a coach stanza

The status now covers default `off`, invalid fail, `auto`, `on` requiring a roster, and `.env` support. It still does not decide what happens when `ICUVISOR_COACH_MODE=off` and a `coach` stanza is present, especially if that stanza contains invalid athlete IDs or unknown ACL tool names.

This is not a minor edge case: the prompt simultaneously requires default-off behavior to remain unchanged and requires config-load validation to fail loudly on unknown ACL tool names. The implementation needs one explicit rule, such as:

- always parse and validate any present `coach` stanza, even when mode is off; or
- when effective mode is off, ignore the stanza for runtime behavior but still validate only JSON shape; or
- when effective mode is off, skip all coach validation and document that typo defense applies only when `on`/effective `auto`.

Whichever rule is chosen, add tests for `off + valid coach stanza` and `off + invalid coach stanza`, because this determines whether merely adding a dormant coach config can break existing single-athlete startup.

### [P2] Make the default-athlete and empty-ACL semantics explicit

`STATUS.md` says to enforce “default selection” and that `allowed_tools` is a positive allow list while denies override allows, but the plan still leaves two behavior-defining cases ambiguous:

- May `coach.default_athlete_id` be omitted when exactly one roster athlete exists? If yes, Step 2 should normalize/fill it; if no, fail with an actionable error.
- Does an omitted or empty `allowed_tools` mean deny-all, allow-all, or inherit a default? “Positive allow list” suggests deny-all, but tests and error messages should lock that in.

These choices are part of config compatibility and should be set before Step 3’s evaluator consumes the normalized types.

### [P2] Fix or explicitly supersede the prompt’s contradictory ACL example

The revised status correctly notes that deny patterns override allow patterns and that `denied_tools: ["*"]` means deny all. That directly contradicts the prompt’s example labeling `{ "allowed_tools": ["get_*"], "denied_tools": ["*"] }` as “read-only.”

Please make the plan say that Step 2 will update documentation/examples/tests to avoid copying that contradictory example forward. A read-only example under deny-overrides-allow should either omit `denied_tools` or deny only write/delete patterns/names. Otherwise Step 6 documentation and Step 3 ACL tests are likely to encode different semantics.

## Notes

- I did not run tests; this is a plan review only.
- Step 1 remains operator-deferred for the authenticated roster probe, so Step 2 should continue assuming a config-backed roster and no upstream roster dependency.

# Plan Review: Step 4 build + lint + race + live re-validation

Result: revise.

The local verification portion is fine: running `make build`, `make test`, `make test-race`, and `make lint` is exactly what Step 4 requires. The live re-validation plan also correctly adapts the original eight-field acceptance criteria to the Step 1 finding: `feel` is not writable, so the fixed public contract should reject `feel` and accept the remaining subjective write shape.

However, the live-probe safety plan needs to be tightened before proceeding.

## Blocking issue

The Step 4 plan says to live-validate the accepted seven-field subjective bundle, which includes `locked`, and then “restore/clean any Step 4 probe mutation; if blocked by the Step 1 locked row, update Blockers.” Step 1 already established that the public API ignores `locked:false` and null clears, leaving a locked synthetic row that cannot be fully cleaned through the API. Given that known behavior, Step 4 must not create another locked wellness row or defer the cleanup problem until after mutation.

Revise the plan to state explicitly:

- Do not set `locked:true` on any new date/row.
- If live validation must include the `locked` field, use only the already-contaminated Step 1 probe row, snapshot it first, and restore all overwrite-able subjective fields to the pre-Step-4 values afterward. The remaining `locked:true` state should be the existing Step 1 blocker, not a new Step 4 safety gap.
- If using a fresh date for live validation, omit `locked:true` and validate `locked` only via existing unit/request-shape coverage or via the already-locked probe row.
- Abort live mutation if local build/test/race/lint fail.

## Suggested Step 4 shape

1. Run `make build`, `make test`, `make test-race`, and `make lint`; stop on any failure.
2. Source `.env-dev` without printing secrets or raw athlete IDs.
3. Via stdio MCP, call `update_wellness` with `feel` and verify the explicit public rejection with no mutation.
4. For the accepted write path, either:
   - use the existing locked Step 1 probe row for the full seven-field bundle, then re-read and restore overwrite-able fields to the pre-Step-4 snapshot; or
   - use a fresh row for the six non-lock subjective fields and avoid `locked:true` entirely.
5. Re-read through `get_wellness_data` or direct GET and record exactly what remains blocked. The only acceptable unresolved live-account blocker after Step 4 is the pre-existing Step 1 locked row.

With that safety constraint added, the plan is acceptable.

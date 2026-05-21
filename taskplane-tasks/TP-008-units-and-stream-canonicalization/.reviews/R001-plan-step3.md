# Plan Review: TP-008 Step 3 — stream-key canonicalizer

## Verdict: approved

The Step 3 plan is concrete enough to implement. It creates an isolated `internal/streams` package, keeps canonicalization at the response boundary, calls out the required known upstream variants, preserves unknown keys via best-effort snake_case plus `_meta.unknown_stream_keys`, and anticipates future `get_activity_streams` integration without mutating upstream-decoded data.

## What looks good

- The canonical map is planned as a greppable code artifact, which matches the PRD requirement that stream-key canonicalization live in code and be test-covered.
- The planned key set covers the important read-path/dynamics families called out by the PRD and prompt: standard activity streams, running dynamics, left/right balance, core temperature, and W' balance.
- Response-boundary helpers are the right abstraction for this repository state: `get_activity_streams` is not implemented yet, so a reusable helper for stream rows/series is appropriate as long as future tools must opt into it at their response boundary.
- Unknown-key behavior is explicitly non-lossy: keep the data under a best-effort snake_case key and surface the original unknown key(s) in `_meta.unknown_stream_keys`.
- The plan adds fixtures for inconsistent upstream casings, which is important because this task exists specifically to absorb upstream casing drift such as `groundContactTime` vs `ground_contact_time`.

## Implementation guardrails

- Define the helper API and response shape before coding. It should be obvious whether callers pass a single row, a map of stream series, or a wrapper response, and where `_meta.unknown_stream_keys` is attached for each shape.
- Make collision behavior explicit and test it. If two upstream keys canonicalize to the same key in one row/map, do not silently overwrite one value. Prefer a deterministic policy such as: identical values collapse to one canonical field; conflicting values are preserved in a documented collision metadata structure or stable alternate fields.
- Preserve existing `_meta` content when adding `unknown_stream_keys`, following the merge patterns in `internal/response`; do not replace `_meta.server_version`, `_meta.units`, scales, or pagination metadata added elsewhere.
- Keep unknown metadata stable for clients: report original upstream key names, sorted/deduplicated, while the data field uses the best-effort snake_case key.
- The snake_case fallback needs tests for acronyms and numbers, not just lower/camel/snake: examples like `WPrimeBalance`, `DfaAlpha1`, `VO2Max`, `leftRightBalance`, and `secs100m` should not produce surprising keys.
- Do not let the stream package import tool or intervals client code. It should remain a small response-boundary utility that future tools can call.
- Add at least focused package tests when the package is introduced, even if the full `make test` / `make lint` / `make build` gate remains in Step 4.
- Remember the task-level documentation requirement: update `STATUS.md`, and add a `[Unreleased]` changelog entry once the canonicalizer becomes user-visible or before closing TP-008.
